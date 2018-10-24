/*
Copyright (C) 2018 Expedia Group.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package duo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestParseInitAuthenticationResponse(t *testing.T) {

	requestUrl, _ := url.Parse("https://some_url.com?sid=123456")
	devices, err := parseInitAuthenticationResponse(nil, loginResponse{}, requestUrl, []byte(duoAuthenticationResponse))
	require.NoError(t, err)

	require.Equal(t, 1, len(devices))
	require.NotNil(t, devices["phone1"])
	assert.Equal(t, "phone1", devices["phone1"].Name)

	factors := devices["phone1"].Factors
	require.Equal(t, 3, len(factors))
	require.NotNil(t, factors["Duo Push"])
	require.NotNil(t, factors["Phone Call"])
	require.NotNil(t, factors["Passcode"])
	assert.Equal(t, "some_url.com", factors["Duo Push"].duoHost)
	assert.Equal(t, "123456", factors["Duo Push"].sid)
}

var duoAuthenticationResponse = `<!DOCTYPE html>
<head></head>
<body>
  <div class="base-wrapper">
    <div class="base-main">
      <div role="main" class="base-body">
        <form action="/frame/prompt" method="post" id="login-form" class="inline">
          <input type="hidden" name="sid" value="123456">
          <input type="hidden" name="url" value="/frame/prompt">
          <input type="hidden" name="enrollment_message" value="">
          <input type="hidden" name="itype" value="saml">
          <input type="hidden" name="out_of_date" value="False">
          <input type= "hidden" name="days_out_of_date" value="0">
          <input type="hidden" name="preferred_factor" value="">
          <input type="hidden" name="preferred_device" value="">
          <input type="hidden" name="days_to_block" value="None">

          <fieldset class="device-selector hidden">
            <h1 class="cramped-frame-view">Device:</h1>
            <div class="device-select-wrapper">
              <select name="device" aria-label="Device">
                <option value="phone1">Android (+XX XXXX XX5530)</option>
              </select>
            </div>
          </fieldset>

          <fieldset data-device-index="phone1" class="hidden">
            <h2 class="medium-or-larger">Choose an authentication method</h2>

            <!--- device: phone1, factor: Duo Push -->
            <div class="row-label push-label">
              <input type="hidden" name="factor" value="Duo Push">
              <span class="label factor-label">Duo Push<small class="recommended">Recommended</small></span>
              <button type="submit" class="positive auth-button">Send Me a Push</button>
            </div>

            <!-- device: phone1, factor: Phone Call -->
            <div class="row-label phone-label">
              <input type="hidden" name="factor" value="Phone Call">
              <span class="label factor-label"> Call Me</span>
              <button type="submit" class="positive auth-button">Call Me</button>
            </div>

            <!-- device: phone1, factor: Passcode -->
            <div class="passcode-label row-label">
              <input type="hidden" name="factor" value="Passcode">
              <span class="label factor-label"> Passcode</span>
              <input type="text" name="passcode" data-index="phone1" class="hidden passcode-input">
              <button class="positive auth-button" id="passcode" type="submit">Enter a Passcode</button>
              <input name="phone-smsable" type="hidden" value="false">
              <input name="mobile-otpable" type="hidden" value="true">
              <input name="next-passcode" type="hidden" value="None">
            </div>
          </fieldset>

          <input type="hidden" name="has-token" value="false">

        </form>
      </div>
    </div>
  </div>
</body>
</html>`

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
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type SamlLoginForm struct {
	loginResponse loginResponse
	cookie        string
}

func NewSamlLoginForm(loginResponse loginResponse, cookie string) SamlLoginForm {
	return SamlLoginForm{loginResponse: loginResponse, cookie: cookie}
}

func (f SamlLoginForm) Submit(c *http.Client) (*http.Response, error) {

	app, err := f.loginResponse.duoSigRequest.app()
	if err != nil {
		return nil, fmt.Errorf("get saml: %v", err)
	}

	data := url.Values{}
	data.Add("AuthMethod", f.loginResponse.authMethod)
	data.Add("Context", f.loginResponse.context)
	data.Add("sig_response", fmt.Sprintf("%s:%s", f.cookie, app))

	req, err := http.NewRequest(http.MethodPost, f.loginResponse.optionsUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("get saml: %v", err)
	}
	return c.Do(req)
}

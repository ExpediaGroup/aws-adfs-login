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
	"strings"
	"testing"
)

func TestParseLoginForm(t *testing.T) {

	requestUrl, _ := url.Parse("https://test.com")
	loginResponse, err := parseLogin(requestUrl, []byte(duoResponse))
	require.NoError(t, err)

	assert.Equal(t, "api-f4ec.duosecurity.com", loginResponse.duoHost)
	assert.True(t, strings.HasPrefix(string(loginResponse.duoSigRequest), "TX|ZGVj"))
	require.NotNil(t, loginResponse.optionsUrl)
	assert.Equal(t, "https://sso.test.com:443/saml/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices", loginResponse.optionsUrl)
}

func TestParseOptionsFormWithRelativeUrl(t *testing.T) {

	form := `<form id="options" method="post" action="/saml/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices"></form>`
	requestUrl, _ := url.Parse("https://sso.test.com")
	loginResponse, err := parseLogin(requestUrl, []byte(form))
	require.NoError(t, err)

	require.NotNil(t, loginResponse.optionsUrl)
	assert.Equal(t, "https://sso.test.com/saml/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices", loginResponse.optionsUrl)
}

func TestFindTX(t *testing.T) {

	signature := `TX|ZGVjYWZccHJlaXNpbmdlcnxE|f68c38f2d38662f3dad7d916a257b:APP|ZGVjYWZccHJlaXNpbmdlcnxESVBFSV|7be95a3671b67964fe1`
	expected := `TX|ZGVjYWZccHJlaXNpbmdlcnxE|f68c38f2d38662f3dad7d916a257b`

	duoSig := duoSigRequest(signature)
	actual, err := duoSig.tx()
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

var duoResponse = `<html lang="en-US">
<head>
</head>
<body dir="ltr" class="body">
<div id="fullPage">
    <div id="contentWrapper" class="float">
        <div id="content">
            <div id="workArea">
                <div id="authArea" class="groupMargin">
                    <div class='groupMargin'>
                        <input type='hidden' name='duo_host' id='duo_host' value='api-f4ec.duosecurity.com' />
                        <input type='hidden' name='duo_sig_request' id='duo_sig_request' value='TX|ZGVjYWZccHJlaXNpbmdlcnxE|f68c38f2d38662f3dad7d916a257b:APP|ZGVjYWZccHJlaXNpbmdlcnxESVBFSV|7be95a3671b67964fe1' />
                        <form method='POST' id='duo_form'>
                            <input id='context' type='hidden' name='Context' value='encrypted data'/>
                            <input id='authMethod' type='hidden' name='AuthMethod' value='DuoAdfsAdapter' />
                        </form>
                        <script>

                            Duo.init( {'duoHost': 'api-f4ec.duosecurity.com', 'sig_request': 'TX|ZGVjYWZccHJlaXNpbmdlcnxE|f68c38f2d38662f3dad7d916a257b:APP|ZGVjYWZccHJlaXNpbmdlcnxESVBFSV|7be95a3671b67964fe1' } );
                        </script>
                        <iframe id="duo_iframe" frameborder="0"></iframe>
                        <style>
                            #duo_iframe {
                                width: 100%;
                                min-width: 304px;
                                max-width: 620px;
                                height: 330px;
                            }
                        </style>
                    </div>
                    <div id="authOptions">
                    <form id="options"  method="post" action="https://sso.test.com:443/saml/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices">
                        <script type="text/javascript">
                            function SelectOption(option) {
                                var i = document.getElementById('optionSelection');
                                i.value = option;
                                document.forms['options'].submit();
                                return false;
                            }
                        </script>
                        <input id="optionSelection" type="hidden" name="AuthMethod" />
                        <div class='groupMargin'></div>
                    </form>
                </div>
                </div>

            </div>
        </div>
    </div>
</div>
</body>
</html>`

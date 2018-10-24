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

package client

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadLoginFormFromInputWithTwoForms(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(bytes.NewBufferString(htmlWithMultipleForms).Bytes())
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	loginForm, err := loadLoginForm(http.DefaultClient, server.URL, "test-user", "test-password")
	require.NoError(t, err)

	expectedAction := fmt.Sprintf("%s/saml/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices", server.URL)
	assert.Equal(t, expectedAction, loginForm.Action.String())
	assert.Equal(t, "POST", loginForm.Method)
	assert.Equal(t, "test-user", loginForm.Values.Get("UserName"))
	assert.Equal(t, "test-password", loginForm.Values.Get("Password"))
}

var htmlWithMultipleForms = `
<html>
    <head></head>
    <body>
        <form method="post" id="loginForm" action="/saml/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices" >
            <input id="userNameInput" name="UserName" type="email" value="" />     
            <input id="passwordInput" name="Password" type="password" />                                   
            <input type="checkbox" name="Kmsi" id="kmsiInput" value="true" />
            <label for="kmsiInput">Keep me signed in</label>
            <input id="optionForms" type="hidden" name="AuthMethod" value="FormsAuthentication" />
        </form>

        <form id="options"  method="post" action="test">
            <input id="optionSelection" type="hidden" name="AuthMethod" />
        </form>
    </body>
</html> 
`

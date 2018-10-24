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

package saml

import (
	"encoding/base64"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestLoadAWSAccounts(t *testing.T) {

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(samlAssertionFormResponse))
	require.NoError(t, err)

	accounts, err := loadAWSAccounts(doc)
	require.NoError(t, err)

	assert.Equal(t, "eps-lab", accounts["123"])
	assert.Equal(t, "eps-prod", accounts["456"])
	assert.Equal(t, "789", accounts["789"])
}

func TestLoadSamlRoles(t *testing.T) {

	samlAssertionField := base64.StdEncoding.EncodeToString([]byte(samlAssertionFieldDecoded))
	awsRoles, err := loadSamlRoles(samlAssertionField, map[string]string{"123456789": "test account 1", "98765431": "test account 2"})

	require.NoError(t, err)

	accRoles := awsRoles.RolesByAccountId("98765431")
	require.Equal(t, 1, len(accRoles))
	assert.Equal(t, "arn:aws:iam::98765431:role/ADFS-SuperUser", accRoles[0].Arn)
	assert.Equal(t, "ADFS-SuperUser", accRoles[0].Name)
	assert.Equal(t, "98765431", accRoles[0].Account.Id)
}

var samlAssertionFieldDecoded = `<?xml version="1.0" encoding="UTF-8"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="XXXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" Version="2.0" IssueInstant="2018-08-06T09:34:49.660Z" Destination="https://signin.aws.amazon.com/saml" Consent="urn:oasis:names:tc:SAML:2.0:consent:unspecified">
   <Issuer xmlns="urn:oasis:names:tc:SAML:2.0:assertion">http://sso.test.biz/adfs/services/trust</Issuer>
   <samlp:Status>
      <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success" />
   </samlp:Status>
   <Assertion xmlns="urn:oasis:names:tc:SAML:2.0:assertion" ID="XXXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" IssueInstant="2018-08-06T09:34:49.660Z" Version="2.0">
      <Issuer>http://sso.test.biz/adfs/services/trust</Issuer>
      <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
         <ds:SignedInfo>
            <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#" />
            <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256" />
            <ds:Reference URI="#_XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX">
               <ds:Transforms>
                  <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature" />
                  <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#" />
               </ds:Transforms>
               <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256" />
               <ds:DigestValue>XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX=</ds:DigestValue>
            </ds:Reference>
         </ds:SignedInfo>
         <ds:SignatureValue>XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX</ds:SignatureValue>
         <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
            <ds:X509Data>
               <ds:X509Certificate>XXXXXXXXXXXXXXXXXXXXXXXXXXXXX</ds:X509Certificate>
            </ds:X509Data>
         </KeyInfo>
      </ds:Signature>
      <Subject>
         <NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent">SEA\dicktracy</NameID>
         <SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
            <SubjectConfirmationData NotOnOrAfter="2018-08-06T09:39:49.660Z" Recipient="https://signin.aws.amazon.com/saml" />
         </SubjectConfirmation>
      </Subject>
      <Conditions NotBefore="2018-08-06T09:34:49.613Z" NotOnOrAfter="2018-08-06T10:34:49.613Z">
         <AudienceRestriction>
            <Audience>urn:amazon:webservices</Audience>
         </AudienceRestriction>
      </Conditions>
      <AttributeStatement>
         <Attribute Name="https://aws.amazon.com/SAML/Attributes/RoleSessionName">
            <AttributeValue>dicktracy@test.com</AttributeValue>
         </Attribute>
         <Attribute Name="https://aws.amazon.com/SAML/Attributes/Role">
            <AttributeValue>arn:aws:iam::123456789:saml-provider/ADFS,arn:aws:iam::123456789:role/ADFS-User</AttributeValue>
            <AttributeValue>arn:aws:iam::123456789:saml-provider/ADFS,arn:aws:iam::123456789:role/ADFS-Admin</AttributeValue>
            <AttributeValue>arn:aws:iam::98765431:saml-provider/ADFS,arn:aws:iam::98765431:role/ADFS-SuperUser</AttributeValue>
         </Attribute>
         <Attribute Name="https://redshift.amazon.com/SAML/Attributes/DbUser">
            <AttributeValue>dicktracy</AttributeValue>
         </Attribute>
         <Attribute Name="https://redshift.amazon.com/SAML/Attributes/DbGroups">
            <AttributeValue>Domain Users</AttributeValue>
            <AttributeValue>all dead</AttributeValue>
            <AttributeValue>read only</AttributeValue>
         </Attribute>
         <Attribute Name="https://redshift.amazon.com/SAML/Attributes/AutoCreate">
            <AttributeValue>true</AttributeValue>
         </Attribute>
      </AttributeStatement>
      <AuthnStatement AuthnInstant="2018-08-06T09:34:49.424Z" SessionIndex="_XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX">
         <AuthnContext>
            <AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</AuthnContextClassRef>
         </AuthnContext>
      </AuthnStatement>
   </Assertion>
</samlp:Response>`

var samlAssertionFormResponse = `
<!DOCTYPE html>
<html>
<head></head>
<body>
<div id="container">
  <div id="content">
  <form id="saml_form" name="saml_form" action="/saml" method="post">
          <fieldset>
            <div  class="saml-account"> <div onClick="expandCollapse(0);">
              <img id="image0" src="/static/image/down.png" valign="middle"></img>
              <div class="saml-account-name">Account: eps-lab (123)</div>
              </div>
              <hr style="border: 1px solid #ddd;">
              <div id="0" class="saml-account" >  
                <div class="saml-role" onClick="checkRadio(this);">
                    <input type="radio" name="roleIndex" value="arn:aws:iam::123:role/ADFS-Developers" class="saml-radio" id="arn:aws:iam::123:role/ADFS-Developers" />
                    <label for="arn:aws:iam::123:role/ADFS-Developers" class="saml-role-description">ADFS-Developers</label>
                    <span style="clear: both;"></span>
                </div>
                <div class="saml-role" onClick="checkRadio(this);">
                    <input type="radio" name="roleIndex" value="arn:aws:iam::123:role/User" class="saml-radio" id="arn:aws:iam::123:role/User" />
                    <label for="arn:aws:iam::123:role/User" class="saml-role-description">User</label>
                    <span style="clear: both;"></span>
                </div>
                 </div></div><div  class="saml-account"> <div onClick="expandCollapse(1);">
              <img id="image1" src="/static/image/down.png" valign="middle"></img>
              <div class="saml-account-name">Account: eps-prod (456)</div>
              </div>
              <hr style="border: 1px solid #ddd;">
              <div id="1" class="saml-account" >  
                <div class="saml-role" onClick="checkRadio(this);">
                    <input type="radio" name="roleIndex" value="arn:aws:iam::456:role/User" class="saml-radio" id="arn:aws:iam::456:role/User" />
                    <label for="arn:aws:iam::456:role/User" class="saml-role-description">User</label>
                    <span style="clear: both;"></span>
                </div>
                 </div></div><div  class="saml-account"> <div onClick="expandCollapse(2);">
              <img id="image2" src="/static/image/down.png" valign="middle"></img>
              <!-- account with no name, yes this happens -->
              <div class="saml-account-name">Account: 789</div>
              </div>
              <hr style="border: 1px solid #ddd;">
              <div id="11" class="saml-account" >  
                <div class="saml-role" onClick="checkRadio(this);">
                    <input type="radio" name="roleIndex" value="arn:aws:iam::789:role/User" class="saml-radio" id="arn:aws:iam::789:role/User" />
                    <label for="arn:aws:iam::789:role/User" class="saml-role-description">User</label>
                    <span style="clear: both;"></span>
                </div>
          </fieldset>
  </form>
  </div>
</div>
</body>
</html>`

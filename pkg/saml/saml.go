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
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/aws"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/html"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

// http response is expected to include SAML Assertion form
// this will be response from initial company login screen if you are not using MFA
// or response from last step of MFA
type AssertionRequester interface {
	Submit(c *http.Client) (*http.Response, error)
}

func LoadAWSRoles(c *http.Client, requester AssertionRequester) (aws.Roles, error) {

	// submit login form and load saml response
	samlAssertionForm, err := loadSamlAssertionForm(c, requester)
	if err != nil {
		return nil, err
	}

	// submit saml assertion form to load accounts
	resp, err := samlAssertionForm.Submit(c)
	if err != nil {
		return nil, err
	}
	doc, err := html.LoadDocument(resp)
	if err != nil {
		return nil, err
	}
	accounts, err := loadAWSAccounts(doc)
	if err != nil {
		return nil, err
	}

	samlAssertion := samlAssertionForm.Values.Get("SAMLResponse")
	return loadSamlRoles(samlAssertion, accounts)
}

func loadSamlAssertionForm(c *http.Client, requester AssertionRequester) (html.Form, error) {

	loginResponse, err := requester.Submit(c)
	if err != nil {
		return html.Form{}, err
	}

	// load response
	doc, err := html.LoadDocument(loginResponse)
	if err != nil {
		return html.Form{}, err
	}

	// load saml assertion form from response
	samlAssertionForm, err := html.LoadForm(loginResponse.Request.URL, doc.Find("form"))
	if err != nil {
		return html.Form{}, err
	}

	// simple validation
	if v := samlAssertionForm.Values.Get("SAMLResponse"); v == "" {
		return samlAssertionForm, errors.New("response did not contain valid SAML assertion")
	}
	return samlAssertionForm, nil
}

func loadSamlRoles(samlAssertion string, accounts map[string]string) (aws.Roles, error) {

	samlResponseDecoded, err := base64.StdEncoding.DecodeString(samlAssertion)
	if err != nil {
		return nil, fmt.Errorf("cannot decode saml response: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(samlResponseDecoded))
	if err != nil {
		return nil, fmt.Errorf("cannot load saml response: %v", err)
	}

	// select roles from saml response
	var awsRoles aws.Roles
	var errs []string
	doc.Find("Attribute[Name|='https://aws.amazon.com/SAML/Attributes/Role'] AttributeValue").Each(func(i int, s *goquery.Selection) {
		role, err := loadRoleFromSamlRoleAttributeValue(s.Text(), samlAssertion, accounts)
		if err != nil {
			errs = append(errs, err.Error())
			return
		}
		accountName, ok := accounts[role.Account.Id]
		if !ok {
			// account name is not listed for this role
			accountName = role.Account.Id
		}

		role.SamlAssertion = samlAssertion
		role.Account.Name = accountName
		awsRoles = append(awsRoles, role)
	})

	if len(errs) != 0 {
		return nil, fmt.Errorf("cannot load roles: %s", strings.Join(errs, ", "))
	}
	return awsRoles, nil
}

// loads aws role from saml role attribute field 'arn:aws:iam::123456789:saml-provider/ADFS,arn:aws:iam::123456789:role/ADFS-User'
// returned role does not contain saml assertion and account name
func loadRoleFromSamlRoleAttributeValue(attributeValue, samlAssertion string, accounts map[string]string) (aws.Role, error) {

	roleArn, principalArn, err := loadArnFields(attributeValue)
	if err != nil {
		return aws.Role{}, err
	}

	arnFields := strings.Split(roleArn, ":")
	accountId := arnFields[4]
	roleName := strings.TrimPrefix(arnFields[5], "role/")

	accountName, ok := accounts[accountId]
	if !ok {
		// account name is not listed for this role
		accountName = accountId
	}

	return aws.Role{
		Account:       aws.Account{Id: accountId, Name: accountName},
		Arn:           roleArn,
		Name:          roleName,
		PrincipalArn:  principalArn,
		SamlAssertion: samlAssertion,
	}, nil
}

// saml role attribute field contains role arn and principal arn, the order is not fixed, makes sure the order is the same
func loadArnFields(attributeValue string) (roleArn string, principalArn string, err error) {

	arns := strings.Split(attributeValue, ",")
	if len(arns) != 2 {
		return "", "", fmt.Errorf("expected 2 values in saml role attribute, but got %d", len(attributeValue))
	}

	arn1 := strings.Split(arns[0], ":")
	arn2 := strings.Split(arns[1], ":")
	if len(arn1) != 6 || len(arn2) != 6 {
		return "", "", fmt.Errorf("expected 6 fields in role arn fields, but got %d and %d", len(arn1), len(arn2))
	}

	if strings.HasPrefix(arn1[5], "role/") {
		return arns[0], arns[1], nil
	}
	if strings.HasPrefix(arn2[5], "role/") {
		return arns[1], arns[0], nil
	}
	return "", "", errors.New("no role arn found in the supplied field")
}

// submits saml assertion form and loads map of account id and account name from returned html page
func loadAWSAccounts(samlAssertionFormResponse *goquery.Document) (map[string]string, error) {

	// account name does not come inside saml response, so we need to do this insanity
	accountNames := make(map[string]string)
	var errs []string
	samlAssertionFormResponse.Find("fieldset > div.saml-account").Each(func(i int, s *goquery.Selection) {
		// account name is in 'Account: acc_name (123456789)' format
		accNameLabel := s.Find("div.saml-account-name").Text()
		accName := strings.TrimPrefix(accNameLabel, "Account: ")
		accNameParts := strings.Split(accName, "(")
		if len(accNameParts) != 2 {
			// some accounts have only number
			accountNames[accName] = accName
			return
		}
		accountNames[strings.TrimSuffix(accNameParts[1], ")")] = strings.TrimSpace(accNameParts[0])
	})

	if len(errs) != 0 {
		return nil, fmt.Errorf("cannot account names: %s", strings.Join(errs, ", "))
	}
	return accountNames, nil
}

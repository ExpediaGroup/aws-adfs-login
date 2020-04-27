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
	"fmt"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/aws"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/mfa/duo"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/saml"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

func LoadAWSRoles(adfsHost, user, password string) (aws.Roles, error) {
	c := newHttpClient()
	return LoadAWSRolesByClient(adfsHost, user, password, c)
}

func LoadAWSRolesWithTimeout(adfsHost, user, password string, timeout time.Duration) (aws.Roles, error) {
	c := newHttpClientWithTimeout(timeout)
	return LoadAWSRolesByClient(adfsHost, user, password, c)
}

func LoadAWSRolesByClient(adfsHost, user, password string, client *http.Client) (aws.Roles, error) {
	loginForm, err := loadLoginForm(client, getLoginUrl(adfsHost), user, password)
	if err != nil {
		return nil, fmt.Errorf("cannot load login form: %v", err)
	}
	return saml.LoadAWSRoles(client, loginForm)
}

func LoadDuoDevices(adfsHost, user, password string) (duo.Devices, error) {
	c := newHttpClient()
	return LoadDuoDevicesWithClient(adfsHost, user, password, c)
}

func LoadDuoDevicesWithTimeout(adfsHost, user, password string, timeout time.Duration) (duo.Devices, error) {
	c := newHttpClientWithTimeout(timeout)
	return LoadDuoDevicesWithClient(adfsHost, user, password, c)
}

func LoadDuoDevicesWithClient(adfsHost, user, password string, c *http.Client) (duo.Devices, error) {
	loginForm, err := loadLoginForm(c, getLoginUrl(adfsHost), user, password)
	if err != nil {
		return nil, fmt.Errorf("cannot load login form: %v", err)
	}
	return duo.Login(c, loginForm)
}

func getLoginUrl(adfsHost string) string {
	return fmt.Sprintf(
		"%s/adfs/ls/idpinitiatedsignon.aspx?loginToRp=urn:amazon:webservices",
		strings.TrimSuffix(adfsHost, "/"))
}

func newHttpClient() *http.Client {
	// make timeout generous when waiting for mfa duo push notifications
	return newHttpClientWithTimeout(20 * time.Second)
}

func newHttpClientWithTimeout(timeout time.Duration) *http.Client {
	// using the same client with cookie jar to persist session
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar, Timeout: timeout}
}

// +build manual

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	adfsHost = `` // https://<host>
	user     = `` // sea\<user>
	password = ``
)

func TestLoadDuoAWSRoles(t *testing.T) {

	devices, err := LoadDuoDevices(adfsHost, user, password)
	require.NoError(t, err)

	//roles, err := devices["phone1"].Factors["Phone Call"].LoadAWSRoles("")
	roles, err := devices["phone1"].Factors["Duo Push"].LoadAWSRoles("")
	//roles, err := devices["phone1"].Factors["Passcode"].LoadAWSRoles("123123")

	require.NoError(t, err)

	for _, role := range roles {
		t.Log(role.Name)
	}
}

func TestLoadDuoAWSRolesWithTimeout(t *testing.T) {

	devices, err := LoadDuoDevicesWithTimeout(adfsHost, user, password, 1 * time.Minute)
	require.NoError(t, err)

	//roles, err := devices["phone1"].Factors["Phone Call"].LoadAWSRoles("")
	roles, err := devices["phone1"].Factors["Duo Push"].LoadAWSRoles("")
	//roles, err := devices["phone1"].Factors["Passcode"].LoadAWSRoles("123123")

	require.NoError(t, err)

	for _, role := range roles {
		t.Log(role.Name)
	}
}

func TestLoadAWSRoles(t *testing.T) {

	roles, err := LoadAWSRoles(adfsHost, user, password)
	require.NoError(t, err)

	for _, role := range roles {
		t.Log(role.Name)
	}
}

func TestLoadAWSRolesWithTimeout(t *testing.T) {

	roles, err := LoadAWSRolesWithTimeout(adfsHost, user, password, 5 * time.Minute)
	require.NoError(t, err)

	for _, role := range roles {
		t.Log(role.Name)
	}
}

func TestLogin(t *testing.T) {

	roleArn := ""

	roles, err := LoadAWSRoles(adfsHost, user, password)
	require.NoError(t, err)

	role, err := roles.RoleByRoleArn(roleArn)
	require.NoError(t, err)

	creds, err := role.Login()
	require.NoError(t, err)
	assert.NotEmpty(t, creds.SecretAccessKey)
}

func TestLoginWithTimeout(t *testing.T) {

	roleArn := ""

	roles, err := LoadAWSRolesWithTimeout(adfsHost, user, password, 1 * time.Minute)
	require.NoError(t, err)

	role, err := roles.RoleByRoleArn(roleArn)
	require.NoError(t, err)

	creds, err := role.Login()
	require.NoError(t, err)
	assert.NotEmpty(t, creds.SecretAccessKey)
}

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

package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"sort"
	"time"
)

type Roles []Role

// Returns all roles that are under specified account id
func (roles Roles) RolesByAccountId(accountId string) Roles {

	var accountRoles []Role
	for _, role := range roles {
		if role.Account.Id == accountId {
			accountRoles = append(accountRoles, role)
		}
	}

	sort.Slice(accountRoles, func(i, j int) bool {
		return accountRoles[i].Name < accountRoles[j].Name
	})
	return accountRoles
}

// Returns role specified by ARN e.g. 'arn:aws:iam::123456789:role/Admin', or error if role does not exist
func (roles Roles) RoleByRoleArn(roleArn string) (Role, error) {

	for _, role := range roles {
		if role.Arn == roleArn {
			return role, nil
		}
	}
	return Role{}, fmt.Errorf("role with %s arn does not exist", roleArn)
}

func (roles Roles) Accounts() []Account {

	accountsSet := make(map[string]interface{})
	var accounts []Account
	for _, role := range roles {
		if _, ok := accountsSet[role.Account.Id]; !ok {
			accountsSet[role.Account.Id] = nil
			accounts = append(accounts, role.Account)
		}
	}

	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Name < accounts[j].Name
	})
	return accounts
}

type Account struct {
	Id   string
	Name string
}

func (account Account) String() string {
	return fmt.Sprintf("id: %s, name: %s", account.Id, account.Name)
}

type Role struct {
	Account       Account
	Arn           string
	Name          string
	PrincipalArn  string
	SamlAssertion string
}

func (role Role) Login() (Credentials, error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return Credentials{}, fmt.Errorf("load default aws config: %v", err)
	}

	input := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  &role.PrincipalArn,
		RoleArn:       &role.Arn,
		SAMLAssertion: &role.SamlAssertion,
	}

	out, err := sts.New(cfg).AssumeRoleWithSAMLRequest(input).Send()
	if err != nil {
		return Credentials{}, fmt.Errorf("aws assume role %s with saml: %v", role.Arn, err)
	}
	return fromSTSCredentials(out.Credentials), nil
}

func (role Role) String() string {
	return fmt.Sprintf("account: {%s}, arn: %s, name: %s", role.Account, role.Arn, role.Name)
}

type Credentials struct {
	AccessKeyId     string
	Expiration      time.Time
	SecretAccessKey string
	SessionToken    string
}

func fromSTSCredentials(stsCreds *sts.Credentials) Credentials {

	if stsCreds == nil {
		return Credentials{}
	}

	return Credentials{
		AccessKeyId:     *stsCreds.AccessKeyId,
		Expiration:      *stsCreds.Expiration,
		SecretAccessKey: *stsCreds.SecretAccessKey,
		SessionToken:    *stsCreds.SessionToken,
	}
}

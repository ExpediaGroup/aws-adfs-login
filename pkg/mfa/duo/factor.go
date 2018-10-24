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
	"github.com/HotelsDotCom/aws-adfs-login/pkg/aws"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/saml"
	"net/http"
	"time"
)

type Devices map[string]Device

type Device struct {
	Name    string            // phone1, ...
	Factors map[string]Factor // 'Duo Push', 'Phone Call', 'Passcode'
}

type factorFactory struct {
	client        *http.Client
	loginResponse loginResponse
	duoHost       string
	sid           string
}

func newFactorFactory(client *http.Client, loginResponse loginResponse, duoHost, sid string) factorFactory {
	return factorFactory{
		client:        client,
		loginResponse: loginResponse,
		duoHost:       duoHost,
		sid:           sid,
	}
}

func (f factorFactory) newFactor(deviceName, name string) Factor {
	return Factor{
		client:        f.client,
		loginResponse: f.loginResponse,
		duoHost:       f.duoHost,
		sid:           f.sid,
		Device:        deviceName,
		Name:          name,
	}
}

type Factor struct {
	client        *http.Client
	duoHost       string
	loginResponse loginResponse
	sid           string
	Device        string
	Name          string
}

// passcode is required only for 'Passcode' factor
func (f Factor) LoadAWSRoles(passcode string) (aws.Roles, error) {

	frame := NewFrame(f.client, f.duoHost, f.sid)
	if err := frame.SubmitPrompt(f.Device, f.Name, passcode); err != nil {
		return nil, fmt.Errorf("device %s factor %s submit frame prompt: %v", f.Device, f.Name, err)
	}

	// TODO make number of retries and time sleep configurable
	for i := 0; i < 20; i++ {
		allow, err := frame.IsStatusAllowed()
		if err != nil {
			return nil, fmt.Errorf("device %s factor %s status: %v", f.Device, f.Name, err)
		}

		if allow {
			samlLoginForm, err := frame.LoadSamlLogin(f.loginResponse)
			if err != nil {
				return nil, fmt.Errorf("device %s factor %s load saml login: %v", f.Device, f.Name, err)
			}
			return saml.LoadAWSRoles(f.client, samlLoginForm)
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("device %s factor %s status: time out", f.Device, f.Name)
}

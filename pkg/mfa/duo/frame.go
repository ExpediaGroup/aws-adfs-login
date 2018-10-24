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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Frame struct {
	client  *http.Client
	duoHost string
	sid     string
	// tx id is set by 'SubmitPrompt' method
	txid string
	// resultUrl is set by 'IsStatusAllowed' method, when return values is true
	resultUrl string
}

func NewFrame(client *http.Client, duoHost, sid string) *Frame {

	return &Frame{
		client:  client,
		duoHost: duoHost,
		sid:     sid,
	}
}

func (f *Frame) SubmitPrompt(device, name, passcode string) error {

	data := url.Values{}
	data.Add("sid", f.sid)
	data.Add("device", device)
	data.Add("factor", name)
	data.Add("out_of_date", "")
	if passcode != "" {
		data.Add("passcode", passcode)
	}

	fr, err := f.sendRequest("prompt", data)
	if err != nil {
		return fmt.Errorf("submit prompt request: %v", err)
	}

	f.txid = fr.Response["txid"]
	return nil
}

func (f *Frame) IsStatusAllowed() (bool, error) {

	if f.txid == "" {
		return false, fmt.Errorf("no txid set on the frame, looks like frame prompt was not submited")
	}

	data := url.Values{}
	data.Add("sid", f.sid)
	data.Add("txid", f.txid)

	fr, err := f.sendRequest("status", data)
	if err != nil {
		return false, fmt.Errorf("send frame status request: %v", err)
	}

	if fr.Response["status_code"] == "allow" {
		f.resultUrl = strings.TrimPrefix(fr.Response["result_url"], "/frame/")
		return true, nil
	}
	return false, nil
}

func (f *Frame) LoadSamlLogin(loginResponse loginResponse) (SamlLoginForm, error) {

	if f.resultUrl == "" {
		return SamlLoginForm{}, fmt.Errorf("no result ulr set on the frame, looks like frame status is not allowed")
	}

	data := url.Values{}
	data.Add("sid", f.sid)
	data.Add("txid", f.txid)

	fr, err := f.sendRequest(f.resultUrl, data)
	if err != nil {
		return SamlLoginForm{}, fmt.Errorf("send frame status request: %v", err)
	}

	cookie := fr.Response["cookie"]
	return NewSamlLoginForm(loginResponse, cookie), nil
}

// --- helper methods ---

func (f *Frame) sendRequest(action string, data url.Values) (frameResponse, error) {

	requestUrl, err := url.Parse(fmt.Sprintf("https://%s", f.duoHost))
	if err != nil {
		return frameResponse{}, fmt.Errorf("parse duo duoHost %s: %v", requestUrl.String(), err)
	}
	requestUrl.Path = fmt.Sprintf("/frame/%s", action)

	req, err := http.NewRequest(http.MethodPost, requestUrl.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return frameResponse{}, fmt.Errorf("new request: %v", err)
	}

	req.Header = getDefaultHeaders()
	response, err := f.client.Do(req)
	if err != nil {
		return frameResponse{}, fmt.Errorf("response: %v", err)
	}

	if response.StatusCode/100 != 2 {
		return frameResponse{}, fmt.Errorf("response status code %d", response.StatusCode)
	}

	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return frameResponse{}, fmt.Errorf("response body: %v", err)
	}

	fr, err := loadFrameResponse(b)
	if err != nil {
		return frameResponse{}, fmt.Errorf("load frame response: %v", err)
	}
	return fr, nil
}

type frameResponse struct {
	Stat     string            `json:"stat"`    // OK|FAIL
	Message  string            `json:"message"` // only present if stat is 'FAIL'
	Response map[string]string `json:"response"`
}

func loadFrameResponse(httpBody []byte) (frameResponse, error) {

	var response frameResponse
	if err := json.Unmarshal(httpBody, &response); err != nil {
		return response, err
	}

	if response.Stat != "OK" {
		return response, fmt.Errorf("parse factor response: %s %s", response.Stat, response.Message)
	}
	return response, nil
}

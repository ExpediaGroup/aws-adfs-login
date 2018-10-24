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
	"bytes"
	"errors"
	"fmt"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/html"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

// http response is expected to include DUO login form (response from initial company login screen)
type LoginRequester interface {
	Submit(c *http.Client) (*http.Response, error)
}

// DUO entry point, client needs to be configured with cookiejar and requester submit method needs to return
// initial DUO login screen
func Login(c *http.Client, requester LoginRequester) (Devices, error) {

	loginResponse, err := login(c, requester)
	if err != nil {
		return nil, fmt.Errorf("duo: %v", err)
	}
	return initAuthentication(c, loginResponse)
}

func initAuthentication(c *http.Client, loginResponse loginResponse) (Devices, error) {

	authResponse, err := postInitAuthentication(c, loginResponse)
	if err != nil {
		return nil, fmt.Errorf("duo: %v", err)
	}

	authResponseBody, err := html.ReadResponseBody(authResponse)
	if err != nil {
		return nil, fmt.Errorf("duo: read auth response: %v", err)
	}
	return parseInitAuthenticationResponse(c, loginResponse, authResponse.Request.URL, authResponseBody)
}

func postInitAuthentication(c *http.Client, loginResponse loginResponse) (*http.Response, error) {

	parent := strings.Join([]string{
		loginResponse.optionsUrl,
		"java_version=",
		"flash_version=",
		"screen_resolution_width=1280",
		"screen_resolution_height=800",
		"color_depth=24",
	}, "&")

	tx, err := loginResponse.duoSigRequest.tx()
	if err != nil {
		return nil, fmt.Errorf("initiate authentication: %v", err)
	}

	requestUrl, err := url.Parse(fmt.Sprintf("https://%s", loginResponse.duoHost))
	if err != nil {
		return nil, fmt.Errorf("initiate authentication: parse duo duoHost %s: %v", loginResponse.duoHost, err)
	}
	requestUrl.Path = "/frame/web/v1/auth"
	params := url.Values{}
	params.Add("tx", tx)
	params.Add("parent", parent)
	params.Add("v", "2.3")
	requestUrl.RawQuery = params.Encode()

	data := url.Values{}
	data.Add("parent", parent)
	data.Add("java_version", "")
	data.Add("flash_version", "22.0.0.209")
	data.Add("screen_resolution_width", "1280")
	data.Add("screen_resolution_height", "800")
	data.Add("color_depth", "24")

	req, err := http.NewRequest(http.MethodPost, requestUrl.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("initiate authentication: %v", err)
	}

	req.Header.Set("Host", "duo_host")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:52.0) Gecko/20100101 Firefox/52.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.Do(req)
}

func parseInitAuthenticationResponse(c *http.Client, loginResponse loginResponse, requestUrl *url.URL, response []byte) (Devices, error) {

	sid := requestUrl.Query().Get("sid")
	if sid == "" {
		return nil, errors.New("parse initiate authentication: no sid found in request url")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return nil, fmt.Errorf("parseinitiate authentication: %v", err)
	}

	// select all device names from options
	devices := make(Devices)
	doc.Find("select[name='device'] option").Each(func(i int, selection *goquery.Selection) {
		if val, ok := selection.Attr("value"); ok {
			devices[val] = Device{Name: val, Factors: make(map[string]Factor)}
		}
	})

	// populate all devices with factors
	dFactory := newFactorFactory(c, loginResponse, requestUrl.Host, sid)
	for deviceName := range devices {
		selector := fmt.Sprintf("fieldset[data-device-index=%q] input[name='factor']", deviceName)
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if val, ok := selection.Attr("value"); ok && val != "" {
				devices[deviceName].Factors[val] = dFactory.newFactor(deviceName, val)
			}
		})
	}
	return devices, nil
}

func getDefaultHeaders() http.Header {

	headers := http.Header{}
	headers.Set("Accept-Language", "en")
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko")
	headers.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	headers.Set("Accept", "text/plain, */*; q=0.01")
	return headers
}

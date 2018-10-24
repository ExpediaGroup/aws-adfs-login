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
	"regexp"
)

type loginResponse struct {
	duoHost       string
	duoSigRequest duoSigRequest
	optionsUrl    string
	authMethod    string
	context       string
}

type duoSigRequest string

func (sig duoSigRequest) tx() (string, error) {

	re, err := regexp.Compile(`(TX\|[^:]+):APP.+`)
	if err != nil {
		return "", fmt.Errorf("tx: %v", err)
	}

	submatch := re.FindStringSubmatch(string(sig))
	if len(submatch) < 2 {
		return "", errors.New("tx: cannot find tx in duo duoSigRequest")
	}
	return submatch[1], nil
}

func (sig duoSigRequest) app() (string, error) {

	re, err := regexp.Compile(`.*(APP\|[^:]+)`)
	if err != nil {
		return "", fmt.Errorf("tx: %v", err)
	}

	submatch := re.FindStringSubmatch(string(sig))
	if len(submatch) < 2 {
		return "", errors.New("tx: cannot find tx in duo duoSigRequest")
	}
	return submatch[1], nil
}

func login(c *http.Client, requester LoginRequester) (loginResponse, error) {

	response, err := requester.Submit(c)
	if err != nil {
		return loginResponse{}, fmt.Errorf("submit login form: %v", err)
	}

	responseBody, err := html.ReadResponseBody(response)
	if err != nil {
		return loginResponse{}, fmt.Errorf("read login response: %v", err)
	}
	return parseLogin(response.Request.URL, responseBody)
}

func parseLogin(requestUrl *url.URL, response []byte) (loginResponse, error) {

	var formResponse loginResponse

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return formResponse, fmt.Errorf("parse duo login form: %v", err)
	}

	doc.Find("form#duo_form input").Each(func(i int, selection *goquery.Selection) {
		if val, ok := selection.Attr("name"); ok {
			if val == "Context" {
				formResponse.context, _ = selection.Attr("value")
				return
			}
			if val == "AuthMethod" {
				formResponse.authMethod, _ = selection.Attr("value")
				return
			}
		}
	})

	doc.Find("form#duo_form").Parent().Find("input").Each(func(i int, selection *goquery.Selection) {
		if val, ok := selection.Attr("name"); ok {
			if val == "duo_host" {
				formResponse.duoHost, _ = selection.Attr("value")
				return
			}
			if val == "duo_sig_request" {
				r, _ := selection.Attr("value")
				formResponse.duoSigRequest = duoSigRequest(r)
				return
			}
		}
	})

	val, ok := doc.Find("form#options").Attr("action")
	if !ok {
		return formResponse, errors.New("parse duo login form: cannot find action attribute of [form#optons]")
	}

	parsedUrl, err := requestUrl.Parse(val)
	if err != nil {
		return formResponse, fmt.Errorf("parse duo login form: cannot parse action attribute of [form#optons] %s: %v", val, err)
	}
	formResponse.optionsUrl = parsedUrl.String()
	return formResponse, nil
}

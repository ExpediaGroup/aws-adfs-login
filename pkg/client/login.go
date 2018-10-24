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
	"errors"
	"github.com/HotelsDotCom/aws-adfs-login/pkg/html"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func loadLoginForm(c *http.Client, url string, username, password string) (html.Form, error) {

	r, err := c.Get(url)
	if err != nil {
		return html.Form{}, err
	}

	loginFormSelection, err := selectLoginForm(r)
	if err != nil {
		return html.Form{}, err
	}

	form, err := html.LoadForm(r.Request.URL, loginFormSelection)
	if err != nil {
		return html.Form{}, err
	}

	// fill in user and password fields
	for name := range form.Values {

		if strings.Contains(strings.ToLower(name), "user") ||
			strings.Contains(strings.ToLower(name), "email") {
			form.Values.Set(name, username)
		}
		if strings.Contains(strings.ToLower(name), "pass") {
			form.Values.Set(name, password)
		}
	}
	return form, nil
}

func selectLoginForm(r *http.Response) (*goquery.Selection, error) {

	doc, err := html.LoadDocument(r)
	if err != nil {
		return nil, err
	}

	var loginFormSelection *goquery.Selection
	doc.Find("form").Each(func(i int, formDoc *goquery.Selection) {
		if loginFormSelection != nil {
			return
		}
		formDoc.Find("input").Each(func(_ int, inputDoc *goquery.Selection) {
			if loginFormSelection != nil {
				return
			}
			if name, ok := inputDoc.Attr("name"); ok {
				if strings.Contains(strings.ToLower(name), "pass") {
					loginFormSelection = formDoc
				}
			}
		})
	})

	if loginFormSelection == nil {
		return nil, errors.New("cannot find login form in the response")
	}
	return loginFormSelection, nil
}

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

package html

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

type Form struct {
	Action *url.URL
	Method string
	Values url.Values
}

func LoadForm(pageUrl *url.URL, formSelection *goquery.Selection) (Form, error) {

	// initial form with default Values
	form := Form{Method: http.MethodGet, Action: pageUrl, Values: make(url.Values)}

	// update Action and Method if it has been returned in response
	if a, ok := formSelection.Attr("action"); ok && a != "" {
		u, err := form.Action.Parse(a)
		if err != nil {
			return form, fmt.Errorf("cannot parse form Action attribute %s: %v", a, err)
		}
		form.Action = u
	}
	if a, ok := formSelection.Attr("method"); ok && a != "" {
		form.Method = strings.ToUpper(a)
	}

	// load form input fields
	formSelection.Find("input").Each(func(i int, selection *goquery.Selection) {
		if name, ok := selection.Attr("name"); ok {
			value, _ := selection.Attr("value")
			form.Values.Add(name, value)
		}
	})
	return form, nil
}

func (f Form) Submit(c *http.Client) (*http.Response, error) {

	req, err := http.NewRequest(f.Method, f.Action.String(), strings.NewReader(f.Values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

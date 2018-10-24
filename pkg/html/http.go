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
	"compress/gzip"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
)

func ReadResponseBody(response *http.Response) ([]byte, error) {

	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("status code %d", response.StatusCode)
	}
	defer response.Body.Close()

	// cannot rely on 'response.Uncompressed', anything that is not 'gzip' will be just read
	contentEncoding := response.Header.Get("Content-Encoding")
	if contentEncoding != "gzip" {
		return ioutil.ReadAll(response.Body)
	}

	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func LoadDocument(r *http.Response) (*goquery.Document, error) {

	defer r.Body.Close()
	if r.StatusCode/100 != 2 {
		return nil, fmt.Errorf("status code %d", r.StatusCode)
	}
	return goquery.NewDocumentFromReader(r.Body)
}

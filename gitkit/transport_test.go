// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitkit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"code.google.com/p/goauth2/oauth"
)

type roundTripper struct {
	statusCode int
	respBody   string
}

func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		Status:        fmt.Sprintf("%d reason phrase", r.statusCode),
		StatusCode:    r.statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewReader([]byte(r.respBody))),
		ContentLength: int64(len(r.respBody)),
		Request:       req,
	}
	return resp, nil
}

func TestServiceAccountTransport(t *testing.T) {
	st := &ServiceAccountTransport{
		Auth: &PEMKeyAuthenticator{
			token: &oauth.Token{AccessToken: "access_token", Expiry: time.Now().Add(1 * time.Hour)},
		},
		Transport: roundTripper{},
	}
	req, err := http.NewRequest("POST", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Test-Header", "test_header")

	resp, err := st.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	headerTests := []struct {
		key, origValue, newValue string
	}{
		{"X-Test-Header", "test_header", "test_header"},
		{"Authorization", "", "Bearer access_token"},
		{"User-Agent", "", "gitkit-go-client/0.1.1"},
		{"Content-type", "", "application/json"},
	}
	newReq := resp.Request
	for i, ht := range headerTests {
		if h := req.Header.Get(ht.key); h != ht.origValue {
			t.Errorf("%d. req.Header.Get(%q) = %q; want %q", i, ht.key, h, ht.origValue)
		}

		if h := newReq.Header.Get(ht.key); h != ht.newValue {
			t.Errorf("%d. newReq.Header.Get(%q) = %q; want %q", i, ht.key, h, ht.newValue)
		}
	}
}

func TestAPIKeyTransport(t *testing.T) {
	at := &APIKeyTransport{
		APIKey:    "API_KEY",
		Transport: roundTripper{},
	}
	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Test-Header", "test_header")

	resp, err := at.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	newURL := "http://localhost?key=API_KEY"
	if resp.Request.URL.String() != newURL {
		t.Errorf("newReq.URL.String() = %q; want %q", resp.Request.URL.String(), newURL)
	}

	headerTests := []struct {
		key, origValue, newValue string
	}{
		{"X-Test-Header", "test_header", "test_header"},
		{"User-Agent", "", "gitkit-go-client/0.1.1"},
	}
	newReq := resp.Request
	for i, ht := range headerTests {
		if h := req.Header.Get(ht.key); h != ht.origValue {
			t.Errorf("%d. req.Header.Get(%q) = %q; want %q", i, ht.key, h, ht.origValue)
		}

		if h := newReq.Header.Get(ht.key); h != ht.newValue {
			t.Errorf("%d. newReq.Header.Get(%q) = %q; want %q", i, ht.key, h, ht.newValue)
		}
	}
}

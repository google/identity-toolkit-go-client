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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestBuildOOBCodeURL(t *testing.T) {
	oobURLTests := []struct {
		widgetURL string
		r         *http.Request
		action    string
		oobCode   string
		url       *url.URL // Result OOB code URL
	}{
		// No widgte URL.
		{
			"",
			&http.Request{Host: "localhost", URL: &url.URL{Path: "/oobAction"}},
			"resetPassword",
			"OOBCode",
			nil,
		},
		// Relatvie widget URL.
		{
			"/widget",
			&http.Request{Host: "localhost", URL: &url.URL{Path: "/oobAction"}},
			"changeEmail",
			"OOBCode",
			&url.URL{
				Scheme:   "http",
				Host:     "localhost",
				Path:     "/widget",
				RawQuery: "mode=changeEmail&oobCode=OOBCode",
			},
		},
		// Absolute widget URL.
		{
			"http://localhost/widget",
			&http.Request{Host: "localhost", URL: &url.URL{Path: "/oobAction"}},
			"resetPassword",
			"OOBCode",
			&url.URL{
				Scheme:   "http",
				Host:     "localhost",
				Path:     "/widget",
				RawQuery: "mode=resetPassword&oobCode=OOBCode",
			},
		},
	}
	for i, ot := range oobURLTests {
		var u *url.URL
		if ot.widgetURL != "" {
			u, _ = url.Parse(ot.widgetURL)
		}
		c := &Client{widgetURL: u, config: &Config{WidgetModeParamName: "mode"}}
		url := c.buildOOBCodeURL(ot.r, ot.action, ot.oobCode)
		if !((url == nil && ot.url == nil) || (url != nil && ot.url != nil && *url == *ot.url)) {
			t.Errorf("%d. Client.buildOOBCodeURL() = %s; want %s", i, url, ot.url)
		}
	}
}

func TestSuccessResponse(t *testing.T) {
	r := SuccessResponse()
	s := struct {
		Success bool `json:"success"`
	}{}
	err := json.Unmarshal([]byte(r), &s)
	if err != nil {
		t.Fatalf("SuccessResponse() returns a non JSON: %q", r)
	}
	if !s.Success {
		t.Fatalf("SuccessResponse() = %q; expect to include \"success\": true", r)
	}
}

func TestErrorResponse(t *testing.T) {
	r := ErrorResponse(fmt.Errorf("an error"))
	e := struct {
		Error string `json:"error"`
	}{}
	err := json.Unmarshal([]byte(r), &e)
	if err != nil {
		t.Fatalf("ErrorResponse() returns a non JSON: %q", r)
	}
	if e.Error != "an error" {
		t.Fatalf("ErrorResponse() = %q; expect to include \"error\": \"an error\"", r)
	}
}

func TestExtracRequestURL(t *testing.T) {
	urlTests := []struct {
		r   *http.Request
		url string
	}{
		{
			&http.Request{
				Host: "localhost",
				URL:  &url.URL{Path: "/path", RawQuery: "a=b"},
			},
			"http://localhost/path",
		},
		{
			&http.Request{
				Host: "www.myhost.com",
				URL:  &url.URL{Path: "/"},
				TLS:  &tls.ConnectionState{},
			},
			"https://www.myhost.com/",
		},
	}
	for i, ut := range urlTests {
		if url := extractRequestURL(ut.r); url.String() != ut.url {
			t.Errorf("%d. extractRequestURL() = %q; want %q", i, url.String(), ut.url)
		}
	}
}

func TestExtractRemoteIP(t *testing.T) {
	ipTests := []struct {
		r  *http.Request
		ip string
	}{
		{&http.Request{RemoteAddr: "127.0.0.1:12345"}, "127.0.0.1"},
		{&http.Request{}, ""},
	}
	for i, it := range ipTests {
		if ip := extractRemoteIP(it.r); ip != it.ip {
			t.Errorf("%d. extractRemoteIP() = %q; want %q", i, ip, it.ip)
		}
	}
}

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

import "net/http"

const (
	clientUserAgent = "gitkit-go-client/0.1.1"
	contentType     = "application/json"
)

// transport is an implementation of http.RoundTripper that add a User-Agent
// HTTP header in the request.
type transport struct {
	http.RoundTripper // Underlying HTTP trans
}

// RoundTrip implements the http.RoundTripper interface.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Copy the request to avoid modifying the original request.
	// This is required by the specification of http.RoundTripper.
	newReq := *req
	newReq.Header = make(http.Header)
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	// Add the User-Agent header.
	newReq.Header.Set("User-Agent", clientUserAgent)
	newReq.Header.Set("Content-Type", contentType)
	return t.RoundTripper.RoundTrip(&newReq)
}

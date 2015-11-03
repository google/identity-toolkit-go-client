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

// +build !appengine

package gitkit

import (
	"net/http"

	"golang.org/x/net/context"
)

// defaultTransport returns the default HTTP transport.
func defaultTransport(ctx context.Context) http.RoundTripper {
	return http.DefaultTransport
}

// apiClient returns the APIClient instance in the Client.
func (c *Client) apiClient(ctx context.Context) *APIClient {
	return c.api
}

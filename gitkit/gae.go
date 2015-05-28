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

// +build appengine

package gitkit

import (
	"errors"
	"net/http"
	"sync"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func runInGAEProd() bool {
	return !appengine.IsDevAppServer()
}

var (
	gaeAppToken   *oauth2.Token
	gaeAppTokenMu sync.Mutex
)

// GAEAppAuthenticator uses Google App Engine App Identity API to authenticate.
type GAEAppAuthenticator struct {
	ctx context.Context
}

// SetContext implements Authenticator interface
func (a *GAEAppAuthenticator) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// AccessToken implements Authenticator interface
func (a *GAEAppAuthenticator) AccessToken(http.RoundTripper) (string, error) {
	gaeAppTokenMu.Lock()
	defer gaeAppTokenMu.Unlock()

	if gaeAppToken == nil || !gaeAppToken.Valid() {
		token, expiry, err := appengine.AccessToken(a.ctx, identitytoolkitScope)
		if err != nil {
			return "", err
		}
		gaeAppToken = &oauth2.Token{
			AccessToken: token,
			Expiry:      expiry,
		}
	}
	return gaeAppToken.AccessToken, nil
}

// NewWithContext creates a Client from the global one and associates it with an
// appengine.Context, which is required by most appengine APIs.
func NewWithContext(ctx context.Context, client *Client) (*Client, error) {
	if _, isGAEAuth := client.authenticator.(*GAEAppAuthenticator); isGAEAuth {
		return nil, errors.New("global client shouldn't have GAEAppAuthenticator")
	}
	newClient := *client
	if newClient.authenticator == nil {
		newClient.authenticator = &GAEAppAuthenticator{ctx}
	} else {
		newClient.authenticator.SetContext(ctx)
	}
	newClient.transport = urlfetch.Client(ctx).Transport
	return &newClient, nil
}

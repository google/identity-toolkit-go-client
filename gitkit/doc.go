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

/*
Package gitkit provides convenient utilities for third party website to
integrate Google Identity Toolkit.

See more at https://developers.google.com/identity-toolkit

To use Identity Toolkit Go client in your own server:

	var client *gitkit.Client

	func handleSignIn(w http.ResponseWriter, r *http.Request) {
		// If there is no valid session, check identity tookit ID token.
		ts := client.TokenFromRequest(r)
		token, err := client.ValidateToken(ts)
		if err != nil {
			// Not a valid token. Handle error.
		}
		// Token is validate and it contains the user account information
		// including user ID, email address, etc.
		// Issue your own session cookie to finish the sign in.
	}

	func main() {
		// Provide configuration. gitkit.LoadConfig() can also be used to load
		// the configuration from a JSON file.
		config := &gitkit.Config{
			ClientID: "123.apps.googleusercontent.com",
			WidgetURL: "http://localhost/gitkit",
			ServerAPIKey: "server_api_key",
			ServiceAccount: "123-abc@developer.gserviceaccount.com",
			PEMKeyPath: "/path/to/service_account/private-key.pem",
		}
		var err error
		client, err = gitkit.New(config)
		if err != nil {
			// Handle error.
		}

		// Provide HTTP handler.
		http.HandleFunc("/signIn", handleSignIn)
		// Start the server.
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

The integration is similar with some differences if you want to use the client
in a Google App Engine app.

	var client *gitkit.Client

	func handleSignIn(w http.ResponseWriter, r *http.Request) {
		// If there is no valid session, check identity tookit ID token.
		// gitkit.NewWithContext needs to be called with the appengine.Context
		// such that the new client is associated with it since most App Engine
		// APIs require a context.
		ctx := appengine.NewContext(r)
		c, err := gitkit.NewWithContext(ctx, client)
		if err != nil {
			// Handle error.
		}

		// Validate the token in the same way.
		ts := client.TokenFromRequest(r)
		token, err := client.ValidateToken(ts)
		if err != nil {
			// Not a valid token. Handle error.
		}
		// Token is validate and it contains the user account information
		// including user ID, email address, etc.
		// Issue your own session cookie to finish the sign in.
	}

	func init() {
		// Provide configuration. gitkit.LoadConfig() can also be used to load
		// the configuration from a JSON file.
		config := &gitkit.Config{
			ClientID: "123.apps.googleusercontent.com",
			WidgetURL: "http://localhost/gitkit",
			ServerAPIKey: "server_api_key",
		}
		// Service account and private key are not required in Google App Engine
		// Prod environment. GAE App Identity API is used to identify the app.
		if appengine.IsDevAppServer() {
			config.ServiceAccount = "123-abc@developer.gserviceaccount.com"
			config.PEMKeyPath = "/path/to/service_account/private-key.pem"
		}
		var err error
		client, err = gitkit.New(config)
		if err != nil {
			// Handle error.
		}

		// Provide HTTP handler
		http.HandleFunc("/signIn", handleSignIn)
	}
*/
package gitkit

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

To use Identity Toolkit Go client:

	var client *gitkit.Client

	func handleSignIn(w http.ResponseWriter, r *http.Request) {
		// If there is no valid session, check identity tookit ID token.
		ts := client.TokenFromRequest(r)
		token, err := client.ValidateToken(context.Background(), ts)
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
			ClientID:	"123.apps.googleusercontent.com",
			WidgetURL:	"http://localhost/gitkit",
			CookieName:	"gtoken",
		}
		var err error
		client, err = gitkit.New(context.Background(), config)
		if err != nil {
			// Handle error.
		}

		// Provide HTTP handler.
		http.HandleFunc("/signIn", handleSignIn)
		// Start the server.
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

The integration with Google App Engine is similar except for the context
variable should be created from the request, i.e., appengine.NewContext(r):

	var client *gitkit.Client

	func handleSignIn(w http.ResponseWriter, r *http.Request) {
		// If there is no valid session, check identity tookit ID token.
		ts := client.TokenFromRequest(r)
		token, err := client.ValidateToken(appengine.NewContext(r), ts)
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
			ClientID:	"123.apps.googleusercontent.com",
			WidgetURL:	"http://localhost/gitkit",
			CookieName:	"gtoken",
		}
		// Set the JSON key file path if running dev server in local.
		if appengine.IsDevAppServer() {
			c.GoogleAppCredentialsPath = googleAppCredentialsPath
		}
		var err error
		client, err = gitkit.New(context.Background(), config)
		if err != nil {
			// Handle error.
		}

		// Provide HTTP handler.
		http.HandleFunc("/signIn", handleSignIn)
		// Start the server.
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

The Go client uses Google Application Default Credentials to access
authentication required Identity Toolkit API. The credentials returned are
determined by the environment the code is running in. Conditions are checked in
the following order:

1. The environment variable GOOGLE_APPLICATION_CREDENTIALS is checked. If this
variable is specified it should point to a file that defines the credentials.
The simplest way to get a credential for this purpose is to create a service
account using the Google Developers Console in the section APIs & Auth, in the
sub-section Credentials. Create a service account or choose an existing one and
select Generate new JSON key. Set the environment variable to the path of the
JSON file downloaded.

2. If you have installed the Google Cloud SDK on your machine and have run the
command gcloud auth login, your identity can be used as a proxy to test code
calling APIs from that machine.

3. If you are running in Google App Engine production, the built-in service
account associated with the application will be used.

4. If you are running in Google Compute Engine production, the built-in
service account associated with the virtual machine instance will be used.

5. If none of these conditions is true, an error will occur.

See more about Google Application Default Credentials at
https://developers.google.com/identity/protocols/application-default-credentials

If Application Default Credentials doesn't work for your use case, you can
set GoogleAppCredentialsPath in the config to the JSON key file path.
*/
package gitkit

# Google Identity Toolkit Go client

[![Build Status][travisimg]][travis]

This is the Go client library for [Google Identity Toolkit][gitkit] services.
Documentation at http://godoc.org/github.com/google/identity-toolkit-go-client/gitkit

The `gitkit` package provides convenient utilities for websites to integrate
with the Google Identity Toolkit service.

See more at https://developers.google.com/identity-toolkit

To use Identity Toolkit Go client:
```go
var client *gitkit.Client

func handleSignIn(w http.ResponseWriter, r *http.Request) {
	// If there is no valid session, check identity tookit ID token.
	ts := client.TokenFromRequest(r)
	token, err := client.ValidateToken(ts)
	if err != nil {
		// Not a valid token. Handle error.
	}
	// Token is valid and it contains the user account information
	// including user ID, email address, etc.
	// Issue your own session cookie to finish the sign in.
}

func main() {
	// Provide configuration. gitkit.LoadConfig() can also be used to load
	// the configuration from a JSON file.
	config := &gitkit.Config{
		ClientID:   "123.apps.googleusercontent.com",
		WidgetURL:  "http://localhost/gitkit",
		CookieName: "gtoken",
	}
	var err error
	client, err = gitkit.New(context.TODO(), config)
	if err != nil {
		// Handle error.
	}

	// Provide HTTP handler.
	http.HandleFunc("/signIn", handleSignIn)
	// Start the server.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

The integration with Google App Engine is similar except for the context
variable should be created from the request, i.e., appengine.NewContext(r):
```go
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
```

The client also provides other methods to help manage user accounts, for
example,

To validate the token and also fetch the account information from the
identity toolkit service:
```go
user, err := client.UserByToken(ctx, token)
```
or:
```go
user, err := client.UserByEmail(ctx, email)
```
or:
```go
user, err := client.UserByLocalID(ctx, localID)
```

To update, or delete the account information of a user:
```go
err := client.UpdateUser(ctx, user)
err := client.DeleteUser(ctx, user)
```

The Go client uses [Google Application Default Credentials][gadc] to access
authentication required Identity Toolkit API. The credentials returned are
determined by the environment the code is running in. Conditions are checked in
the following order:

1. The environment variable `GOOGLE_APPLICATION_CREDENTIALS` is checked. If this
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

If Application Default Credentials doesn't work for your use case, you can
set `GoogleAppCredentialsPath` in the config to the JSON key file path.

[travisimg]: https://api.travis-ci.org/google/identity-toolkit-go-client.svg
[travis]: https://travis-ci.org/google/identity-toolkit-go-client
[gitkit]: https://developers.google.com/identity/toolkit/
[gadc]: https://developers.google.com/identity/protocols/application-default-credentials

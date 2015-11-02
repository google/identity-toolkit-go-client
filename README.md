# Google Identity Toolkit Go client

[![Build Status][travisimg]][travis]

This is the Go client library for [Google Identity Toolkit][gitkit] services.
Documentation at http://godoc.org/github.com/google/identity-toolkit-go-client/gitkit

The `gitkit` package provides convenient utilities for websites to integrate
with the Google Identity Toolkit service.

See more at https://developers.google.com/identity-toolkit

To use Identity Toolkit Go client in your own server:
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
		ClientID: "123.apps.googleusercontent.com",
		WidgetURL: "http://localhost/gitkit",
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
```

The integration is similar with some differences if you want to use the client
in a Google App Engine app.
```go
var client *gitkit.Client

func handleSignIn(w http.ResponseWriter, r *http.Request) {
	// If there is no valid session, check identity toolkit ID token.
	// gitkit.NewWithContext needs to be called with the appengine.Context
	// such that the new client is associated with it since most App Engine
	// APIs require a context.
	ctx := appengine.NewContext(r)
	c, err := gitkit.NewWithContext(ctx, client)
	if err != nil {
		// Handle error.
	}

	// Validate the token in the same way.
	ts := c.TokenFromRequest(r)
	token, err := c.ValidateToken(ts)
	if err != nil {
		// Not a valid token. Handle error.
	}
	// Token is valid and it contains the user account information
	// including user ID, email address, etc.
	// Issue your own session cookie to finish the sign in.
}

func init() {
	// Provide configuration. gitkit.LoadConfig() can also be used to load
	// the configuration from a JSON file.
	config := &gitkit.Config{
		ClientID: "123.apps.googleusercontent.com",
		WidgetURL: "http://localhost/gitkit",
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
```

The client also provides other methods to help manage user accounts, for
example,

To validate the token and also fetch the account information from the
identity toolkit service:
```go
user, err := client.UserByToken(token)
```
or:
```go
user, err := client.UserByEmail(email)
```
or:
```go
user, err := client.UserByLocalID(localID)
```

To update, or delete the account information of a user:
```go
err := client.UpdateUser(user)
err := client.DeleteUser(user)
```

[travisimg]: https://api.travis-ci.org/google/identity-toolkit-go-client.svg
[travis]: https://travis-ci.org/google/identity-toolkit-go-client
[gitkit]: https://developers.google.com/identity/toolkit/

This is the Go client library for Google Identity Toolkit services.
Documentation at http://godoc.org/github.com/google/identity-toolkit-go-client/gitkit

The `gitkit` package provides convenient utilities for websites to integrate with Google Identity Toolkit service.

See more at https://developers.google.com/identity-toolkit

To use Identity Toolkit Go client in your own server:
```
var client *gitkit.Client

func handleSignIn(w http.ResponseWriter, r *http.Request) {
	token := client.TokenFromRequest(r)
	user := client.ValidateToken(token)
	if user != nil {
		// Token is validate and user contains the user account information
		// including user ID, email address, etc.
		// Issue your own session cookie to finish the sign in.
	}
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
```

The integration is similar with some differences if you want to use the client
in a Google App Engine app.
```
var client *gitkit.Client

func handleSignIn(w http.ResponseWriter, r *http.Request) {
	// gitkit.NewWithContext needs to be called with the appengine.Context
	// such that the new client is associated with it since most App Engine
	// APIs require a context.
	ctx := appengine.NewContext(r)
	c, err := gitkit.NewWithContext(ctx, client)
	if err != nil {
		// Handle error.
	}

	// Validate the token in the same way.
	token := c.TokenFromRequest(r)
	user := c.ValidateToken(token)
	if user != nil {
		// Token is validate and user contains the user account information
		// including user ID, email address, etc.
		// Issue your own session cookie to finish the sign in.
	}
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
```

The client also provides other methods to help manage user account, for example,

To fetch the account information:
```
user, err := client.UserByToken(token)
```
or:
```
user, err := client.UserByEmail(email)
```
or:
```
user, err := client.UserByLocalID(localID)
```

To update, or delete the account information of a user:
```
err := client.UpdateUser(user)
err := client.DeleteUser(user)
```

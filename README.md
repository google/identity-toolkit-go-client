This is the Go client library for Google Identity Toolkit services.
Documentation at http://godoc.org/github.com/google/identity-toolkit-go-client/gitkit

The `gitkit` package provides convenient utilities for websites to integrate with Google Identity Toolkit service.

To create a new gitkit client:
```
config := gitkit.Config{
	ClientID: "123.apps.googleusercontent.com",
	WidgetURL: "http://localhost/gitkit",
	ServiceAccount: "123-abc@developer.gserviceaccount.com",
	PEMKeyPath: "private-key.pem",
}
client, err := gitkit.New(&config, nil)
```

To fetch the Identity Toolkit ID token from the request and validate it:
```
token := client.TokenFromRequest(req)
// Upon success, a gitkit.User is returned with fields populated from the ID token.
user, err := client.ValidateToken(token)
```

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

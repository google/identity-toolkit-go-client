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

To create a new gitkit client:
	config := gitkit.Config{
		ClientID: "123.apps.googleusercontent.com",
		WidgetURL: "http://localhost/gitkit",
		ServiceAccount: "123-abc@developer.gserviceaccount.com",
		PEMKeyPath: "private-key.pem",
	}
	c, err := gitkit.New(&config, nil)
*/
package gitkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"code.google.com/p/goauth2/oauth/jwt"
)

const (
	identitytoolkitScope = "https://www.googleapis.com/auth/identitytoolkit"
	publicCertsURL       = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/publicKeys"
)

// Config contains the configurations for creating a Client.
type Config struct {
	// ClientID is the Google OAuth2 client ID for the server.
	ClientID string `json:"clientId,omitempty"`
	// WidgetURL is the identitytoolkit javascript widget URL.
	// It is used to generate the reset password or change email URL and
	// could be an absolute URL, an absolute path or a relative path.
	WidgetURL string `json:"widgetUrl,omitempty"`
	// WidgetModeParamName is the parameter name used by the javascript widget.
	// A default value is used if left unspecified. If the parameter name is set
	// to other value in the javascript widget, this field should be set to the
	// same value.
	WidgetModeParamName string `json:"widgetModeParamName,omitempty"`
	// CookieName is the name of the cookie that stores the ID token.
	CookieName string `json:"cookieName,omitempty"`
	// ServerAPIKey is the API key for the server to fetch the identitytoolkit
	// public certificates.
	ServerAPIKey string `json:"serverApiKey,omitempty"`
	// ServiceAccount is the Google OAuth2 service account email address.
	ServiceAccount string `json:"serviceAccountEmail,omitempty"`
	// PEMKeyPath is the path of the PEM enconding private key file for the
	// service account.
	// When obtaining a key from the Google API console it will be  downloaded
	// in a PKCS12 encoding, which can be converted to PEM encoding by openssl:
	//
	//	$ openssl pkcs12 -in <key.p12> -nocerts -passin pass:notasecret -nodes -out <key.pem>
	PEMKeyPath string `json:"serviceAccountPrivateKeyFile,omitempty"`
	// PEMKey is the PEM enconding private key for the service account.
	// Either PEMKeyPath or PEMKey should be provided if a service account is
	// required.
	PEMKey []byte `json:"-"`
}

// LoadConfig loads the configuration from the config file specified by path.
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Client provides convenient utilities for integrating identitytoolkit service
// into a web service.
type Client struct {
	config    *Config
	widgetURL *url.URL
	apiClient *APIClient
	certs     *Certificates
}

const (
	defaultWidgetModeParamName = "mode"
	defaultCookieName          = "gtoken"
)

// New creates a Client from the configuration.
// If the transport is nil, a ServiceAccountTransport is used and the service
// account and PEM key should be specified in the configuration.
func New(config *Config, transport http.RoundTripper) (*Client, error) {
	conf := *config
	if conf.ClientID == "" {
		return nil, errors.New("missing ClientID in config")
	}
	if conf.WidgetURL == "" {
		return nil, errors.New("missing WidgetURL in config")
	}
	widgetURL, err := url.Parse(conf.WidgetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid WidgetURL: %s", conf.WidgetURL)
	}
	if conf.WidgetModeParamName == "" {
		conf.WidgetModeParamName = defaultWidgetModeParamName
	}
	if conf.CookieName == "" {
		conf.CookieName = defaultCookieName
	}
	if transport == nil {
		if conf.ServiceAccount == "" {
			return nil, errors.New("missing ServiceAccount in config")
		}
		if len(conf.PEMKey) == 0 {
			if conf.PEMKeyPath == "" {
				return nil, errors.New("missing PEMKey or PEMKeyPath in config")
			}
			key, err := ioutil.ReadFile(conf.PEMKeyPath)
			if err != nil {
				return nil, err
			}
			conf.PEMKey = key
		}
		transport = &ServiceAccountTransport{
			Assertion: jwt.NewToken(conf.ServiceAccount, identitytoolkitScope, conf.PEMKey),
		}
	}
	api := APIClient{http.Client{Transport: transport}}
	if conf.ServerAPIKey == "" {
		return nil, errors.New("missing ServerAPIKey in config")
	}
	certs, err := LoadCerts(fmt.Sprintf("%s?key=%s", publicCertsURL, conf.ServerAPIKey), transport)
	if err != nil {
		return nil, err
	}
	return &Client{config: &conf, widgetURL: widgetURL, apiClient: &api, certs: certs}, nil
}

// TokenFromRequest extracts the ID token from the HTTP request if present.
func (c *Client) TokenFromRequest(req *http.Request) string {
	cookie, _ := req.Cookie(c.config.CookieName)
	if cookie == nil {
		return ""
	}
	return cookie.Value
}

// ValidateToken validates the ID token and returns a User with fields populated
// from the ID token.
// Beside verifying the token is a valid JWT, it also validates that the token
// is not expired and is issued to the client.
func (c *Client) ValidateToken(token string) (*User, error) {
	t, err := VerifyToken(token, c.certs)
	if err != nil {
		return nil, err
	}
	if t.Expired() {
		return nil, fmt.Errorf("token has expired at: %s", t.ExpireAt)
	}
	if t.Audience != c.config.ClientID {
		return nil, fmt.Errorf("incorrect audience in token: %s", t.Audience)
	}
	u := &User{
		LocalID:       t.LocalID,
		Email:         t.Email,
		EmailVerified: t.EmailVerified,
		ProviderID:    t.ProviderID,
	}
	return u, nil
}

// UserByToken retrieves the account information of the user specified by the ID
// token.
func (c *Client) UserByToken(token string) (*User, error) {
	u, err := c.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	localID := u.LocalID
	providerID := u.ProviderID
	u, err = c.UserByLocalID(localID)
	if err != nil {
		return nil, err
	}
	u.ProviderID = providerID
	return u, nil
}

// UserByEmail retrieves the account information of the user specified by the
// email address.
func (c *Client) UserByEmail(email string) (*User, error) {
	resp, err := c.apiClient.GetAccountInfo(&GetAccountInfoRequest{Emails: []string{email}})
	if err != nil {
		return nil, err
	}
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user %s not found", email)
	}
	return resp.Users[0], nil
}

// UserByLocalID retrieves the account information of the user specified by the
// local ID.
func (c *Client) UserByLocalID(localID string) (*User, error) {
	resp, err := c.apiClient.GetAccountInfo(&GetAccountInfoRequest{LocalIDs: []string{localID}})
	if err != nil {
		return nil, err
	}
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user %s not found", localID)
	}
	return resp.Users[0], nil
}

// UpdateUser updates the account information of the user.
func (c *Client) UpdateUser(user *User) error {
	_, err := c.apiClient.SetAccountInfo(&SetAccountInfoRequest{
		LocalID:       user.LocalID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Password:      user.Password,
		EmailVerified: user.EmailVerified})
	return err
}

// DeleteUser deletes a user specified by the local ID.
func (c *Client) DeleteUser(user *User) error {
	_, err := c.apiClient.DeleteAccount(&DeleteAccountRequest{LocalID: user.LocalID})
	return err
}

// UploadUsers uploads the users to identitytoolkit service.
// algorithm, key, saltSeparator specify the password hash algorithm, signer key
// and separator between password and salt accordingly.
func (c *Client) UploadUsers(users []*User, algorithm string, key, saltSeparator []byte) error {
	resp, err := c.apiClient.UploadAccount(&UploadAccountRequest{users, algorithm, key, saltSeparator})
	if err != nil {
		return err
	}
	if len(resp.Error) != 0 {
		return resp.Error
	}
	return nil
}

// ListUsersN lists the next n users.
// For the first n users, the pageToken should be empty. Upon success, the users
// and pageToken for next n users are returned.
func (c *Client) ListUsersN(n int, pageToken string) ([]*User, string, error) {
	resp, err := c.apiClient.DownloadAccount(&DownloadAccountRequest{n, pageToken})
	if err != nil {
		return nil, "", err
	}
	return resp.Users, resp.NextPageToken, nil
}

const maxResultsPerPage = 50

// A UserList holds a channel that delivers all the users.
type UserList struct {
	C     <-chan *User // The channel on which the users are delivered.
	Error error        // Indicates an error occurs when listing the users.

	client    *Client
	pageToken string
}

func (l *UserList) start() {
	ch := make(chan *User, maxResultsPerPage)
	l.C = ch
	go func() {
		for {
			users, pageToken, err := l.client.ListUsersN(maxResultsPerPage, l.pageToken)
			if err != nil {
				l.Error = err
				close(ch)
				return
			}
			if len(users) == 0 || pageToken == "" {
				close(ch)
			} else {
				l.pageToken = pageToken
				for _, u := range users {
					ch <- u
				}
			}
		}
	}()
}

// Retry resets Error to nil and resumes the downloading.
func (l *UserList) Retry() {
	if l.Error != nil {
		l.Error = nil
		l.start()
	}
}

// ListUsers lists all the users.
//
// For example,
//	l := c.ListUsers()
//	for {
//		for u := range l.C {
//			// Do something
//		}
//		if l.Error != nil {
//			l.Retry()
//		} else {
//			break
//		}
//	}
func (c *Client) ListUsers() *UserList {
	l := &UserList{client: c}
	l.start()
	return l
}

// Parameter names used to extract the OOB code request.
const (
	OOBActionParam           = "action"
	OOBEmailParam            = "email"
	OOBCAPTCHAChallengeParam = "challenge"
	OOBCAPTCHAResponseParam  = "response"
	OOBNewEmailParam         = "newEmail"
	OOBCodeParam             = "oobCode"
)

// Acceptable OOB code request types.
const (
	OOBActionChangeEmail   = "changeEmail"
	OOBActionResetPassword = "resetPassword"
)

// OOBCodeResponse wraps the OOB code response.
type OOBCodeResponse struct {
	// Action identifies the request type.
	Action string
	// The email address of the user.
	Email string
	// The new email address of the user.
	// This field is only populated when Action is OOBActionChangeEmail.
	NewEmail string
	// The OOB confirmation code.
	OOBCode string
	// The URL that contains the OOB code and can be sent to the user for
	// confirming the action, e.g., sending the URL to the email address and
	// the user can click the URL to continue to reset the password.
	OOBCodeURL string
}

// GenerateOOBCode generates an OOB code based on the request.
func (c *Client) GenerateOOBCode(req *http.Request) (*OOBCodeResponse, error) {
	q := req.URL.Query()
	action := q.Get(OOBActionParam)
	var requestType, email, newEmail, captchaChallenge, captchaResponse, token string
	switch action {
	case OOBActionResetPassword:
		requestType = ResetPasswordRequestType
		email = q.Get(OOBEmailParam)
		captchaChallenge = q.Get(OOBCAPTCHAChallengeParam)
		captchaResponse = q.Get(OOBCAPTCHAResponseParam)
	case OOBActionChangeEmail:
		requestType = ChangeEmailRequestType
		email = q.Get(OOBEmailParam)
		newEmail = q.Get(OOBNewEmailParam)
		token = c.TokenFromRequest(req)
	default:
		return nil, fmt.Errorf("unrecognized action: %s", action)
	}

	// Set all possible fields in request and let APIClient do the validation.
	r := &GetOOBCodeRequest{
		RequestType:      requestType,
		Email:            email,
		CAPTCHAChallenge: captchaChallenge,
		CAPTCHAResponse:  captchaResponse,
		NewEmail:         newEmail,
		Token:            token,
		UserIP:           extractRemoteIP(req),
	}
	resp, err := c.apiClient.GetOOBCode(r)
	if err != nil {
		return nil, err
	}

	// Build the OOB code URL.
	url := extractRequestURL(req).ResolveReference(c.widgetURL)
	url.Query().Set(c.config.WidgetModeParamName, action)
	url.Query().Set(OOBCodeParam, resp.OOBCode)

	return &OOBCodeResponse{action, email, newEmail, resp.OOBCode, url.String()}, nil
}

// SuccessResponse generates a JSON response which indicates the request is
// processed successfully.
func SuccessResponse() string {
	return `{"success": true}`
}

// ErrorResponse generates a JSON error response from the given error.
func ErrorResponse(err error) string {
	return fmt.Sprintf(`{"error": "%s"}`, err)
}

func extractRequestURL(req *http.Request) *url.URL {
	var scheme string
	if req.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}
	return &url.URL{Scheme: scheme, Host: req.Host, Path: req.URL.Path}
}

func extractRemoteIP(req *http.Request) string {
	return strings.Split(req.RemoteAddr, ":")[0]
}

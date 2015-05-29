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

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/jwt"
)

const (
	identitytoolkitScope = "https://www.googleapis.com/auth/identitytoolkit"
	publicCertsURL       = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/publicKeys"
	tokenEndpointURL     = "https://www.googleapis.com/oauth2/v3/token"
)

// Client provides convenient utilities for integrating identitytoolkit service
// into a web service.
type Client struct {
	config    *Config
	widgetURL *url.URL
	certs     *Certificates

	authenticator Authenticator
	transport     http.RoundTripper
}

// New creates a Client from the configuration.
func New(config *Config) (*Client, error) {
	conf := *config
	requireServiceAccountInfo := !runInGAEProd()
	if err := conf.normalize(requireServiceAccountInfo); err != nil {
		return nil, err
	}
	certs := &Certificates{URL: publicCertsURL}
	var widgetURL *url.URL
	if conf.WidgetURL != "" {
		var err error
		widgetURL, err = url.Parse(conf.WidgetURL)
		if err != nil {
			return nil, fmt.Errorf("invalid WidgetURL: %s", conf.WidgetURL)
		}
	}
	var authenticator Authenticator
	if conf.ServiceAccount != "" && len(conf.PEMKey) != 0 {
		authenticator = &PEMKeyAuthenticator{
			assertion: &jwt.Config{
				Email:      conf.ServiceAccount,
				PrivateKey: conf.PEMKey,
				Scopes:     []string{identitytoolkitScope},
				TokenURL:   tokenEndpointURL},
		}
	}
	return &Client{
		config:        &conf,
		widgetURL:     widgetURL,
		authenticator: authenticator,
		certs:         certs,
	}, nil
}

func (c *Client) defaultTransport() http.RoundTripper {
	if c.transport == nil {
		return http.DefaultTransport
	}
	return c.transport
}

func (c *Client) apiClient() *APIClient {
	return &APIClient{
		http.Client{
			Transport: &ServiceAccountTransport{
				Auth:      c.authenticator,
				Transport: c.defaultTransport(),
			},
		},
	}
}

// TokenFromRequest extracts the ID token from the HTTP request if present.
func (c *Client) TokenFromRequest(req *http.Request) string {
	cookie, _ := req.Cookie(c.config.CookieName)
	if cookie == nil {
		return ""
	}
	return cookie.Value
}

// ValidateToken validates the ID token and returns a Token.
//
// Beside verifying the token is a valid JWT, it also validates that the token
// is not expired and is issued to the client.
func (c *Client) ValidateToken(token string) (*Token, error) {
	transport := &APIKeyTransport{c.config.ServerAPIKey, c.defaultTransport()}
	if err := c.certs.LoadIfNecessary(transport); err != nil {
		return nil, err
	}
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
	return t, nil
}

// UserByToken retrieves the account information of the user specified by the ID
// token.
func (c *Client) UserByToken(token string) (*User, error) {
	t, err := c.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	localID := t.LocalID
	providerID := t.ProviderID
	u, err := c.UserByLocalID(localID)
	if err != nil {
		return nil, err
	}
	u.ProviderID = providerID
	return u, nil
}

// UserByEmail retrieves the account information of the user specified by the
// email address.
func (c *Client) UserByEmail(email string) (*User, error) {
	resp, err := c.apiClient().GetAccountInfo(&GetAccountInfoRequest{Emails: []string{email}})
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
	resp, err := c.apiClient().GetAccountInfo(&GetAccountInfoRequest{LocalIDs: []string{localID}})
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
	_, err := c.apiClient().SetAccountInfo(&SetAccountInfoRequest{
		LocalID:       user.LocalID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Password:      user.Password,
		EmailVerified: user.EmailVerified})
	return err
}

// DeleteUser deletes a user specified by the local ID.
func (c *Client) DeleteUser(user *User) error {
	_, err := c.apiClient().DeleteAccount(&DeleteAccountRequest{LocalID: user.LocalID})
	return err
}

// UploadUsers uploads the users to identitytoolkit service.
// algorithm, key, saltSeparator specify the password hash algorithm, signer key
// and separator between password and salt accordingly.
func (c *Client) UploadUsers(users []*User, algorithm string, key, saltSeparator []byte) error {
	resp, err := c.apiClient().UploadAccount(&UploadAccountRequest{users, algorithm, key, saltSeparator})
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
	resp, err := c.apiClient().DownloadAccount(&DownloadAccountRequest{n, pageToken})
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
	OOBOldEmailParam         = "oldEmail"
	OOBNewEmailParam         = "newEmail"
	OOBCodeParam             = "oobCode"
)

// Acceptable OOB code request types.
const (
	OOBActionChangeEmail   = "changeEmail"
	OOBActionVerifyEmail   = "verifyEmail"
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
	// It can be nil if WidgetURL is not provided in the configuration.
	OOBCodeURL *url.URL
}

// GenerateOOBCode generates an OOB code based on the request.
func (c *Client) GenerateOOBCode(req *http.Request) (*OOBCodeResponse, error) {
	switch action := req.PostFormValue(OOBActionParam); action {
	case OOBActionResetPassword:
		return c.GenerateResetPasswordOOBCode(
			req,
			req.PostFormValue(OOBEmailParam),
			req.PostFormValue(OOBCAPTCHAChallengeParam),
			req.PostFormValue(OOBCAPTCHAResponseParam))
	case OOBActionChangeEmail:
		return c.GenerateChangeEmailOOBCode(
			req,
			req.PostFormValue(OOBOldEmailParam),
			req.PostFormValue(OOBNewEmailParam),
			c.TokenFromRequest(req))
	case OOBActionVerifyEmail:
		return c.GenerateVerifyEmailOOBCode(req, req.PostFormValue(OOBEmailParam))
	default:
		return nil, fmt.Errorf("unrecognized action: %s", action)
	}
}

// GenerateResetPasswordOOBCode generates an OOB code for resetting password.
//
// If WidgetURL is not provided in the configuration, the OOBCodeURL field in
// the returned OOBCodeResponse is nil.
func (c *Client) GenerateResetPasswordOOBCode(
	req *http.Request, email, captchaChallenge, captchaResponse string) (*OOBCodeResponse, error) {
	r := &GetOOBCodeRequest{
		RequestType:      ResetPasswordRequestType,
		Email:            email,
		CAPTCHAChallenge: captchaChallenge,
		CAPTCHAResponse:  captchaResponse,
		UserIP:           extractRemoteIP(req),
	}
	resp, err := c.apiClient().GetOOBCode(r)
	if err != nil {
		return nil, err
	}
	return &OOBCodeResponse{
		Action:     OOBActionResetPassword,
		Email:      email,
		OOBCode:    resp.OOBCode,
		OOBCodeURL: c.buildOOBCodeURL(req, OOBActionResetPassword, resp.OOBCode),
	}, nil
}

// GenerateChangeEmailOOBCode generates an OOB code for changing email address.
//
// If WidgetURL is not provided in the configuration, the OOBCodeURL field in
// the returned OOBCodeResponse is nil.
func (c *Client) GenerateChangeEmailOOBCode(
	req *http.Request, email, newEmail, token string) (*OOBCodeResponse, error) {
	r := &GetOOBCodeRequest{
		RequestType: ChangeEmailRequestType,
		Email:       email,
		NewEmail:    newEmail,
		Token:       token,
		UserIP:      extractRemoteIP(req),
	}
	resp, err := c.apiClient().GetOOBCode(r)
	if err != nil {
		return nil, err
	}
	return &OOBCodeResponse{
		Action:     OOBActionChangeEmail,
		Email:      email,
		NewEmail:   newEmail,
		OOBCode:    resp.OOBCode,
		OOBCodeURL: c.buildOOBCodeURL(req, OOBActionChangeEmail, resp.OOBCode),
	}, nil
}

// GenerateVerifyEmailOOBCode generates an OOB code for verifying email address.
//
// If WidgetURL is not provided in the configuration, the OOBCodeURL field in
// the returned OOBCodeResponse is nil.
func (c *Client) GenerateVerifyEmailOOBCode(req *http.Request, email string) (*OOBCodeResponse, error) {
	r := &GetOOBCodeRequest{
		RequestType: VerifyEmailRequestType,
		Email:       email,
		UserIP:      extractRemoteIP(req),
	}
	resp, err := c.apiClient().GetOOBCode(r)
	if err != nil {
		return nil, err
	}
	return &OOBCodeResponse{
		Action:     OOBActionVerifyEmail,
		Email:      email,
		OOBCode:    resp.OOBCode,
		OOBCodeURL: c.buildOOBCodeURL(req, OOBActionVerifyEmail, resp.OOBCode),
	}, nil
}

func (c *Client) buildOOBCodeURL(req *http.Request, action, oobCode string) *url.URL {
	// Return nil if widget URL is not provided.
	if c.widgetURL == nil {
		return nil
	}
	url := extractRequestURL(req).ResolveReference(c.widgetURL)
	q := url.Query()
	q.Set(c.config.WidgetModeParamName, action)
	q.Set(OOBCodeParam, oobCode)
	url.RawQuery = q.Encode()
	return url
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
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		// Ignore error since GAE returns V6_ADDR instead of [V6_ADDR]:port.
		return req.RemoteAddr
	}
	return host
}

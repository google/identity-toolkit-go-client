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
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
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
	api       *APIClient // Don't use this field directly. Use apiClient() instead.
	jc        *jwt.Config
}

// New creates a Client from the configuration.
func New(ctx context.Context, config *Config) (*Client, error) {
	conf := *config
	certs := &Certificates{URL: publicCertsURL}
	var widgetURL *url.URL
	if conf.WidgetURL != "" {
		var err error
		widgetURL, err = url.Parse(conf.WidgetURL)
		if err != nil {
			return nil, fmt.Errorf("invalid WidgetURL: %s", conf.WidgetURL)
		}
	}
	var jc *jwt.Config
	if config.GoogleAppCredentialsPath != "" {
		b, err := ioutil.ReadFile(config.GoogleAppCredentialsPath)
		if err != nil {
			return nil, fmt.Errorf("invalid GoogleAppCredentialsPath: %v", err)
		}
		jc, err = google.JWTConfigFromJSON(b, identitytoolkitScope)
		if err != nil {
			return nil, err
		}
	}
	api, err := newAPIClient(ctx, jc)
	if err != nil {
		return nil, err
	}
	if err := retrieveProjectConfigIfNeeded(api, &conf); err != nil {
		return nil, fmt.Errorf("unable to retrieve API key and client ID: %#v", err)
	}
	if err := conf.normalize(); err != nil {
		return nil, err
	}
	return &Client{
		config:    &conf,
		widgetURL: widgetURL,
		certs:     certs,
		api:       api,
		jc:        jc,
	}, nil
}

func newAPIClient(ctx context.Context, jc *jwt.Config) (*APIClient, error) {
	var hc *http.Client
	if jc != nil {
		hc = jc.Client(ctx)
	} else {
		var err error
		hc, err = google.DefaultClient(ctx, identitytoolkitScope)
		if err != nil {
			return nil, err
		}
	}
	return &APIClient{
		http.Client{
			Transport: &transport{hc.Transport},
		},
	}, nil
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
func (c *Client) ValidateToken(ctx context.Context, token string) (*Token, error) {
	if err := c.certs.LoadIfNecessary(defaultTransport(ctx)); err != nil {
		return nil, err
	}
	t, err := VerifyToken(token, c.config.ClientID, nil, c.certs)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// UserByToken retrieves the account information of the user specified by the ID
// token.
func (c *Client) UserByToken(ctx context.Context, token string) (*User, error) {
	t, err := c.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	localID := t.LocalID
	providerID := t.ProviderID
	u, err := c.UserByLocalID(ctx, localID)
	if err != nil {
		return nil, err
	}
	u.ProviderID = providerID
	return u, nil
}

// UserByEmail retrieves the account information of the user specified by the
// email address.
func (c *Client) UserByEmail(ctx context.Context, email string) (*User, error) {
	resp, err := c.apiClient(ctx).GetAccountInfo(&GetAccountInfoRequest{Emails: []string{email}})
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
func (c *Client) UserByLocalID(ctx context.Context, localID string) (*User, error) {
	resp, err := c.apiClient(ctx).GetAccountInfo(&GetAccountInfoRequest{LocalIDs: []string{localID}})
	if err != nil {
		return nil, err
	}
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user %s not found", localID)
	}
	return resp.Users[0], nil
}

// UpdateUser updates the account information of the user.
func (c *Client) UpdateUser(ctx context.Context, user *User) error {
	_, err := c.apiClient(ctx).SetAccountInfo(&SetAccountInfoRequest{
		LocalID:       user.LocalID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Password:      user.Password,
		EmailVerified: user.EmailVerified})
	return err
}

// DeleteUser deletes a user specified by the local ID.
func (c *Client) DeleteUser(ctx context.Context, user *User) error {
	_, err := c.apiClient(ctx).DeleteAccount(&DeleteAccountRequest{LocalID: user.LocalID})
	return err
}

// UploadUsers uploads the users to identitytoolkit service.
// algorithm, key, saltSeparator specify the password hash algorithm, signer key
// and separator between password and salt accordingly.
func (c *Client) UploadUsers(ctx context.Context, users []*User, algorithm string, key, saltSeparator []byte) error {
	resp, err := c.apiClient(ctx).UploadAccount(&UploadAccountRequest{users, algorithm, key, saltSeparator})
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
func (c *Client) ListUsersN(ctx context.Context, n int, pageToken string) ([]*User, string, error) {
	resp, err := c.apiClient(ctx).DownloadAccount(&DownloadAccountRequest{n, pageToken})
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

func (l *UserList) start(ctx context.Context) {
	ch := make(chan *User, maxResultsPerPage)
	l.C = ch
	go func() {
		for {
			users, pageToken, err := l.client.ListUsersN(ctx, maxResultsPerPage, l.pageToken)
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
func (l *UserList) Retry(ctx context.Context) {
	if l.Error != nil {
		l.Error = nil
		l.start(ctx)
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
func (c *Client) ListUsers(ctx context.Context) *UserList {
	l := &UserList{client: c}
	l.start(ctx)
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
func (c *Client) GenerateOOBCode(ctx context.Context, req *http.Request) (*OOBCodeResponse, error) {
	switch action := req.PostFormValue(OOBActionParam); action {
	case OOBActionResetPassword:
		return c.GenerateResetPasswordOOBCode(
			ctx,
			req,
			req.PostFormValue(OOBEmailParam),
			req.PostFormValue(OOBCAPTCHAChallengeParam),
			req.PostFormValue(OOBCAPTCHAResponseParam))
	case OOBActionChangeEmail:
		return c.GenerateChangeEmailOOBCode(
			ctx,
			req,
			req.PostFormValue(OOBOldEmailParam),
			req.PostFormValue(OOBNewEmailParam),
			c.TokenFromRequest(req))
	case OOBActionVerifyEmail:
		return c.GenerateVerifyEmailOOBCode(ctx, req, req.PostFormValue(OOBEmailParam))
	default:
		return nil, fmt.Errorf("unrecognized action: %s", action)
	}
}

// GenerateResetPasswordOOBCode generates an OOB code for resetting password.
//
// If WidgetURL is not provided in the configuration, the OOBCodeURL field in
// the returned OOBCodeResponse is nil.
func (c *Client) GenerateResetPasswordOOBCode(
	ctx context.Context, req *http.Request, email, captchaChallenge, captchaResponse string) (*OOBCodeResponse, error) {
	r := &GetOOBCodeRequest{
		RequestType:      ResetPasswordRequestType,
		Email:            email,
		CAPTCHAChallenge: captchaChallenge,
		CAPTCHAResponse:  captchaResponse,
		UserIP:           extractRemoteIP(req),
	}
	resp, err := c.apiClient(ctx).GetOOBCode(r)
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
	ctx context.Context, req *http.Request, email, newEmail, token string) (*OOBCodeResponse, error) {
	r := &GetOOBCodeRequest{
		RequestType: ChangeEmailRequestType,
		Email:       email,
		NewEmail:    newEmail,
		Token:       token,
		UserIP:      extractRemoteIP(req),
	}
	resp, err := c.apiClient(ctx).GetOOBCode(r)
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
func (c *Client) GenerateVerifyEmailOOBCode(
	ctx context.Context, req *http.Request, email string) (*OOBCodeResponse, error) {
	r := &GetOOBCodeRequest{
		RequestType: VerifyEmailRequestType,
		Email:       email,
		UserIP:      extractRemoteIP(req),
	}
	resp, err := c.apiClient(ctx).GetOOBCode(r)
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

func retrieveProjectConfigIfNeeded(apiClient *APIClient, conf *Config) error {
	if conf.BrowserAPIKey != "" && conf.ClientID != "" && conf.WidgetURL != "" && len(conf.SignInOptions) > 0 {
		return nil
	}

	resp, err := apiClient.GetProjectConfig()
	if err != nil {
		return err
	}
	if conf.BrowserAPIKey == "" {
		conf.BrowserAPIKey = resp.APIKey
	}
	var opts []string
	var googleClientID string
	for _, element := range resp.IdpConfigs {
		if element.Provider == "GOOGLE" {
			googleClientID = element.ClientID
		}
		if element.Enabled {
			opts = append(opts, strings.ToLower(element.Provider))
		}
	}
	if googleClientID == "" {
		return fmt.Errorf("should at least have Google Idp Config")
	}
	if conf.ClientID == "" {
		conf.ClientID = googleClientID
	}
	if resp.AllowPasswordUser {
		opts = append(opts, "password")
	}
	if len(conf.SignInOptions) == 0 {
		if len(opts) == 0 {
			return fmt.Errorf("should have at least one sign in option")
		}
		conf.SignInOptions = opts
	}
	return nil
}

// BrowserAPIKey returns browser API key from the config.
func (c *Client) BrowserAPIKey() string {
	return c.config.BrowserAPIKey
}

// SignInOptions returns sign in options from the config.
func (c *Client) SignInOptions() []string {
	return c.config.SignInOptions
}

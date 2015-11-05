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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"google.golang.org/api/googleapi"
)

// Bytes is a slice of bytes.
// It is URL safe Base64 instead of standard Base64 encoded when being
// marshalled to JSON.
type Bytes []byte

// MarshalJSON encodes the byte slice to a web safe base64 string.
func (b Bytes) MarshalJSON() ([]byte, error) {
	return []byte("\"" + base64.URLEncoding.EncodeToString(b) + "\""), nil
}

// UnmarshalJSON decodes a web safe base64 string into a byte slice.
func (b *Bytes) UnmarshalJSON(src []byte) error {
	quoteLength := len("\"")
	dst := make([]byte, base64.URLEncoding.DecodedLen(len(src)))
	l, err := base64.URLEncoding.Decode(dst, src[quoteLength:len(src)-quoteLength])
	if err != nil {
		return err
	}
	*b = dst[:l]
	return nil
}

// TimestampMilli represents the Unix time in milliseconds.
// float64 is used here as the underlying type because the API returns a float.
type TimestampMilli float64

// AsTime converts the TimestampMilli into a time.Time.
func (t TimestampMilli) AsTime() time.Time {
	return time.Unix(0, int64(t*1e6))
}

// String implements the fmt.Stringer interface.
func (t TimestampMilli) String() string {
	return t.AsTime().String()
}

// ProviderUserInfo holds the user information from an identity provider (IDP).
type ProviderUserInfo struct {
	// ProviderID is the identifer for the IDP, usually the TLD, e.g., google.com.
	ProviderID string `json:"providerId,omitempty"`
	// FederatedID is a unique identifier for the user within the IDP.
	FederatedID string `json:"federatedId,omitempty"`
	// DisplayName is the name of the user at the IDP.
	DisplayName string `json:"displayName,omitempty"`
	// PhotoURL is the profile picture URL of the user at the IDP.
	PhotoURL string `json:"photoUrl,omitempty"`
}

// User holds the user account information.
type User struct {
	// LocalID is the locally unique identifier for the user.
	LocalID string `json:"localId,omitempty"`
	// Email is the email address of the user.
	Email string `json:"email,omitempty"`
	// EmailVerified indicates if the email address of the user has been verifed.
	EmailVerified bool `json:"emailVerified,omitempty"`
	// DisplayName is the current name of the user. For instance, if the user
	// currently signs in with Google, the DisplayName is the one from Google IDP.
	DisplayName string `json:"displayName,omitempty"`
	// PhotoURL is the current profile picture URL of the user. For instance, if the
	// user currently signs in with Google, the PhotoURL is the one from Google IDP.
	PhotoURL string `json:"photoUrl,omitempty"`
	// ProviderUserInfo holds user information from all IDPs.
	ProviderUserInfo []ProviderUserInfo `json:"providerUserInfo,omitempty"`
	// PasswordHash is the hashed user password.
	PasswordHash Bytes `json:"passwordHash,omitempty"`
	// PasswordUpdateAt is the Unix time in milliseconds of the last password update.
	PasswordUpdateAt TimestampMilli `json:"passwordUpdateAt,omitempty"`
	// Salt is the salt used for hashing password.
	Salt Bytes `json:"salt,omitempty"`
	// ProviderID, if present, indicates the IDP with which the user signs in.
	ProviderID string `json:"providerId,omitempty"`
	// Password is the raw password of the user. It is only used to set new password.
	Password string `json:"-"`
}

// IdpConfig holds the IDP configuration.
type IdpConfig struct {
	Provider string `json:"provider,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	ClientID string `json:"clientId,omitempty"`
}

// Identitytoolkit API endpoint URL common parts.
var (
	APIBaseURI = "https://www.googleapis.com/identitytoolkit"
	APIVersion = "v3"
	APIPath    = "relyingparty"
)

type apiMethod string

const (
	getAccountInfo   apiMethod = "getAccountInfo"
	setAccountInfo   apiMethod = "setAccountInfo"
	deleteAccount    apiMethod = "deleteAccount"
	uploadAccount    apiMethod = "uploadAccount"
	downloadAccount  apiMethod = "downloadAccount"
	getOOBCode       apiMethod = "getOobConfirmationCode"
	getProjectConfig apiMethod = "getProjectConfig"
)

// URL returns the full URL of the API method.
func (m apiMethod) url() string {
	return strings.Join([]string{APIBaseURI, APIVersion, APIPath, string(m)}, "/")
}

// An APIClient is an HTTP client that sends requests and receives responses
// from identitytoolkit APIs.
//
// The underlying http.Client should add appropriate auth credentials according
// to the auth level of the API.
type APIClient struct {
	http.Client
}

type httpMethod string

const (
	GET  httpMethod = "GET"
	POST httpMethod = "POST"
)

func (c *APIClient) do(httpMethod httpMethod, m apiMethod, body []byte) ([]byte, error) {
	var req *http.Request
	if httpMethod == POST {
		req, _ = http.NewRequest(string(httpMethod), m.url(), bytes.NewReader(body))
	} else {
		req, _ = http.NewRequest(string(httpMethod), m.url(), nil)
	}
	googleapi.SetOpaque(req.URL)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := googleapi.CheckResponse(resp); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (c *APIClient) request(httpMethod httpMethod, m apiMethod, req, resp interface{}) error {
	t := reflect.TypeOf(resp)
	if t.Kind() != reflect.Ptr {
		log.Fatal("Resp must be a pointer.")
	}
	var body []byte
	var err error
	if req != nil {
		body, err = json.Marshal(req)
	}
	if err != nil {
		return err
	}
	body, err = c.do(httpMethod, m, body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return err
	}
	return nil
}

// GetAccountInfoRequest contains the email addresses or user IDs which are used
// to retrieve the user account information.
type GetAccountInfoRequest struct {
	Emails   []string `json:"email,omitempty"`
	LocalIDs []string `json:"localId,omitempty"`
}

// GetAccountInfoResponse contains the user account information specified by the
// corresponding GetAccountInfoRequest upon success.
type GetAccountInfoResponse struct {
	Users []*User `json:"users,omitempty"`
}

// GetAccountInfo retreives the users' account information.
func (c *APIClient) GetAccountInfo(req *GetAccountInfoRequest) (*GetAccountInfoResponse, error) {
	if len(req.Emails) == 0 && len(req.LocalIDs) == 0 {
		return nil, fmt.Errorf("GetAccountInfo: must provide an email or a local ID")
	}

	resp := &GetAccountInfoResponse{}
	if err := c.request(POST, getAccountInfo, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SetAccountInfoRequest contains account information to update.
// Either LocalID or Email should be provided to find the account.
// The Password field contains the new raw password if provided.
type SetAccountInfoRequest struct {
	LocalID       string `json:"localId,omitempty"`
	Email         string `json:"email,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	Password      string `json:"password,omitempty"`
	EmailVerified bool   `json:"emailVerified,omitempty"`
}

// SetAccountInfoResponse is the response for a SetAccountInfoRequest upon success.
// It is an empty response.
type SetAccountInfoResponse struct {
}

// SetAccountInfo updates the account information.
func (c *APIClient) SetAccountInfo(req *SetAccountInfoRequest) (*SetAccountInfoResponse, error) {
	if req.Email == "" && req.LocalID == "" {
		return nil, fmt.Errorf("SetAccountInfo: must provide an email or a local ID")
	}

	resp := &SetAccountInfoResponse{}
	if err := c.request(POST, setAccountInfo, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteAccountRequest contains the user ID to be deleted.
type DeleteAccountRequest struct {
	LocalID string `json:"localId,omitempty"`
}

// DeleteAccountResponse is the response for a DeleteAccountRequest upon success.
// It is an empty response.
type DeleteAccountResponse struct {
}

// DeleteAccount deletes an account.
func (c *APIClient) DeleteAccount(req *DeleteAccountRequest) (*DeleteAccountResponse, error) {
	if req.LocalID == "" {
		return nil, fmt.Errorf("DeleteAccount: must provide a local ID")
	}

	resp := &DeleteAccountResponse{}
	if err := c.request(POST, deleteAccount, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UploadAccountRequest the account information of users to upload.
// The hash algorithm and signer key for the password are required.
type UploadAccountRequest struct {
	Users         []*User `json:"users,omitempty"`
	HashAlgorithm string  `json:"hashAlgorithm,omitempty"`
	SignerKey     Bytes   `json:"signerKey,omitempty"`
	SaltSeparator Bytes   `json:"saltSeparator,omitempty"`
}

// UploadError is the error object for partial upload failure.
type UploadError []*struct {
	// Index indicates the index of the failed account.
	Index int `json:"index,omitempty"`
	// Message is the uploading error message for the failed account.
	Message string `json:"message,omitempty"`
}

// Error implements error interface.
func (e UploadError) Error() string {
	var b bytes.Buffer
	for _, v := range e {
		fmt.Fprintf(&b, "{%d: %s}", v.Index, v.Message)
	}
	return b.String()
}

// UploadAccountResponse contains the error information if some accounts are
// failed to upload.
type UploadAccountResponse struct {
	Error UploadError `json:"error,omitempty"`
}

// UploadAccount uploads accounts to identitytoolkit service.
func (c *APIClient) UploadAccount(req *UploadAccountRequest) (*UploadAccountResponse, error) {
	if len(req.Users) == 0 {
		return nil, fmt.Errorf("UploadAccount: must provide at lease one account")
	}
	if req.HashAlgorithm == "" {
		return nil, fmt.Errorf("UploadAccount: must provide the hash algorithm")
	}
	if len(req.SignerKey) == 0 {
		return nil, fmt.Errorf("UploadAccount: must provide the signer key")
	}

	resp := &UploadAccountResponse{}
	if err := c.request(POST, uploadAccount, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DownloadAccountRequest contains the information for downloading accounts.
// MaxResults specifies the max number of accounts in one response.
// NextPageToken should be empty for the first request and the value from the
// previous response afterwards.
type DownloadAccountRequest struct {
	MaxResults    int    `json:"maxResults,omitempty"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// DownloadAccountResponse contains the downloaded accounts and the page token
// for next request.
type DownloadAccountResponse struct {
	Users         []*User `json:"users,omitempty"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
}

// DownloadAccount donwloads accounts from identitytoolkit service.
func (c *APIClient) DownloadAccount(req *DownloadAccountRequest) (*DownloadAccountResponse, error) {
	resp := &DownloadAccountResponse{}
	if err := c.request(POST, downloadAccount, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Request types accepted by identitytoolkit getOobConfirmationCode API.
const (
	ResetPasswordRequestType = "PASSWORD_RESET"
	ChangeEmailRequestType   = "NEW_EMAIL_ACCEPT"
	VerifyEmailRequestType   = "VERIFY_EMAIL"
)

// GetOOBCodeRequest contains the information to get an OOB code
// from identitytoolkit service.
//
// There are three kinds of OOB code:
//
// 1. OOB code for password recovery. The RequestType should be PASSWORD_RESET
// and Email, CAPTCHAChallenge and CAPTCHAResponse are required.
//
// 2. OOB code for email change. The RequestType should be NEW_EMAIL_ACCEPT and
// Email, newEmail and Token are required.
//
// 3. OOB code for email verification. The RequestType should be VERIFY_EMAIL
// and Email is required.
type GetOOBCodeRequest struct {
	RequestType      string `json:"requestType,omitempty"`
	Email            string `json:"email,omitempty"`
	CAPTCHAChallenge string `json:"challenge,omitempty"`
	CAPTCHAResponse  string `json:"captchaResp,omitempty"`
	NewEmail         string `json:"newEmail,omitempty"`
	Token            string `json:"idToken,omitempty"`
	UserIP           string `json:"userIp,omitempty"`
}

// GetOOBCodeResponse contains the OOB code upon success.
type GetOOBCodeResponse struct {
	OOBCode string `json:"oobCode,omitempty"`
}

// GetOOBCode retrieves an OOB code.
func (c *APIClient) GetOOBCode(req *GetOOBCodeRequest) (*GetOOBCodeResponse, error) {
	switch req.RequestType {

	case ResetPasswordRequestType:
		if req.Email == "" {
			return nil, fmt.Errorf("GetOOBCode: must provide an email")
		}
		if req.CAPTCHAResponse == "" {
			return nil, fmt.Errorf("GetOOBCode: must provide CAPTCHA response")
		}

	case ChangeEmailRequestType:
		if req.Email == "" {
			return nil, fmt.Errorf("GetOOBCode: must provide the old email")
		}
		if req.NewEmail == "" {
			return nil, fmt.Errorf("GetOOBCode: must provide the new email")
		}
		if req.Token == "" {
			return nil, fmt.Errorf("GetOOBCode: must provide the Gitkit token")
		}

	case VerifyEmailRequestType:
		if req.Email == "" {
			return nil, fmt.Errorf("GetOOBCode: must provide an email")
		}

	default:
		return nil, fmt.Errorf("GetOOBCode: unrecognized request type [%s]", req.RequestType)
	}

	resp := &GetOOBCodeResponse{}
	if err := c.request(POST, getOOBCode, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetProjectConfigResponse contains the project ID, API key, whether password login is
// enabled and a list of IDP configs.
type GetProjectConfigResponse struct {
	ProjectID         string       `json:"projectId,omitempty"`
	APIKey            string       `json:"apiKey,omitempty"`
	AllowPasswordUser bool         `json:"allowPasswordUser,omitempty"`
	IdpConfigs        []*IdpConfig `json:"idpConfig,omitempty"`
}

// GetProjectConfig retrieves the configuration information for the project.
func (c *APIClient) GetProjectConfig() (*GetProjectConfigResponse, error) {
	resp := &GetProjectConfigResponse{}
	if err := c.request(GET, getProjectConfig, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

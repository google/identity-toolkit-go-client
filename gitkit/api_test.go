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
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestBytes(t *testing.T) {
	b := Bytes{250, 252, 195, 135, 113, 40, 49, 187, 250, 93, 111}
	encoded := []byte(`"-vzDh3EoMbv6XW8="`)
	r, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("unable to json.Marshal(%v): %v", b, err)
	}

	if !bytes.Equal(r, encoded) {
		t.Errorf("json.Marshal(b) = %v; want %v", r, encoded)
	}
}

func TestAPIMethod(t *testing.T) {
	methodTests := []struct {
		m   apiMethod
		url string
	}{
		{getAccountInfo, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/getAccountInfo"},
		{setAccountInfo, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/setAccountInfo"},
		{deleteAccount, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/deleteAccount"},
		{uploadAccount, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/uploadAccount"},
		{downloadAccount, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/downloadAccount"},
		{getOOBCode, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/getOobConfirmationCode"},
		{getProjectConfig, "https://www.googleapis.com/identitytoolkit/v3/relyingparty/getProjectConfig"},
	}
	for i, mt := range methodTests {
		if mt.m.url() != mt.url {
			t.Errorf("%d. url() = %q; want %q", i, mt.m.url(), mt.url)
		}
	}
}

func prepareClient(err bool, respBody string) *APIClient {
	var statusCode int
	if err {
		statusCode = 403
	} else {
		statusCode = 200
	}
	return &APIClient{http.Client{Transport: &roundTripper{statusCode, respBody}}}
}

func TestGetAccountInfo(t *testing.T) {
	getAccountTests := []struct {
		name string
		req  *GetAccountInfoRequest
		err  bool
		json string
		resp *GetAccountInfoResponse
	}{
		{
			"no_email_and_user_ID",
			&GetAccountInfoRequest{},
			true,
			"",
			nil,
		},
		{
			"api_error",
			&GetAccountInfoRequest{Emails: []string{"user@example.com"}},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
			nil,
		},
		{
			"success",
			&GetAccountInfoRequest{LocalIDs: []string{"12345"}},
			false,
			`{"users": [{"localId": "12345", "email": "user@example.com", "emailVerified": true}]}`,
			&GetAccountInfoResponse{[]*User{{LocalID: "12345", Email: "user@example.com", EmailVerified: true}}},
		},
	}
	for _, gt := range getAccountTests {
		c := prepareClient(gt.err, gt.json)
		resp, err := c.GetAccountInfo(gt.req)
		if gt.err && err == nil {
			t.Errorf("%s: GetAccountInfo() = %v, nil; want nil, err", gt.name, resp)
		}
		if !gt.err {
			if err != nil || resp == nil {
				t.Errorf("%s: GetAccountInfo() = %v, %v; want %v, nil", gt.name, resp, err, gt.resp)
			} else if !reflect.DeepEqual(*resp.Users[0], *gt.resp.Users[0]) {
				t.Errorf("%s: GetAccountInfo() returns %v; want %v", gt.name, *resp.Users[0], *gt.resp.Users[0])
			}
		}
	}
}

func TestSetAccountInfo(t *testing.T) {
	setAccountTests := []struct {
		name string
		req  *SetAccountInfoRequest
		err  bool
		json string
	}{
		{
			"no_email_and_user_ID",
			&SetAccountInfoRequest{},
			true,
			"",
		},
		{
			"api_error",
			&SetAccountInfoRequest{Email: "user@example.com", DisplayName: "Test User"},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
		},
		{
			"success",
			&SetAccountInfoRequest{LocalID: "12345", DisplayName: "Test User"},
			false,
			"{}",
		},
	}
	for _, st := range setAccountTests {
		c := prepareClient(st.err, st.json)
		resp, err := c.SetAccountInfo(st.req)
		if st.err && err == nil {
			t.Errorf("%s: SetAccountInfo() = %v, nil; want nil, err", st.name, resp)
		}
		if !st.err && (err != nil || resp == nil) {
			t.Errorf("%s: SetAccountInfo() = %v, %v; want non-nil, nil", st.name, resp, err)
		}
	}
}

func TestDeleteAccount(t *testing.T) {
	deleteAccountTests := []struct {
		name string
		req  *DeleteAccountRequest
		err  bool
		json string
	}{
		{
			"no_user_ID",
			&DeleteAccountRequest{},
			true,
			"",
		},
		{
			"api_error",
			&DeleteAccountRequest{"12345"},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
		},
		{
			"success",
			&DeleteAccountRequest{"12345"},
			false,
			"{}",
		},
	}
	for _, dt := range deleteAccountTests {
		c := prepareClient(dt.err, dt.json)
		resp, err := c.DeleteAccount(dt.req)
		if dt.err && err == nil {
			t.Errorf("%s: DeleteAccountInfo() = %v, nil; want nil, err", dt.name, resp)
		}
		if !dt.err && (err != nil || resp == nil) {
			t.Errorf("%s: DeleteAccountInfo() = %v, %v; want non-nil, nil", dt.name, resp, err)
		}
	}
}

func TestUploadAccount(t *testing.T) {
	uploadAccountTests := []struct {
		name string
		req  *UploadAccountRequest
		err  bool
		json string
		resp *UploadAccountResponse
	}{
		{
			"no_account",
			&UploadAccountRequest{},
			true,
			"",
			nil,
		},
		{
			"no_hash_alg_and_key",
			&UploadAccountRequest{Users: []*User{{LocalID: "12345"}}},
			true,
			"",
			nil,
		},
		{
			"no_key",
			&UploadAccountRequest{Users: []*User{{LocalID: "12345"}}, HashAlgorithm: "HMAC_SHA1"},
			true,
			"",
			nil,
		},
		{
			"api_error",
			&UploadAccountRequest{
				Users:         []*User{{LocalID: "12345"}},
				HashAlgorithm: "HMAC_SHA1",
				SignerKey:     Bytes{123},
			},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
			nil,
		},
		{
			"success",
			&UploadAccountRequest{
				Users:         []*User{{LocalID: "12345"}},
				HashAlgorithm: "HMAC_SHA1",
				SignerKey:     Bytes{123},
			},
			false,
			"{}",
			&UploadAccountResponse{},
		},
		{
			"partial_success",
			&UploadAccountRequest{
				Users:         []*User{{LocalID: "12345"}},
				HashAlgorithm: "HMAC_SHA1",
				SignerKey:     Bytes{123},
			},
			false,
			`{"error": [{"index": 0, "message": "upload error"}]}`,
			&UploadAccountResponse{UploadError{{0, "upload error"}}},
		},
	}
	for _, ut := range uploadAccountTests {
		c := prepareClient(ut.err, ut.json)
		resp, err := c.UploadAccount(ut.req)
		if ut.err && err == nil {
			t.Errorf("%s: UploadAccount() = %v, nil; want nil, err", ut.name, resp)
		}
		if !ut.err {
			if err != nil || resp == nil || len(resp.Error) != len(ut.resp.Error) {
				t.Errorf("%s: UploadAccount() = %v, %v; want %v, nil", ut.name, resp, err, ut.resp)
			} else {
				for k, e := range resp.Error {
					if *e != *ut.resp.Error[k] {
						t.Errorf("%s: UploadAccount() returns error %+v; want %+v", ut.name, e, ut.resp.Error[k])
					}
				}
			}
		}
	}
}

func TestDownloadAccount(t *testing.T) {
	downloadAccountTests := []struct {
		name string
		req  *DownloadAccountRequest
		err  bool
		json string
		resp *DownloadAccountResponse
	}{
		{
			"api_error",
			&DownloadAccountRequest{},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
			nil,
		},
		{
			"first_request",
			&DownloadAccountRequest{5, ""},
			false,
			`{"users": [{"localId": "123"}], "nextPageToken": "abcde"}`,
			&DownloadAccountResponse{[]*User{{LocalID: "123"}}, "abcde"},
		},
		{
			"next_request",
			&DownloadAccountRequest{5, "abcde"},
			false,
			`{"users": [{"localId": "456"}, {"localId": "789"}]}`,
			&DownloadAccountResponse{[]*User{{LocalID: "456"}, {LocalID: "789"}}, ""},
		},
	}
	for _, dt := range downloadAccountTests {
		c := prepareClient(dt.err, dt.json)
		resp, err := c.DownloadAccount(dt.req)
		if dt.err && err == nil {
			t.Errorf("%s: DownloadAccount() = %v, nil; want nil, err", dt.name, resp)
		}
		if !dt.err {
			if err != nil ||
				resp == nil ||
				resp.NextPageToken != dt.resp.NextPageToken ||
				len(resp.Users) != len(dt.resp.Users) {
				t.Errorf("%s: DownloadAccount() = %v, %v; want %v, nil", dt.name, resp, err, dt.resp)
			} else {
				for k, u := range resp.Users {
					if !reflect.DeepEqual(*u, *dt.resp.Users[k]) {
						t.Errorf("%s: DownloadAccount() returns user %v; want %v", dt.name, *u, dt.resp.Users[k])
					}
				}
			}
		}
	}
}

func TestGetOOBCode(t *testing.T) {
	getOOBCodeTestss := []struct {
		name string
		req  *GetOOBCodeRequest
		err  bool
		json string
		resp *GetOOBCodeResponse
	}{
		{
			"unknown_request_type",
			&GetOOBCodeRequest{RequestType: "UNKNOWN"},
			true,
			"",
			nil,
		},
		{
			"reset_password_no_email_and_challenge_and_response",
			&GetOOBCodeRequest{RequestType: ResetPasswordRequestType},
			true,
			"",
			nil,
		},
		{
			"reset_password_no_challenge_and_response",
			&GetOOBCodeRequest{RequestType: ResetPasswordRequestType, Email: "user@example.com"},
			true,
			"",
			nil,
		},
		{
			"reset_password_no_response",
			&GetOOBCodeRequest{
				RequestType:      ResetPasswordRequestType,
				Email:            "user@example.com",
				CAPTCHAChallenge: "123",
			},
			true,
			"",
			nil,
		},
		{
			"reset_password_api_error",
			&GetOOBCodeRequest{
				RequestType:      ResetPasswordRequestType,
				Email:            "user@example.com",
				CAPTCHAChallenge: "123",
				CAPTCHAResponse:  "abc",
			},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
			nil,
		},
		{
			"reset_password_success",
			&GetOOBCodeRequest{
				RequestType:      ResetPasswordRequestType,
				Email:            "user@example.com",
				CAPTCHAChallenge: "123",
				CAPTCHAResponse:  "abc",
			},
			false,
			`{"oobCode": "123abc"}`,
			&GetOOBCodeResponse{"123abc"},
		},
		{
			"change_email_no_email_and_new_email_and_token",
			&GetOOBCodeRequest{RequestType: ChangeEmailRequestType},
			true,
			"",
			nil,
		},
		{
			"change_email_no_new_email_and_token",
			&GetOOBCodeRequest{RequestType: ChangeEmailRequestType, Email: "user@example.com"},
			true,
			"",
			nil,
		},
		{
			"change_email_no_token",
			&GetOOBCodeRequest{
				RequestType: ChangeEmailRequestType,
				Email:       "user@example.com",
				NewEmail:    "newuser@example.com",
			},
			true,
			"",
			nil,
		},
		{
			"change_email_api_error",
			&GetOOBCodeRequest{
				RequestType: ChangeEmailRequestType,
				Email:       "user@example.com",
				NewEmail:    "newuser@example.com",
				Token:       "token",
			},
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
			nil,
		},
		{
			"change_email_success",
			&GetOOBCodeRequest{
				RequestType: ChangeEmailRequestType,
				Email:       "user@example.com",
				NewEmail:    "newuser@example.com",
				Token:       "token",
			},
			false,
			`{"oobCode": "123abc"}`,
			&GetOOBCodeResponse{"123abc"},
		},
	}
	for _, gt := range getOOBCodeTestss {
		c := prepareClient(gt.err, gt.json)
		resp, err := c.GetOOBCode(gt.req)
		if gt.err && err == nil {
			t.Errorf("%s: GetOOBConfirmationCode() = %v, nil; want nil, err", gt.name, resp)
		}
		if !gt.err && (err != nil || resp == nil || resp.OOBCode != gt.resp.OOBCode) {
			t.Errorf("%s: DownloadAccount() = %v, %v; want %v, nil", gt.name, resp, err, gt.resp)
		}
	}
}

func TestGetProjectConfig(t *testing.T) {
	getConfigTests := []struct {
		name string
		err  bool
		json string
		resp *GetProjectConfigResponse
	}{
		{
			"api_error",
			true,
			`{"error": {"code": 403, "errors": [{"reason": "accessNotConfigured"}]}}`,
			nil,
		},
		{
			"success",
			false,
			`{"projectId": "project_id", "apiKey": "api_key", "allowPasswordUser": true, "idpConfig": [{"provider": "GOOGLE", "clientId": "client_id"}]}`,
			&GetProjectConfigResponse{ProjectID: "project_id", APIKey: "api_key", AllowPasswordUser: true, IdpConfigs: []*IdpConfig{{Provider: "GOOGLE", ClientID: "client_id"}}},
		},
	}
	for _, gt := range getConfigTests {
		c := prepareClient(gt.err, gt.json)
		resp, err := c.GetProjectConfig()
		if gt.err && err == nil {
			t.Errorf("%s: GetProjectConfig() = %v, nil; want nil, err", gt.name, resp)
		}
		if !gt.err && (err != nil || resp == nil || len(resp.IdpConfigs) != len(gt.resp.IdpConfigs)) {
			t.Errorf("%s: GetProjectConfig() = %v, %v; want %v, nil", gt.name, resp, err, gt.resp)
		}
	}

}

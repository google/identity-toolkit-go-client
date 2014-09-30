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
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "crypto/sha256" // sha256 is required to check the token signature.
)

// Token is a verified ID token issued by identitytoolkit service.
// See http://openid.net/specs/openid-connect-core-1_0.html#IDToken for more
// information about ID token.
type Token struct {
	// Issuer identifies the issuer of the ID token.
	Issuer string
	// Audience identifies the client that this ID token is intended for.
	Audience string
	// IssueAt is the time at which this ID token was issued.
	IssueAt time.Time
	// ExpireAt is the expiration time on or after which the ID token must not
	// be accepted for processing.
	ExpireAt time.Time
	// LocalID is the locally unique identifier within the client for the user.
	LocalID string
	// Email is the email address of the user.
	Email string
	// EmailVerified indicates whether or not the email address of the user has
	// been verifed.
	EmailVerified bool
	// ProviderID is the identifier for the identity provider (IDP) for the
	// user. It is usually the top level domain of the IDP, e.g., google.com.
	ProviderID string
}

// Expired checks whether or not the ID token is expired.
func (t *Token) Expired() bool {
	return time.Now().After(t.ExpireAt)
}

// VerifyToken verifies the JWT is valid and signed by identitytoolkit service
// and returns the verfied token.
func VerifyToken(token string, certs *Certificates) (*Token, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("not a JWT: %s", token)
	}
	// Check the header to extract the "kid" field.
	h, err := decodeSegment(parts[0])
	if err != nil {
		return nil, err
	}
	hh := struct {
		KeyID string `json:"kid"`
	}{}
	if err = json.Unmarshal(h, &hh); err != nil {
		return nil, err
	}
	cert, err := certs.Cert(hh.KeyID)
	if err != nil {
		return nil, err
	}
	// Check the claim set.
	c, err := decodeSegment(parts[1])
	if err != nil {
		return nil, err
	}
	t := struct {
		Iss        string `json:"iss,omitempty"`
		Aud        string `json:"aud,omitempty"`
		Iat        int64  `json:"iat,omitempty"`
		Exp        int64  `json:"exp,omitempty"`
		UserID     string `json:"user_id,omitempty"`
		Email      string `json:"email,omitempty"`
		Verified   bool   `json:"verified,omitempty"`
		ProviderID string `json:"providerId,omitempty"`
	}{}
	if err = json.Unmarshal(c, &t); err != nil {
		return nil, err
	}
	if t.Iss == "" || t.Aud == "" || t.Iat == 0 || t.Exp == 0 || t.UserID == "" {
		return nil, fmt.Errorf("missing required fields: %v", t)
	}
	// Check the signature.
	s, err := decodeSegment(parts[2])
	if err != nil {
		return nil, err
	}
	if err := cert.CheckSignature(x509.SHA256WithRSA, []byte(parts[0]+"."+parts[1]), s); err != nil {
		return nil, err
	}
	return &Token{
		Issuer:        t.Iss,
		Audience:      t.Aud,
		IssueAt:       time.Unix(t.Iat, 0),
		ExpireAt:      time.Unix(t.Exp, 0),
		LocalID:       t.UserID,
		Email:         t.Email,
		EmailVerified: t.Verified,
		ProviderID:    t.ProviderID,
	}, nil
}

// decodeSegment decodes the Base64 encoding segment of the JWT token.
// It pads the string if necessary.
func decodeSegment(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s = s + "=="
	case 3:
		s = s + "="
	}
	return base64.URLEncoding.DecodeString(s)
}

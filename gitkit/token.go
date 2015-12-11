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
	"errors"
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
	// DisplayName is the name that the user wants to be referred to.
	DisplayName string
	// PhotoURL is the URL of the user's profile picture.
	PhotoURL string
	// The token string.
	TokenString string
}

// Expired checks whether or not the ID token is expired.
func (t *Token) Expired() bool {
	return time.Now().After(t.ExpireAt)
}

// Errors that can be returned from the VerifyToken function.
var (
	ErrMalformed        = errors.New("malfored token")
	ErrInvalidAlgorithm = errors.New("invalid algorithm")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrInvalidAudience  = errors.New("invalid audience")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrKeyNotFound      = errors.New("key not found")
	ErrExpired          = errors.New("token expired")
	ErrMissingAudience  = errors.New("missing audiences for token validation")
)

// VerifyToken verifies the JWT is valid and signed by identitytoolkit service
// and returns the verfied token. A token is valid if and only if it passes the
// following checks:
// 1. The value of "iss" field is one of the issuers if issuers is not nil;
// 2. The value of "aud" field is the same as the audience;
// 3. The token is not expired according to the "exp" field;
// 4. The signature can be verified from one of the certs;
func VerifyToken(token string, audiences []string, issuers []string, certs *Certificates) (*Token, error) {
	if len(audiences) == 0 {
		return nil, ErrMissingAudience
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrMalformed
	}
	// Check the claim set.
	c, err := decodeSegment(parts[1])
	if err != nil {
		return nil, ErrMalformed
	}
	claims := struct {
		Iss         string `json:"iss,omitempty"`
		Aud         string `json:"aud,omitempty"`
		Iat         int64  `json:"iat,omitempty"`
		Exp         int64  `json:"exp,omitempty"`
		UserID      string `json:"user_id,omitempty"`
		Email       string `json:"email,omitempty"`
		Verified    bool   `json:"verified,omitempty"`
		ProviderID  string `json:"provider_id,omitempty"`
		DisplayName string `json:"display_name,omitempty"`
		PhotoURL    string `json:"photo_url,omitempty"`
	}{}
	if err = json.Unmarshal(c, &claims); err != nil {
		return nil, ErrMalformed
	}
	if issuers != nil && !inArray(issuers, claims.Iss) {
		return nil, ErrInvalidIssuer
	}
	if !inArray(audiences, claims.Aud) {
		return nil, ErrInvalidAudience
	}
	exp := time.Unix(claims.Exp, 0)
	if time.Now().After(exp) {
		return nil, ErrExpired
	}
	// Check the header to extract the "kid" field.
	h, err := decodeSegment(parts[0])
	if err != nil {
		return nil, err
	}
	header := struct {
		Algorithm string `json:"alg,omitempty"`
		KeyID     string `json:"kid,omitempty"`
	}{}
	if err = json.Unmarshal(h, &header); err != nil {
		return nil, ErrMalformed
	}
	if header.Algorithm != "RS256" {
		return nil, ErrInvalidAlgorithm
	}
	cert, err := certs.Cert(header.KeyID)
	if err != nil {
		return nil, ErrKeyNotFound
	}
	// Check the signature.
	signature, err := decodeSegment(parts[2])
	if err != nil {
		return nil, ErrMalformed
	}
	if err := cert.CheckSignature(x509.SHA256WithRSA, []byte(parts[0]+"."+parts[1]), signature); err != nil {
		return nil, ErrInvalidSignature
	}
	return &Token{
		Issuer:        claims.Iss,
		Audience:      claims.Aud,
		IssueAt:       time.Unix(claims.Iat, 0),
		ExpireAt:      time.Unix(claims.Exp, 0),
		LocalID:       claims.UserID,
		Email:         claims.Email,
		EmailVerified: claims.Verified,
		ProviderID:    claims.ProviderID,
		DisplayName:   claims.DisplayName,
		PhotoURL:      claims.PhotoURL,
		TokenString:   token,
	}, nil
}

func inArray(a []string, e string) bool {
	for _, v := range a {
		if v == e {
			return true
		}
	}
	return false
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

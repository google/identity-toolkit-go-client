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
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func TestExpired(t *testing.T) {
	tokenTests := []struct {
		t       *Token
		expired bool
	}{
		{
			&Token{
				Issuer:   "https://gitkit.google.com/",
				Audience: "217923393573.apps.googleusercontent.com",
				IssueAt:  time.Now(),
				ExpireAt: time.Now().Add(1 * time.Hour),
				LocalID:  "16109857760607106080",
			},
			false,
		},
		{
			&Token{
				Issuer:   "https://gitkit.google.com/",
				Audience: "217923393573.apps.googleusercontent.com",
				IssueAt:  time.Now().Add(-2 * time.Hour),
				ExpireAt: time.Now().Add(-1 * time.Hour),
				LocalID:  "16109857760607106080",
			},
			true,
		},
	}
	for i, tt := range tokenTests {
		expired := tt.t.Expired()
		if expired != tt.expired {
			t.Errorf("%d. t.Expired() = %v; want %v", i, expired, tt.expired)
		}
	}
}

func initCerts() *Certificates {
	// The tokens for testing were signed by this certificate.
	block, _ := pem.Decode([]byte(`-----BEGIN CERTIFICATE-----
MIIC9TCCAd2gAwIBAgIJAPtVrspLw1euMA0GCSqGSIb3DQEBBQUAMBExDzANBgNV
BAMMBkdpdGtpdDAeFw0xNDA1MTQxODA3NTFaFw0yNDA1MTExODA3NTFaMBExDzAN
BgNVBAMMBkdpdGtpdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMlW
d4+FKL7LKWunQok2QAv0p3bpRgymH5la/eXmvruJNFzL73e8PZNMKuHdcWoDCCh2
YQJ2C72H1tDuQ/PnefP/lCauL65WPcznEzrDqGebV99uUzCCZp+yMzDKXbNbQNjH
1jYjJH//gBWX8YNeI8gXkXZOC1WVmhcTM/RbP9t6U7jQWMBszN8cmWo4BISfYKd5
PeapUD6tNGyDGuvWIgHJsq+EVWKYFqnflj3UThGb0S7AmiBk3noS0es3HlD606H2
zqPNiSN/z03yfm5viM10auToMp38On+M9KSf2LlJZhg/q65uQ342v1d0Ez8qsxaU
Ej1zQ+znaU+tzK8ih7UCAwEAAaNQME4wHQYDVR0OBBYEFMBgjduuobBMA48dmjUO
HHT0o5iTMB8GA1UdIwQYMBaAFMBgjduuobBMA48dmjUOHHT0o5iTMAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQEFBQADggEBAI+OtyHQbCxQ07DZnUaYbb+1wtWWBe9j
TOhOPhFPoW8N8ltftT18ey9FEdeOA/3x28iIBSTpMIya/RT9DCmbcc5MseskSJId
1z5MctAeSJrDp9OmF9jQJsnMtkjyNbzc3IOEA8uAby/iRccjuEvmGrra4teHxEOt
7vDtOt/AgMunX14bLCPhRFJHSF1FSGbE/iuhwbglW3nYQ8V7giYBMhD5pYKRlSN3
tyUpuTqGZUKAmXL44F8sqcluMY3l5OKMHabXFNRwCLG8QOqnj/nOzeFlq8zJD7y+
JeyA5HMQFyXEsbPg/i9Xwwx0jerQQHVwG4y2Ew6IPAjX7kSZC2RGX5U=
-----END CERTIFICATE-----`))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return &Certificates{certs: map[string]*x509.Certificate{"40QoZg": cert}}
}

const validToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjQwUW9aZyJ9.eyJpc3MiOiJodHRwczov" +
	"L2dpdGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzL" +
	"mdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIj" +
	"oxNDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJ" +
	"lbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJn" +
	"b29nbGUuY29tIiwidmVyaWZpZWQiOnRydWV9.U_A882B6Msd3D3n0ADLeAuuk" +
	"spTI1DDmDzn6NkXKN1oUhuk8E0scVRziYYz4kMBvlRo0RynWe-VijOt4v2uYy" +
	"xLvPD176FsJmTqdVcUJUDtkhzzmthB1ndtezgNr-HrpZAVpJd0fy8eJ7zmNIG" +
	"AI8OpoWk5Ku9IsC2DcOwt4Hi3daeFRm0uceO4C27lez3loHDmIG-zWSWoNjic" +
	"GNJ5qNya7lowSYOQOEgBcvVQeOrMz26B09SavrJ9rYAER-KgPXDxvOD7_7IYd" +
	"4ja2ThVG0RjKfvDLCgmg6nMIRl0ZW-Bn3FhwHo74NNF7KMF_QuzAvD2hnuesa" +
	"JlaIdAFyIk9WA"

func TestVerifyToken(t *testing.T) {
	certs := initCerts()
	tokenTests := []struct {
		s string
		t *Token
	}{
		{
			validToken,
			&Token{
				Issuer:        "https://gitkit.google.com/",
				Audience:      "217923393573.apps.googleusercontent.com",
				IssueAt:       time.Unix(1400437715, 0),
				ExpireAt:      time.Unix(1401647315, 0),
				LocalID:       "16109857760607106080",
				Email:         "gitkittest@gmail.com",
				EmailVerified: true,
				TokenString:   validToken,
			},
		},
		{
			// Invalid token which has incorrect signature (last byte changed).
			"eyJhbGciOiJSUzI1NiIsImtpZCI6IjQwUW9aZyJ9.eyJpc3MiOiJodHRwczov" +
				"L2dpdGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzL" +
				"mdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIj" +
				"oxNDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJ" +
				"lbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJn" +
				"b29nbGUuY29tIiwidmVyaWZpZWQiOnRydWV9.U_A882B6Msd3D3n0ADLeAuuk" +
				"spTI1DDmDzn6NkXKN1oUhuk8E0scVRziYYz4kMBvlRo0RynWe-VijOt4v2uYy" +
				"xLvPD176FsJmTqdVcUJUDtkhzzmthB1ndtezgNr-HrpZAVpJd0fy8eJ7zmNIG" +
				"AI8OpoWk5Ku9IsC2DcOwt4Hi3daeFRm0uceO4C27lez3loHDmIG-zWSWoNjic" +
				"GNJ5qNya7lowSYOQOEgBcvVQeOrMz26B09SavrJ9rYAER-KgPXDxvOD7_7IYd" +
				"4ja2ThVG0RjKfvDLCgmg6nMIRl0ZW-Bn3FhwHo74NNF7KMF_QuzAvD2hnuesa" +
				"JlaIdAFyIk9Wa",
			nil,
		},
		{
			// Invalid token which has no "kid" in the header.
			"eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJodHRwczovL2dpdGtpdC5nb29nbGUu" +
				"Y29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250Z" +
				"W50LmNvbSIsImlhdCI6MTM5NTI2NzAwOSwiZXhwIjoxMzk2NDc2NjA5LCJ1c2" +
				"VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHR" +
				"lc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVy" +
				"aWZpZWQiOnRydWV9.hBVTNG3CxqHPQ0ciP1QyKJmSeKnmU5b2Izz_F4uKCx66" +
				"O_tWcthb9TCZAuaKKFZm6jBjkE8c_Mzi7SrTGQOilv5OcjDGeWmfeGx_zX_HY" +
				"MoCtMiwkd74SbA8XIIzakWdO4qQv8JKUGQ0WAmsxZcu9RY2zDDE0P4vYruhjj" +
				"U2lGjggQ2U73K5B2UAtNMtBQMrAXsq-O7sbsC7CKeDd1mhlvuj0R3cEik7LPI" +
				"G9l07vCbDhPJ-8n5QKKXbTV0OSL-5K4Q4hwn_ev4rycyXnUK-2597mRWZL5BP" +
				"7Srh20eySZikebrGyRkzUgZHhXVQHI2BPJCjHbNWsYyFlemEZd7D1g",
			nil,
		},
		{
			// Invalid token whose "kid" is not found (kid is changed from
			// 40QoZg to 40QoTg).
			"eyJhbGciOiJSUzI1NiIsImtpZCI6IjQwUW9UZyJ9.eyJpc3MiOiJodHRwczov" +
				"L2dpdGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzL" +
				"mdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIj" +
				"oxNDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJ" +
				"lbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJn" +
				"b29nbGUuY29tIiwidmVyaWZpZWQiOnRydWV9.U_A882B6Msd3D3n0ADLeAuuk" +
				"spTI1DDmDzn6NkXKN1oUhuk8E0scVRziYYz4kMBvlRo0RynWe-VijOt4v2uYy" +
				"xLvPD176FsJmTqdVcUJUDtkhzzmthB1ndtezgNr-HrpZAVpJd0fy8eJ7zmNIG" +
				"AI8OpoWk5Ku9IsC2DcOwt4Hi3daeFRm0uceO4C27lez3loHDmIG-zWSWoNjic" +
				"GNJ5qNya7lowSYOQOEgBcvVQeOrMz26B09SavrJ9rYAER-KgPXDxvOD7_7IYd" +
				"4ja2ThVG0RjKfvDLCgmg6nMIRl0ZW-Bn3FhwHo74NNF7KMF_QuzAvD2hnuesa" +
				"JlaIdAFyIk9WA",
			nil,
		},
	}
	for i, tt := range tokenTests {
		token, err := VerifyToken(tt.s, certs)
		if tt.t != nil && err != nil {
			t.Fatal(i, err)
		}
		if tt.t == nil && token != nil {
			t.Errorf("%d. VerifyToken(%q) = %v, want nil", i, tt.s, *token)
		}
		if tt.t != nil && *token != *tt.t {
			t.Errorf("%d. VerifyToken(%q) = %v; want %v", i, tt.s, *token, *tt.t)
		}
	}
}

func TestDecodeSegment(t *testing.T) {
	segTests := []struct {
		encoded string
		decoded []byte
	}{
		{"bm9wYWRkaW5n", []byte("nopadding")},
		{"cGFkZGluZzE", []byte("padding1")},
		{"cGFkZGluZzE=", []byte("padding1")},
		{"cGFkZGluZ3R3bw", []byte("paddingtwo")},
		{"cGFkZGluZ3R3bw==", []byte("paddingtwo")},
	}
	for i, st := range segTests {
		s, err := decodeSegment(st.encoded)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(s, st.decoded) {
			t.Errorf("%d. decodeSegment(%q) = %v; want %v", i, st.encoded, s, st.decoded)
		}
	}
}

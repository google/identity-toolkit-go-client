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
	"reflect"
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

// All valid tokens are signed with the following test private key:
//
// 	-----BEGIN RSA PRIVATE KEY-----
// 	MIIEowIBAAKCAQEAwusm+sw5BhWXO5r6Cf+bVjPjHi49+pZTqXBx813y0nwlpeqh
// 	YEiodN4HOX6uG03zDJftcy8Yli4bhUtCzHk9sfgi38H3MckCn3vr+TjsNa19nt7K
// 	XEjFY5oePfL5ErxFNwr/MQcOvn+xJi3mLBoTzgoToISdw8GNZpTq/6GgeHBXw5Rb
// 	gYNMk/Nil/Xau91aELcUwFIDnZan6ZhqynyuCjml1wcPP4eY5CTf7w0Jz2UZGpBM
//  MUaTqzcYY4ADH1wDdLdiIu7WzS6J7neYuSpXLDu6fffiGhIpQG0dOizefdJmP09a
// 	ufHebOuq5SEurHUBQfWiHGwR0gH+nRiP0NCtawIDAQABAoIBAQCow39JcJ8C1JjP
// 	g9xv4b72ARDpJbKGD7g6IWCYDUTb15+lKyQ0QFSSc1agAj3u6CvDQJvXufQ4u40a
// 	eNJKRcMUHzrar48+rMK3oUzryMxDoYX9kvp0spEtLth2XLQo85eeE8hmA61RZS/V
// 	VbJPhviU0Yay2X/baO7r3Y4mbSOKgnxBUUitlb1cf1rEL7HfEmwV9bJFJBRPnd0E
// 	V3/es6kqkgE9mR764ENASMcTfqFV4A3czB3xdm/zXotyVOWk6vzbAANwJUhB/lSq
// 	izEDCtosWtwBJAn/0kj0p4sRPhNNGQsWHeU1PkBnyWARKFw9nCtErF8gPSMObyWJ
// 	DCc2I+25AoGBAP9auxQD8MaBVW3hDx4UeC23UzosEx00RumS+HFgy1+kpO/Hx2JB
// 	54oVnuuaInLUjReelw4LS9GT2l3QJxOFJ8YnWgO/EdJYtruMYBe8ealveWW9RB1y
// 	Sugn0Djnd8FrBp2dhzjvd8TzW6p41EBOjCNPcbP3Io2nV1VEcRK0m0vNAoGBAMNp
// 	TnZ5fxxVgJfrBXp/Tdfb+nneq+bsLF2M/dKvRIPMmOmMIcHFYgXvPIyvqdTaj83H
// 	IJ70ezt2Xn+IgSpsz3BAXOuK+Iy4nfrojiav29XacMmOqgk9q2DYJjiHKbCo3OG5
// 	fqHussHkPnuWJLVsA/WiP0Pbs8jztdY41ecO/lYXAoGAGa1H7oVO6scKjhf5ay7n
// 	vsEWb10LNv6Qut/6d3wiNyXZJ/tQw0KZQ7tU1aAUIrPY1C0wR0s/fn19kb1ikIXt
// 	L/G5DXNjGWGuTkGZUAm1uiHpdR+I/+UEPz/xoic0VZCrg6HZfRInWviVmRi5TxYW
// 	5V4RsuUA5nmTQJG3RQhALWkCgYBonmWKmmuKMOCi9Qm+1o9HhRdpzeQ8wq7R+x0d
// 	ggZthclMKGEeS926DaZeL+H9+nHJGtYFJCEh6bKYFu4IV68Livn0hTeip2i1H3KT
// 	lc1dhhCiaIXyYpjxRWfZFcwlL5LqKpR4f8ImeltVcUH/+293P8EdRG/Ro2Dz6wPi
// 	K63RSwKBgHmFGsDNzmgcS/boYtrLoCrl3BfKn7vxZ1Y4kL77v+sp/fDrSajrGT4N
// 	EZIcKz+rQJUxBl5O32gsFOwPMH2OsN+ePnHV3K/EmYNy4H1bNjGe2cZdIcVkCppw
// 	S4NYN5F4lb5WaiPCjAiKSyfFhUty8Co+grsufOlRoTprqP/HAHRj
// 	-----END RSA PRIVATE KEY-----
//
// To generate more test tokens:
// 	header := "<ENCODED_HEADER>"
// 	claims := "<ENCODED_CLAIM_SET>"
// 	der, _ := pem.Decode([]byte(privateKeyPEM))
// 	key, _ := x509.ParsePKCS1PrivateKey(der.Bytes)
// 	hashed := sha256.Sum256([]byte(header + "." + claims))
// 	sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
// 	token := fmt.Sprintf("%s.%s.%s", header, claims, base64.RawURLEncoding.EncodeToString(sig))
func initCerts() *Certificates {
	block, _ := pem.Decode([]byte(`-----BEGIN CERTIFICATE-----
MIIDfzCCAmegAwIBAgIJAPHWUOaiAGAlMA0GCSqGSIb3DQEBCwUAMFUxCzAJBgNV
BAYTAlVTMQswCQYDVQQIDAJDQTEeMBwGA1UECgwVSWRlbnRpdHkgVG9vbGtpdCBU
ZXN0MRkwFwYDVQQDDBBJZGVudGl0eSBUb29sa2l0MCAXDTE1MTEwMTAwMTAzMFoY
DzIwNjUxMDE5MDAxMDMwWjBVMQswCQYDVQQGEwJVUzELMAkGA1UECAwCQ0ExHjAc
BgNVBAoMFUlkZW50aXR5IFRvb2xraXQgVGVzdDEZMBcGA1UEAwwQSWRlbnRpdHkg
VG9vbGtpdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMLrJvrMOQYV
lzua+gn/m1Yz4x4uPfqWU6lwcfNd8tJ8JaXqoWBIqHTeBzl+rhtN8wyX7XMvGJYu
G4VLQsx5PbH4It/B9zHJAp976/k47DWtfZ7eylxIxWOaHj3y+RK8RTcK/zEHDr5/
sSYt5iwaE84KE6CEncPBjWaU6v+hoHhwV8OUW4GDTJPzYpf12rvdWhC3FMBSA52W
p+mYasp8rgo5pdcHDz+HmOQk3+8NCc9lGRqQTDFGk6s3GGOAAx9cA3S3YiLu1s0u
ie53mLkqVyw7un334hoSKUBtHTos3n3SZj9PWrnx3mzrquUhLqx1AUH1ohxsEdIB
/p0Yj9DQrWsCAwEAAaNQME4wHQYDVR0OBBYEFLaLn7BXkWZg7KBjAv74UsDPKBru
MB8GA1UdIwQYMBaAFLaLn7BXkWZg7KBjAv74UsDPKBruMAwGA1UdEwQFMAMBAf8w
DQYJKoZIhvcNAQELBQADggEBAGKygI/uB63m64z/5HTryaZcDa3mToGaAXn+UoYH
scszQDWZEkDWTooFoXFQQ1JqYfJvO2VnER3AsmD/eCucTRgj62amWtvSwBr/trKN
3v1oPkYTet/G3rR6xKaLO1jrgXqYsZ15s1u8FYJJ7vpzMdLaj3apfeHsJYyJcx7a
NqpzTSsUZBc5VTEdDNGCG7JKG3HCfTzninUo1dLFPjeaQZCpFaUCzto2khRs8sBu
k39kTtZpzEegO5LxK/7iZ9StknZoo/1sK+YM4S+f0ZoT/6zdq8c5lOb/wmrj+5zH
ZKDo4vXBY/5H/NLE6TFS9wGfiWDCHyyqvxSkOGVFU+PO/wM=
-----END CERTIFICATE-----`))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return &Certificates{certs: map[string]*x509.Certificate{"qwYevA": cert}}
}

const (
	validToken            = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QSJ9.eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoyMDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.RHBOMH_hPemD3Dvqljf9zsrug04WBaaLfkMdPJqycuGEiXWkxnGakVWtUWV_HmfpwxG4YEi6sSlTioquqswogM_PzUx2s0gBHKAT8OuqPBdnrnaiuurZKAVfIuqEBRpKQyvvbkj7fzblZSgIkyeV6SMQ91ZENAizrywkUgqrHvj7up7sOjGPfqC0BQw4zvKsAORTB_NugEEavba16covSp8d80TpU2pL2_iFfcCP7j8vrMt6F-IIBa1BlfqK9dOIp4M-fKgshexD-GYiqZi09dFO8r-ANnfrM6sis5HsV2znO1ZjDZGKeQ_FUbqoiBhBNAd4wcCvBP26EehoQR1evQ"
	malformedToken        = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QSJ9eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoyMDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.RHBOMH_hPemD3Dvqljf9zsrug04WBaaLfkMdPJqycuGEiXWkxnGakVWtUWV_HmfpwxG4YEi6sSlTioquqswogM_PzUx2s0gBHKAT8OuqPBdnrnaiuurZKAVfIuqEBRpKQyvvbkj7fzblZSgIkyeV6SMQ91ZENAizrywkUgqrHvj7up7sOjGPfqC0BQw4zvKsAORTB_NugEEavba16covSp8d80TpU2pL2_iFfcCP7j8vrMt6F-IIBa1BlfqK9dOIp4M-fKgshexD-GYiqZi09dFO8r-ANnfrM6sis5HsV2znO1ZjDZGKeQ_FUbqoiBhBNAd4wcCvBP26EehoQR1evQ"
	invalidAlgorithmToken = "eyJhbGciOiJIUzI1NiIsImtpZCI6InF3WWV2QSJ9.eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoyMDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.W13HM9n9DZY_wDe9wuybP09JfyhAY7SYNFls5eT6X2zlq96loVPpmdqoyXrPhpMA066W8KRdXcBH3ExjzZzsfE1nyOWE82yQSsM2WFiEUq4bB3oyRQNCU-dojdoQHaHZlMSN5IpbSRQeGjDCnt_XYzZluz8MgdVARuVC4RO2xvJBGD8hXZIyALhyJjLqXzdWKwLLeWgr-mg1vLu3jnZpznGqn3K_siDMb4iGqG8rUqdLzUId5FPwQOY20Ia72-xmX-onfc_xDwO2cZcDyHkx_Yeyq68F6aFBfZ1yVboQdF2bCUojzj1mHovpRezf7x2BhCgXA9tWs07O21pIFlTv3g"
	invalidIssuerToken    = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QSJ9.eyJpc3MiOiJodHRwczovL2V4YW1wbGUuY29tIiwiYXVkIjoiMjE3OTIzMzkzNTczLmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tIiwiaWF0IjoxNDAwNDM3NzE1LCJleHAiOjIwMDE2NDczMTUsInVzZXJfaWQiOiIxNjEwOTg1Nzc2MDYwNzEwNjA4MCIsImVtYWlsIjoiZ2l0a2l0dGVzdEBnbWFpbC5jb20iLCJwcm92aWRlcl9pZCI6Imdvb2dsZS5jb20iLCJ2ZXJpZmllZCI6dHJ1ZSwiZGlzcGxheV9uYW1lIjoiSm9obiBEb2UiLCJwaG90b191cmwiOiJodHRwczovL2V4YW1wbGUuY29tL3BpY3R1cmUvMTYxMDk4NTc3NjA2MDcxMDYwODAifQ.oF3XPw9xrvUjL4LSHB9ucyGTDfvLZWFSxt9Iqu806eVSr3YIIwQ-1kCwI2VFJ7ur4q2ZpokmNDsl0gcKapV8KjzPtesQqwhxOMFO3X_NtOwDtDBsv7dLPUb5FNezSHA9cDNGg-LHLx_4ajYc33jwzQVlChvlctCrwYkCb-FMXJHMY3wWyHQV43B6zgm4FSkaQe1DuyJ6cGfGF8Im8U_ftXi3fvM8xAeAxUmhdUTwVXfbcj4J4gfVgU3RXWo4PlDnaTMthDE8s6rZC0JFeHtQ9eeM-5ax9s1zf64jEG3_XIH4SDkMw1liB7SFwOM_idmxbyjeZNtIdpqygpr868GqEQ"
	invalidAudienceToken  = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QSJ9.eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3NC5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoyMDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.HCLmHHk0rY8N1rb9duQb9cLxL8md0d11Dhd2TaQrHqitwiF_-Sw5uQsXkYa1YPkWvpECkB5NJsUSl1Naqlse68XcRFnPBNDUgV5NPJAVrrBxMII1eQxnotn3HOZ8F5QRjiFr6cihTSdw7J0-f1VFP0A8KXD0mDpiovykF2YFcdo3E6yOFwrohSJ1zMP4I1QJqYC7XLpq2TYg5e041I-KOFDNSu-nAKitH_fhJVhv8bwVAke_mnSqKlpW4oBA5zLsWkOUIp2wyKwe61rsoR8cXk-h60h1G6Fy23FdpcsC7TMI3HGuDDYrTD8E_l0HV8flW576Bz_lYrn6yHA7zwM9oA"
	keyNotFoundToken      = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QiJ9.eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoyMDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.ZhTMPJuTOjWjVKtHccGZ46EbK7uDivl5E_FgsaGAyMy6mrBFqnE6ugDvq6YbsHjhLfrcmwXhMNRaFJ1cT4gXf8H9LjOneEk_0wxckWdruW7ww9y3KRO8JP9S-CPpqEjNhIil3uhRmKizXBmsh5OM_FrqQV5B8d-vK_i8UXKYCW-Qol9N7uZb0j-CnYThezp2Do3dbAZ5LiXK3g6KTYBTarcUaKnMNf1kRZGoARXypI4qjqBJkOScOC95tVfVYRc9SDtMehJlW3OQBRbVZEYnjD8czPQCTWG9zHpcHIS6jvFiZafvLDNiTq_6L9NYe8kHeAP1bWaK6iZXMf4tV25QCQ"
	invalidSignatureToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QSJ9.eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoyMDAxNjQ3MzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.RHBOMH_hPemD3Dvqljf9zsrug04WBaaLfkMdPJqycuGEiXWkxnGakVWtUWV_HmfpwxG4YEi6sSlTioquqswogM_PzUx2s0gBHKAT8OuqPBdnrnaiuurZKAVfIuqEBRpKQyvvbkj7fzblZSgIkyeV6SMQ91ZENAizrywkUgqrHvj7up7sOjGPfqC0BQw4zvKsAORTB_NugEEavba16covSp8d80TpU2pL2_iFfcCP7j8vrMt6F-IIBa1BlfqK9dOIp4M-fKgshexD-GYiqZi09dFO8r-ANnfrM6sis5HsV2znO1ZjDZGKeQ_FUbqoiBhBNAd4wcCvBP26EehoQR1evq"
	expiredToken          = "eyJhbGciOiJSUzI1NiIsImtpZCI6InF3WWV2QSJ9.eyJpc3MiOiJodHRwczovL2lkZW50aXR5dG9vbGtpdC5nb29nbGUuY29tLyIsImF1ZCI6IjIxNzkyMzM5MzU3My5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImlhdCI6MTQwMDQzNzcxNSwiZXhwIjoxNDAwNDM4NzE1LCJ1c2VyX2lkIjoiMTYxMDk4NTc3NjA2MDcxMDYwODAiLCJlbWFpbCI6ImdpdGtpdHRlc3RAZ21haWwuY29tIiwicHJvdmlkZXJfaWQiOiJnb29nbGUuY29tIiwidmVyaWZpZWQiOnRydWUsImRpc3BsYXlfbmFtZSI6IkpvaG4gRG9lIiwicGhvdG9fdXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9waWN0dXJlLzE2MTA5ODU3NzYwNjA3MTA2MDgwIn0.M9g6YFuTy9CYoLyD8hhpTRE2B-EbQsIP-GCiju3JbbD-OqkURREyh85NikqysciI1N85slUV1Ov1g-8N3I6pN59fzHoD9Nw9i-PJ9r6oYdtIR3Y7iZ8yhJKHyWpDxMBez4pVwY9sUXjU-P0APNr7uZSpxwCnmEe_YZndHwLQmTqM3Sz-AIlCpRCyGef-Ff7Y885ZMbMmD2wCMnsq0pY9P6w2iBZHjYo7GLXdOjAdU8bxsq6LZE_EckOhI-0o95hTKKeWLQGTIVQmLyOScU9nxhzN77e3Hj0bv04AZkdFCseXtpHfFt0dsU6o8g-t3_bmARDFljs7XjjmGhTxjkieGA"

	issuer   = "https://identitytoolkit.google.com/"
	audience = "217923393573.apps.googleusercontent.com"
)

func TestVerifyToken(t *testing.T) {
	certs := initCerts()
	tokenTests := []struct {
		name  string
		s     string
		token *Token
		err   error
	}{
		{
			"valid",
			validToken,
			&Token{
				Issuer:        issuer,
				Audience:      audience,
				IssueAt:       time.Unix(1400437715, 0),
				ExpireAt:      time.Unix(2001647315, 0),
				LocalID:       "16109857760607106080",
				Email:         "gitkittest@gmail.com",
				EmailVerified: true,
				ProviderID:    "google.com",
				DisplayName:   "John Doe",
				PhotoURL:      "https://example.com/picture/16109857760607106080",
				TokenString:   validToken,
			},
			nil,
		},
		{
			"malformed",
			malformedToken,
			nil,
			ErrMalformed,
		},
		{
			"invalid alg",
			invalidAlgorithmToken, // "alg":"HS256"
			nil,
			ErrInvalidAlgorithm,
		},
		{
			"invalid issuer",
			invalidIssuerToken, // "iss":"https://example.com"
			nil,
			ErrInvalidIssuer,
		},
		{
			"invalid audience",
			invalidAudienceToken, // "aud":"217923393574.apps.googleusercontent.com"
			nil,
			ErrInvalidAudience,
		},
		{
			"key not found",
			keyNotFoundToken, // "kid":"qwYevB"
			nil,
			ErrKeyNotFound,
		},
		{
			"invalid signature",
			invalidSignatureToken, // Same as validToken except for the last character of signature changed from "Q" to "q".
			nil,
			ErrInvalidSignature,
		},
		{
			"expired",
			expiredToken, // "exp":1400438715. Expired at Sun May 18 11:28:35 PDT 2014
			nil,
			ErrExpired,
		},
	}
	for i, tt := range tokenTests {
		token, err := VerifyToken(tt.s, []string{audience}, []string{issuer}, certs)
		if tt.err != err {
			t.Errorf("[%d]%s: expected error=%v, but got %v", i, tt.name, tt.err, err)
		}
		if !reflect.DeepEqual(tt.token, token) {
			t.Errorf("[%d]%s: expected token=%v, but got %v", i, tt.name, tt.token, token)
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
			t.Errorf("%d. expected decodeSegment(%q) = %v, but got %v", i, st.encoded, st.decoded, s)
		}
	}
}

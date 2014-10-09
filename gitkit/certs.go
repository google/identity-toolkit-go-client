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
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Certificates contains a set of availabe identitytoolkit public certificates
// which are indexed by key IDs ("kid"). It caches the certificates according
// to the HTTP response cache setting and refetches them upon cache expiring.
// It is safe to use a Certificates from multiple concurrent goroutines.
type Certificates struct {
	certs map[string]*x509.Certificate
	mu    sync.RWMutex // Lock for updating the map
	t     *time.Timer  // Timer that triggers the next certificates update.
	url   string       // Certificates URL.
	err   error        // Last error for updating the certificates.
}

// LoadCerts downloads the certificates from the given URL.
func LoadCerts(url string, transport http.RoundTripper) (*Certificates, error) {
	c := Certificates{url: url}
	c.update(transport)
	if c.err != nil {
		return nil, c.err
	}
	return &c, nil
}

// Cert returns the public certificate for the given key ID.
func (c *Certificates) Cert(keyID string) (*x509.Certificate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cert, found := c.certs[keyID]
	if found {
		return cert, nil
	} else if c.err == nil {
		return nil, fmt.Errorf("certificate not found for: %s", c.err)
	} else {
		return nil, c.err
	}
}

const retryInterval = 30 * time.Second

// update fetches the certificates and starts a time for the next update.
func (c *Certificates) update(transport http.RoundTripper) {
	certs, cacheTime, err := downloadCerts(c.url, transport)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
	// In case of failure, save the error and retry the certificates fetching
	// after retryInterval.
	if err != nil {
		cacheTime = retryInterval
	} else {
		c.certs = certs
	}
	if c.t != nil {
		c.t.Stop()
	}
	c.t = time.AfterFunc(cacheTime, func() { c.update(transport) })
}

// downloadCerts downloads and parses the certificates from the given URL.
func downloadCerts(url string, transport http.RoundTripper) (map[string]*x509.Certificate, time.Duration, error) {
	client := http.Client{Transport: transport}
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("get %s: %s", url, resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	certs, err := parseCerts(b)
	if err != nil {
		return nil, 0, err
	}
	return certs, cacheTime(resp), nil
}

// parseCerts parses the JSON encoding certificates response.
// The response has the following format:
//	{
//	  "40QoZg": "-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----",
//	  "5DaIwB": "-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----"
//	}
func parseCerts(resp []byte) (map[string]*x509.Certificate, error) {
	m := make(map[string]string)
	if err := json.Unmarshal(resp, &m); err != nil {
		return nil, err
	}
	certs := make(map[string]*x509.Certificate)
	for k, v := range m {
		block, _ := pem.Decode([]byte(v))
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs[k] = c
	}
	return certs, nil
}

const defaultCertsCacheTime = 1 * time.Hour

// cacheTime extracts the cache time from the HTTP response header.
// The certificates endpoint should always return a valid max-age in the HTTP
// response header. If it fails to do so, cache the certificates for a default
// time to avoid performance regression and DDOS for the certificates endpoint.
func cacheTime(resp *http.Response) time.Duration {
	ss := strings.Split(resp.Header.Get("Cache-Control"), ",")
	const prefix = "max-age="
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, prefix) {
			if d, err := strconv.Atoi(s[len(prefix):]); err == nil {
				return time.Duration(d) * time.Second
			}
		}
	}
	return defaultCertsCacheTime
}

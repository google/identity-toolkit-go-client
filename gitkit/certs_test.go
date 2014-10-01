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
	"net/http"
	"testing"
	"time"
)

func TestParseCerts(t *testing.T) {
	tests := []struct {
		resp []byte
		err  bool
	}{
		{
			// Valid certificates.
			[]byte("{\"40QoZg\": \"-----BEGIN CERTIFICATE-----\\n" +
				"MIIDDzCCAfegAwIBAgIJAMKLYPybcIAZMA0GCSqGSIb3DQEBBQUAMB4xHDAaBgNV\\n" +
				"BAMME0dvb2dsZSBBdXRoIFRvb2xraXQwHhcNMTMwNDI1MTUyMDExWhcNMTQwNDI1\\n" +
				"MTUyMDExWjAeMRwwGgYDVQQDDBNHb29nbGUgQXV0aCBUb29sa2l0MIIBIjANBgkq\\n" +
				"hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyVZ3j4Uovsspa6dCiTZAC/SndulGDKYf\\n" +
				"mVr95ea+u4k0XMvvd7w9k0wq4d1xagMIKHZhAnYLvYfW0O5D8+d58/+UJq4vrlY9\\n" +
				"zOcTOsOoZ5tX325TMIJmn7IzMMpds1tA2MfWNiMkf/+AFZfxg14jyBeRdk4LVZWa\\n" +
				"FxMz9Fs/23pTuNBYwGzM3xyZajgEhJ9gp3k95qlQPq00bIMa69YiAcmyr4RVYpgW\\n" +
				"qd+WPdROEZvRLsCaIGTeehLR6zceUPrTofbOo82JI3/PTfJ+bm+IzXRq5Ogynfw6\\n" +
				"f4z0pJ/YuUlmGD+rrm5Dfja/V3QTPyqzFpQSPXND7OdpT63MryKHtQIDAQABo1Aw\\n" +
				"TjAdBgNVHQ4EFgQUwGCN266hsEwDjx2aNQ4cdPSjmJMwHwYDVR0jBBgwFoAUwGCN\\n" +
				"266hsEwDjx2aNQ4cdPSjmJMwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOC\\n" +
				"AQEABDl3G5Ao3ZTXdeNoeF8knWl//6pyxz/Jhv1/PApA9NQpyhqijmGyDMvCLt0F\\n" +
				"02HVTqg/MYG5zwUCroV9daraEdn5302sx8kh1Ei8SBCKzoDa7B8wSd2/KrEd6zsX\\n" +
				"/7ZVzSNx37xk5Jhzz6EmXfY7z22DmFWggxyeTYGgR5YgKkuslbIxxEKjVhK5YK60\\n" +
				"1pyRhl0tqe2xt+FMn0tvLdkCfVCvyDj2cD7g5XBVXZS4rqwfy1XpzQfSuU4sQcgn\\n" +
				"VpgjVOtnax48yJFXeNTrOoTPiQV2AZQSrGuKoJ8GojM6oZuEv5S2moB3IMKyU5F3\\n" +
				"RQ1NcLfJHhAz2ccdbaBXJaP4Hw==\\n" +
				"-----END CERTIFICATE-----\"}"),
			false,
		},
		{
			// Invalid certificates (The begining 2 bytes are changed).
			[]byte("{\"40QoZg\": \"-----BEGIN CERTIFICATE-----\\n" +
				"DEIDDzCCAfegAwIBAgIJAMKLYPybcIAZMA0GCSqGSIb3DQEBBQUAMB4xHDAaBgNV\\n" +
				"BAMME0dvb2dsZSBBdXRoIFRvb2xraXQwHhcNMTMwNDI1MTUyMDExWhcNMTQwNDI1\\n" +
				"MTUyMDExWjAeMRwwGgYDVQQDDBNHb29nbGUgQXV0aCBUb29sa2l0MIIBIjANBgkq\\n" +
				"hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyVZ3j4Uovsspa6dCiTZAC/SndulGDKYf\\n" +
				"mVr95ea+u4k0XMvvd7w9k0wq4d1xagMIKHZhAnYLvYfW0O5D8+d58/+UJq4vrlY9\\n" +
				"zOcTOsOoZ5tX325TMIJmn7IzMMpds1tA2MfWNiMkf/+AFZfxg14jyBeRdk4LVZWa\\n" +
				"FxMz9Fs/23pTuNBYwGzM3xyZajgEhJ9gp3k95qlQPq00bIMa69YiAcmyr4RVYpgW\\n" +
				"qd+WPdROEZvRLsCaIGTeehLR6zceUPrTofbOo82JI3/PTfJ+bm+IzXRq5Ogynfw6\\n" +
				"f4z0pJ/YuUlmGD+rrm5Dfja/V3QTPyqzFpQSPXND7OdpT63MryKHtQIDAQABo1Aw\\n" +
				"TjAdBgNVHQ4EFgQUwGCN266hsEwDjx2aNQ4cdPSjmJMwHwYDVR0jBBgwFoAUwGCN\\n" +
				"266hsEwDjx2aNQ4cdPSjmJMwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOC\\n" +
				"AQEABDl3G5Ao3ZTXdeNoeF8knWl//6pyxz/Jhv1/PApA9NQpyhqijmGyDMvCLt0F\\n" +
				"02HVTqg/MYG5zwUCroV9daraEdn5302sx8kh1Ei8SBCKzoDa7B8wSd2/KrEd6zsX\\n" +
				"/7ZVzSNx37xk5Jhzz6EmXfY7z22DmFWggxyeTYGgR5YgKkuslbIxxEKjVhK5YK60\\n" +
				"1pyRhl0tqe2xt+FMn0tvLdkCfVCvyDj2cD7g5XBVXZS4rqwfy1XpzQfSuU4sQcgn\\n" +
				"VpgjVOtnax48yJFXeNTrOoTPiQV2AZQSrGuKoJ8GojM6oZuEv5S2moB3IMKyU5F3\\n" +
				"RQ1NcLfJHhAz2ccdbaBXJaP4Hw==\\n" +
				"-----END CERTIFICATE-----\"}"),
			true,
		},
		{
			// Invalid JSON.
			[]byte("{\"40QoZg: \"-----BEGIN CERTIFICATE-----\\n" +
				"MIIDDzCCAfegAwIBAgIJAMKLYPybcIAZMA0GCSqGSIb3DQEBBQUAMB4xHDAaBgNV\\n" +
				"BAMME0dvb2dsZSBBdXRoIFRvb2xraXQwHhcNMTMwNDI1MTUyMDExWhcNMTQwNDI1\\n" +
				"MTUyMDExWjAeMRwwGgYDVQQDDBNHb29nbGUgQXV0aCBUb29sa2l0MIIBIjANBgkq\\n" +
				"hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyVZ3j4Uovsspa6dCiTZAC/SndulGDKYf\\n" +
				"mVr95ea+u4k0XMvvd7w9k0wq4d1xagMIKHZhAnYLvYfW0O5D8+d58/+UJq4vrlY9\\n" +
				"zOcTOsOoZ5tX325TMIJmn7IzMMpds1tA2MfWNiMkf/+AFZfxg14jyBeRdk4LVZWa\\n" +
				"FxMz9Fs/23pTuNBYwGzM3xyZajgEhJ9gp3k95qlQPq00bIMa69YiAcmyr4RVYpgW\\n" +
				"qd+WPdROEZvRLsCaIGTeehLR6zceUPrTofbOo82JI3/PTfJ+bm+IzXRq5Ogynfw6\\n" +
				"f4z0pJ/YuUlmGD+rrm5Dfja/V3QTPyqzFpQSPXND7OdpT63MryKHtQIDAQABo1Aw\\n" +
				"TjAdBgNVHQ4EFgQUwGCN266hsEwDjx2aNQ4cdPSjmJMwHwYDVR0jBBgwFoAUwGCN\\n" +
				"266hsEwDjx2aNQ4cdPSjmJMwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOC\\n" +
				"AQEABDl3G5Ao3ZTXdeNoeF8knWl//6pyxz/Jhv1/PApA9NQpyhqijmGyDMvCLt0F\\n" +
				"02HVTqg/MYG5zwUCroV9daraEdn5302sx8kh1Ei8SBCKzoDa7B8wSd2/KrEd6zsX\\n" +
				"/7ZVzSNx37xk5Jhzz6EmXfY7z22DmFWggxyeTYGgR5YgKkuslbIxxEKjVhK5YK60\\n" +
				"1pyRhl0tqe2xt+FMn0tvLdkCfVCvyDj2cD7g5XBVXZS4rqwfy1XpzQfSuU4sQcgn\\n" +
				"VpgjVOtnax48yJFXeNTrOoTPiQV2AZQSrGuKoJ8GojM6oZuEv5S2moB3IMKyU5F3\\n" +
				"RQ1NcLfJHhAz2ccdbaBXJaP4Hw==\\n" +
				"-----END CERTIFICATE-----\"}"),
			true,
		},
	}
	for i, tt := range tests {
		c, err := parseCerts(tt.resp)
		if !tt.err {
			if err != nil {
				t.Errorf("%d parseCert() returns error: %v; want nil", i, err)
			} else if c["40QoZg"] == nil {
				t.Errorf("%d cannot find certificate for kid: 40QoZg", i)
			}
		} else {
			if err == nil {
				t.Errorf("%d parseCert() returns nil error; want non nil", i)
			}
		}
	}
}

func TestCacheTime(t *testing.T) {
	tests := []struct {
		cache string
		d     time.Duration
	}{
		{"public, max-age=7200", 7200 * time.Second},
		{"public, max-age=invalid value", defaultCertsCacheTime},
		{"public", defaultCertsCacheTime},
		{"", defaultCertsCacheTime},
	}
	for i, tt := range tests {
		r := &http.Response{Header: make(http.Header)}
		if tt.cache != "" {
			r.Header.Set("cache-control", tt.cache)
		}
		if d := cacheTime(r); d != tt.d {
			t.Errorf("%d cache setting = %q. cacheTime() = %v, want %v", i, tt.cache, d, tt.d)
		}
	}
}

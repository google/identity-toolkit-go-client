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
	"encoding/json"
	"errors"
	"io/ioutil"
)

// Config contains the configurations for creating a Client.
type Config struct {
	// ClientID is the Google OAuth2 client ID for the server.
	ClientID string `json:"clientId,omitempty"`
	// WidgetURL is the identitytoolkit javascript widget URL.
	// It is used to generate the reset password or change email URL and
	// could be an absolute URL, an absolute path or a relative path.
	WidgetURL string `json:"widgetUrl,omitempty"`
	// WidgetModeParamName is the parameter name used by the javascript widget.
	// A default value is used if left unspecified. If the parameter name is set
	// to other value in the javascript widget, this field should be set to the
	// same value.
	WidgetModeParamName string `json:"widgetModeParamName,omitempty"`
	// CookieName is the name of the cookie that stores the ID token.
	CookieName string `json:"cookieName,omitempty"`
}

// LoadConfig loads the configuration from the config file specified by path.
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

const (
	DefaultWidgetModeParamName = "mode"
	DefaultCookieName          = "gtoken"
)

func (conf *Config) normalize() error {
	if conf.ClientID == "" {
		return errors.New("missing ClientID in config")
	}
	if conf.WidgetModeParamName == "" {
		conf.WidgetModeParamName = DefaultWidgetModeParamName
	}
	if conf.CookieName == "" {
		conf.CookieName = DefaultCookieName
	}
	return nil
}

package gitkit

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

const (
	config = `{
		"browserApiKey": "browser_key",
		"clientId": "client_id",
		"widgetUrl": "widget_url",
		"widgetModeParamName": "widget_mode_param_name",
		"cookieName": "cookie_name",
		"signInOptions": ["option1", "option2"],
		"googleAppCredentialsPath": "/some/path"
	}`
	configWithUnrecognized = `{
		"browserApiKey": "browser_key",
		"clientId": "client_id",
		"widgetUrl": "widget_url",
		"widgetModeParamName": "widget_mode_param_name",
		"cookieName": "cookie_name",
		"signInOptions": ["option1", "option2"],
		"googleAppCredentialsPath": "/some/path",
		"unrecognized": "blabla"
	}`
)

func TestLoadConfig_notFound(t *testing.T) {
	_, err := LoadConfig("/some/path/not/exist")
	if err == nil {
		t.Errorf("expected error for loading non exist config file, but got nil")
	}
}

func TestLoadConfig_notJSON(t *testing.T) {
	f, err := createConfigFile("not a JSON file")
	if err != nil {
		t.Errorf("cannot create temp config file")
	}
	defer os.Remove(f)
	_, err = LoadConfig(f)
	if err == nil {
		t.Errorf("expected error for loading non exist config file, but got nil")
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		config string
	}{{config}, {configWithUnrecognized}}
	conf := Config{"browser_key", "client_id", "widget_url", "widget_mode_param_name", "cookie_name", []string{"option1", "option2"}, "/some/path"}
	for i, tt := range tests {
		f, err := createConfigFile(tt.config)
		if err != nil {
			t.Errorf("[%d]: cannot create temp config file", i)
		}
		defer os.Remove(f)
		c, err := LoadConfig(f)
		if err != nil {
			t.Errorf("[%d]: expected no error for LoadConfig(), but got [%v]", i, err)
		} else if !reflect.DeepEqual(*c, conf) {
			t.Errorf("[%d]: expected LoadConfig()=%v, but got %v", i, conf, c)
		}
	}
}

func TestConfig_normalize(t *testing.T) {
	tests := []struct {
		orig       *Config
		normalized *Config
	}{
		{
			&Config{"", "", "/", "mode", "gtoken", nil, ""},
			nil,
		},
		{
			&Config{"", "client_id", "/", "", "", nil, ""},
			&Config{"", "client_id", "/", "mode", "gtoken", nil, ""},
		},
		{
			&Config{"", "client_id", "/", "mode", "gtoken", nil, "/some/path"},
			&Config{"", "client_id", "/", "mode", "gtoken", nil, "/some/path"},
		},
		{
			&Config{"", "client_id", "/", "", "gitkittoken", nil, ""},
			&Config{"", "client_id", "/", "mode", "gitkittoken", nil, ""},
		},
		{
			&Config{"", "client_id", "/", "gitkitmode", "", nil, ""},
			&Config{"", "client_id", "/", "gitkitmode", "gtoken", nil, ""},
		},
		{
			&Config{"", "client_id", "/", "gitkitmode", "gitkittoken", nil, ""},
			&Config{"", "client_id", "/", "gitkitmode", "gitkittoken", nil, ""},
		},
	}
	for i, tt := range tests {
		err := tt.orig.normalize()
		if tt.normalized == nil && err == nil {
			t.Errorf("[%d]: expected normalize() to return error, but got nil", i)
		}
		if tt.normalized != nil && !reflect.DeepEqual(*tt.normalized, *tt.orig) {
			t.Errorf("[%d]: expected normalize()=%v, but got %v", i, *tt.normalized, *tt.orig)
		}
	}
}

func createConfigFile(config string) (string, error) {
	f, err := ioutil.TempFile("", "testconf")
	if err != nil {
		return "", err
	}
	defer f.Close()
	f.WriteString(config)
	return f.Name(), nil
}

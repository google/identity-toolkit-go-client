package gitkit

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

const (
	config = `{
		"clientId": "client_id",
		"widgetUrl": "widget_url",
		"widgetModeParamName": "widget_mode_param_name",
		"cookieName": "cookie_name",
		"googleAppCredentialsPath": "/some/path"
	}`
	configWithUnrecognized = `{
		"clientId": "client_id",
		"widgetUrl": "widget_url",
		"widgetModeParamName": "widget_mode_param_name",
		"cookieName": "cookie_name",
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
	conf := Config{"client_id", "widget_url", "widget_mode_param_name", "cookie_name", "/some/path"}
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
			&Config{"", "/", "mode", "gtoken", ""},
			&Config{"", "/", "mode", "gtoken", ""},
		},
		{
			&Config{"client_id", "/", "", "", ""},
			&Config{"client_id", "/", "mode", "gtoken", ""},
		},
		{
			&Config{"client_id", "/", "mode", "gtoken", "/some/path"},
			&Config{"client_id", "/", "mode", "gtoken", "/some/path"},
		},
		{
			&Config{"client_id", "/", "", "gitkittoken", ""},
			&Config{"client_id", "/", "mode", "gitkittoken", ""},
		},
		{
			&Config{"client_id", "/", "gitkitmode", "", ""},
			&Config{"client_id", "/", "gitkitmode", "gtoken", ""},
		},
		{
			&Config{"client_id", "/", "gitkitmode", "gitkittoken", ""},
			&Config{"client_id", "/", "gitkitmode", "gitkittoken", ""},
		},
	}
	for i, tt := range tests {
		tt.orig.normalize()
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

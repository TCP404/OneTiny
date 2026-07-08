package conf

import (
	"reflect"
	"testing"
)

func TestConfigDoesNotContainRuntimeOnlyFields(t *testing.T) {
	cfgType := reflect.TypeOf(Config{})
	for _, name := range []string{"Output", "OS", "IP", "Pwd", "SessionVal"} {
		if _, ok := cfgType.FieldByName(name); ok {
			t.Fatalf("Config contains runtime-only field %s", name)
		}
	}
}

func TestCurrentReturnsCopy(t *testing.T) {
	resetViper(t)
	writeUserConfig(t, `
server:
  port: 8192
  allow_upload: false
  max_level: 0
`)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	cfg.Port = 12345

	if got := Current().Port; got == cfg.Port {
		t.Fatalf("Current() returned mutable trusted config, port = %d", got)
	}
}

func TestSavePatchPersistsAndReloadsTrustedConfig(t *testing.T) {
	resetViper(t)
	writeUserConfig(t, `
server:
  port: 8192
  allow_upload: false
  max_level: 0
`)
	if _, err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	port := 10080
	allow := true
	maxLevel := uint8(3)
	got, err := SavePatch(ConfigPatch{
		Port:          &port,
		IsAllowUpload: &allow,
		MaxLevel:      &maxLevel,
	})
	if err != nil {
		t.Fatalf("SavePatch returned error: %v", err)
	}

	if got.Port != port {
		t.Fatalf("saved port = %d, want %d", got.Port, port)
	}
	if got.IsAllowUpload != allow {
		t.Fatalf("saved allow upload = %t, want %t", got.IsAllowUpload, allow)
	}
	if got.MaxLevel != maxLevel {
		t.Fatalf("saved max level = %d, want %d", got.MaxLevel, maxLevel)
	}

	reloaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("reload returned error: %v", err)
	}
	if reloaded.Port != port || reloaded.IsAllowUpload != allow || reloaded.MaxLevel != maxLevel {
		t.Fatalf("reloaded config = %+v, want persisted patch", reloaded)
	}
}

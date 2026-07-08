package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const validBcryptHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi47OhMKn8IW0YFDCw5Ac1TYT2RP1xG"

func newTestStore(t *testing.T, content string) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile config.yml: %v", err)
		}
	}
	return NewStore(path)
}

func TestConfigDoesNotContainRuntimeOnlyFields(t *testing.T) {
	cfgType := reflect.TypeOf(Config{})
	for _, name := range []string{"Output", "OS", "IP", "Pwd", "SessionVal"} {
		if _, ok := cfgType.FieldByName(name); ok {
			t.Fatalf("Config contains runtime-only field %s", name)
		}
	}
}

func TestStoreLoadCreatesDefaults(t *testing.T) {
	store := newTestStore(t, "")
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Port == 0 {
		t.Fatal("Load returned zero port")
	}
	if cfg.RootPath == "" {
		t.Fatal("Load returned empty root path")
	}
	if _, err := os.Stat(store.Path()); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
}

func TestStoreCurrentReturnsCopy(t *testing.T) {
	store := newTestStore(t, `
server:
  port: 8192
  allow_upload: false
  max_level: 0
`)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	cfg.Port = 12345

	if got := store.Current().Port; got == cfg.Port {
		t.Fatalf("Current returned mutable trusted config, port = %d", got)
	}
}

func TestStorePatchPersistsAndReloadsTrustedConfig(t *testing.T) {
	store := newTestStore(t, `
server:
  port: 8192
  allow_upload: false
  max_level: 0
`)
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	port := 10080
	allow := true
	maxLevel := uint8(3)
	got, err := store.Patch(ConfigPatch{
		Port:          &port,
		IsAllowUpload: &allow,
		MaxLevel:      &maxLevel,
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
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

	reloaded, err := store.Load()
	if err != nil {
		t.Fatalf("reload returned error: %v", err)
	}
	if reloaded.Port != port || reloaded.IsAllowUpload != allow || reloaded.MaxLevel != maxLevel {
		t.Fatalf("reloaded config = %+v, want persisted patch", reloaded)
	}
}

func TestStoreLoadReadsCredentialFieldsWithoutSecureValidation(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		wantUsername string
		wantPassword string
	}{
		{
			name: "secure bcrypt",
			config: `
account:
  secure: true
  custom:
    user: admin
    pass_hash: ` + validBcryptHash + `
    pass_hash_algo: bcrypt
`,
			wantUsername: "admin",
			wantPassword: validBcryptHash,
		},
		{
			name: "secure legacy md5",
			config: `
account:
  secure: true
  custom:
    user: 21232f297a57a5a743894a0e4a801fc3
    pass: 21232f297a57a5a743894a0e4a801fc3
`,
			wantUsername: "21232f297a57a5a743894a0e4a801fc3",
		},
		{
			name: "unprotected legacy md5",
			config: `
account:
  secure: false
  custom:
    user: 21232f297a57a5a743894a0e4a801fc3
    pass: 21232f297a57a5a743894a0e4a801fc3
`,
			wantUsername: "21232f297a57a5a743894a0e4a801fc3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t, tt.config)

			cfg, err := store.Load()
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			if cfg.Username != tt.wantUsername {
				t.Fatalf("Config.Username = %q, want %q", cfg.Username, tt.wantUsername)
			}
			if cfg.PasswordHash != tt.wantPassword {
				t.Fatalf("Config.PasswordHash = %q, want %q", cfg.PasswordHash, tt.wantPassword)
			}
		})
	}
}

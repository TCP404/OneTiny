package config

import (
	"fmt"
	"math"
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
	cfgType := reflect.TypeFor[Config]()
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

func TestStoreLoadCreatesScratchDefaults(t *testing.T) {
	store := newTestStore(t, "")
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.ScratchMaxItems != 500 {
		t.Fatalf("ScratchMaxItems = %d, want 500", cfg.ScratchMaxItems)
	}
	if cfg.ScratchMaxItemSize != "10MB" {
		t.Fatalf("ScratchMaxItemSize = %q, want 10MB", cfg.ScratchMaxItemSize)
	}
}

func TestStoreLoadBackfillsScratchDefaultsForOldConfig(t *testing.T) {
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
	if cfg.ScratchMaxItems != 500 || cfg.ScratchMaxItemSize != "10MB" {
		t.Fatalf("scratch defaults = %d %q, want 500 10MB", cfg.ScratchMaxItems, cfg.ScratchMaxItemSize)
	}
}

func TestStorePatchPersistsScratchConfig(t *testing.T) {
	store := newTestStore(t, `
server:
  port: 8192
scratch:
  max_items: 500
  max_item_size: 10MB
`)
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	maxItems := 42
	maxSize := "512KB"
	got, err := store.Patch(ConfigPatch{
		ScratchMaxItems:    &maxItems,
		ScratchMaxItemSize: &maxSize,
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}
	if got.ScratchMaxItems != 42 || got.ScratchMaxItemSize != "512KB" {
		t.Fatalf("patched scratch = %d %q, want 42 512KB", got.ScratchMaxItems, got.ScratchMaxItemSize)
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if reloaded.ScratchMaxItems != 42 || reloaded.ScratchMaxItemSize != "512KB" {
		t.Fatalf("reloaded scratch = %d %q, want 42 512KB", reloaded.ScratchMaxItems, reloaded.ScratchMaxItemSize)
	}
}

func TestStorePatchRejectsInvalidScratchConfigWithoutMutatingTrustedConfig(t *testing.T) {
	tests := []struct {
		name  string
		patch ConfigPatch
	}{
		{
			name: "max items",
			patch: func() ConfigPatch {
				maxItems := 0
				return ConfigPatch{ScratchMaxItems: &maxItems}
			}(),
		},
		{
			name: "max item size",
			patch: func() ConfigPatch {
				maxSize := "nope"
				return ConfigPatch{ScratchMaxItemSize: &maxSize}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t, `
server:
  port: 8192
scratch:
  max_items: 500
  max_item_size: 10MB
`)
			loaded, err := store.Load()
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			beforeFile, err := os.ReadFile(store.Path())
			if err != nil {
				t.Fatalf("ReadFile before patch: %v", err)
			}

			if _, err := store.Patch(tt.patch); err == nil {
				t.Fatal("Patch accepted invalid scratch config")
			}

			if got := store.Current(); !reflect.DeepEqual(got, loaded) {
				t.Fatalf("Current after failed patch = %+v, want %+v", got, loaded)
			}

			afterFile, err := os.ReadFile(store.Path())
			if err != nil {
				t.Fatalf("ReadFile after patch: %v", err)
			}
			if string(afterFile) != string(beforeFile) {
				t.Fatalf("config file changed after failed patch:\nbefore:\n%s\nafter:\n%s", beforeFile, afterFile)
			}

			reloaded, err := store.Load()
			if err != nil {
				t.Fatalf("Load returned error after failed patch: %v", err)
			}
			if !reflect.DeepEqual(reloaded, loaded) {
				t.Fatalf("Reload after failed patch = %+v, want %+v", reloaded, loaded)
			}
		})
	}
}

func TestStoreRejectsInvalidScratchConfig(t *testing.T) {
	store := newTestStore(t, `
scratch:
  max_items: 0
  max_item_size: 10MB
`)
	if _, err := store.Load(); err == nil {
		t.Fatal("Load accepted invalid scratch config")
	}
}

func TestStoreRejectsInvalidScratchSize(t *testing.T) {
	store := newTestStore(t, `
scratch:
  max_items: 1
  max_item_size: nope
`)
	if _, err := store.Load(); err == nil {
		t.Fatal("Load accepted invalid scratch size")
	}
}

func TestParseByteSize(t *testing.T) {
	tests := map[string]int64{
		"1B":    1,
		"10KB":  10 * 1024,
		"10MB":  10 * 1024 * 1024,
		"2GB":   2 * 1024 * 1024 * 1024,
		"4096":  4096,
		" 5mb ": 5 * 1024 * 1024,
	}
	for input, want := range tests {
		got, err := ParseByteSize(input)
		if err != nil {
			t.Fatalf("ParseByteSize(%q) error = %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseByteSize(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestParseByteSizeRejectsInvalidInput(t *testing.T) {
	tests := []string{
		"",
		"nope",
		fmt.Sprintf("%dKB", math.MaxInt64/1024+1),
		fmt.Sprintf("%dMB", math.MaxInt64/(1024*1024)+1),
		fmt.Sprintf("%dGB", math.MaxInt64/(1024*1024*1024)+1),
	}
	for _, input := range tests {
		if _, err := ParseByteSize(input); err == nil {
			t.Fatalf("ParseByteSize(%q) accepted invalid input", input)
		}
	}
}

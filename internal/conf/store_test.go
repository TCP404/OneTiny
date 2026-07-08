package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/security"
)

const validBcryptHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi47OhMKn8IW0YFDCw5Ac1TYT2RP1xG"

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func resetConfig(t *testing.T) {
	t.Helper()
	original := currentConfig
	t.Cleanup(func() {
		currentConfig = original
	})
}

func writeUserConfig(t *testing.T, content string) {
	t.Helper()
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("APPDATA", configRoot)
	t.Setenv("HOME", configRoot)
	t.Setenv("USERPROFILE", configRoot)

	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir returned error: %v", err)
	}
	rel, err := filepath.Rel(configRoot, userCfgDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("os.UserConfigDir returned %q outside isolated root %q", userCfgDir, configRoot)
	}
	cfgDir := filepath.Join(userCfgDir, "tiny")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("MkdirAll config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile config.yml: %v", err)
	}
}

func TestCredentialConfigFromViperReadsBcryptFields(t *testing.T) {
	resetViper(t)
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", "$2a$10$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("account.custom.pass_hash_algo", "bcrypt")

	creds := CredentialConfigFromViper()
	if creds.Username != "admin" {
		t.Fatalf("username = %q, want admin", creds.Username)
	}
	if creds.PasswordHash == "" {
		t.Fatal("missing password hash")
	}
	if creds.HashAlgo != "bcrypt" {
		t.Fatalf("hash algo = %q, want bcrypt", creds.HashAlgo)
	}
}

func TestValidateSecureConfigRejectsLegacyMD5(t *testing.T) {
	resetViper(t)
	viper.Set("account.secure", true)
	viper.Set("account.custom.user", "21232f297a57a5a743894a0e4a801fc3")
	viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")

	if err := ValidateSecureConfig(); err == nil {
		t.Fatal("ValidateSecureConfig accepted legacy MD5 secure config")
	}
}

func TestValidateSecureConfigAllowsUnprotectedLegacyConfig(t *testing.T) {
	resetViper(t)
	viper.Set("account.secure", false)
	viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")

	if err := ValidateSecureConfig(); err != nil {
		t.Fatalf("unprotected legacy config rejected: %v", err)
	}
}

func TestValidateSecureConfigForRejectsLegacyWhenEffectiveSecure(t *testing.T) {
	resetViper(t)
	viper.Set("account.secure", false)
	viper.Set("account.custom.user", "21232f297a57a5a743894a0e4a801fc3")
	viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")

	if err := ValidateSecureConfigFor(true); err == nil {
		t.Fatal("ValidateSecureConfigFor accepted legacy MD5 with effective secure mode")
	}
}

func TestSetCredentialConfigWritesBcryptAndClearsLegacy(t *testing.T) {
	resetViper(t)
	viper.Set("account.secure", true)
	viper.Set("account.custom.pass", "legacy-md5")

	SetCredentialConfig("admin", validBcryptHash)

	if got := viper.GetString("account.custom.user"); got != "admin" {
		t.Fatalf("account.custom.user = %q, want admin", got)
	}
	if got := viper.GetString("account.custom.pass_hash"); got != validBcryptHash {
		t.Fatalf("account.custom.pass_hash = %q, want %q", got, validBcryptHash)
	}
	if got := viper.GetString("account.custom.pass_hash_algo"); got != security.HashAlgoBcrypt {
		t.Fatalf("account.custom.pass_hash_algo = %q, want %q", got, security.HashAlgoBcrypt)
	}
	if got := viper.GetString("account.custom.pass"); got != "" {
		t.Fatalf("account.custom.pass = %q, want empty legacy password", got)
	}
	if err := ValidateSecureConfig(); err != nil {
		t.Fatalf("ValidateSecureConfig rejected SetCredentialConfig output: %v", err)
	}
}

func TestLoadConfigReadsCredentialFieldsWithoutSecureValidation(t *testing.T) {
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
			resetViper(t)
			resetConfig(t)
			writeUserConfig(t, tt.config)

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig returned error: %v", err)
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

package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/urfave/cli/v2"
)

func resetSecureTestViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	cfgFile := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(cfgFile, nil, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yml")
	viper.SetDefault("account.secure", false)
	viper.SetDefault("account.custom.user", "")
	viper.SetDefault("account.custom.pass_hash", "")
	viper.SetDefault("account.custom.pass_hash_algo", "")
	viper.SetDefault("account.custom.pass", "")
	*conf.UnsafeCurrentForTest() = conf.Config{}
	t.Cleanup(func() {
		*conf.UnsafeCurrentForTest() = conf.Config{}
		viper.Reset()
	})
}

func newSecureTestContext(t *testing.T, values map[string]string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("sec", flag.ContinueOnError)
	set.String("user", "", "")
	set.String("pass", "", "")
	set.Bool("secure", false, "")
	for name, value := range values {
		if err := set.Set(name, value); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}

func TestSecureActionWritesBcryptCredentials(t *testing.T) {
	resetSecureTestViper(t)
	viper.Set("account.custom.pass", "legacy-md5")

	weight, err := secureAction(newSecureTestContext(t, map[string]string{
		"user": "admin",
		"pass": "secret",
	}))
	if err != nil {
		t.Fatalf("secureAction returned error: %v", err)
	}
	if weight != USER|PASS {
		t.Fatalf("weight = %d, want %d", weight, USER|PASS)
	}
	if got := viper.GetString("account.custom.user"); got != "admin" {
		t.Fatalf("account.custom.user = %q, want admin", got)
	}
	hash := viper.GetString("account.custom.pass_hash")
	if hash == "" {
		t.Fatal("account.custom.pass_hash is empty")
	}
	if hash == "secret" {
		t.Fatal("password stored in plaintext")
	}
	if got := viper.GetString("account.custom.pass_hash_algo"); got != security.HashAlgoBcrypt {
		t.Fatalf("account.custom.pass_hash_algo = %q, want %q", got, security.HashAlgoBcrypt)
	}
	if got := viper.GetString("account.custom.pass"); got != "" {
		t.Fatalf("legacy account.custom.pass = %q, want empty", got)
	}
	if err := security.VerifyPassword(hash, "secret"); err != nil {
		t.Fatalf("bcrypt hash did not verify password: %v", err)
	}
}

func TestSecureActionPasswordOnlyPreservesExistingUsername(t *testing.T) {
	resetSecureTestViper(t)
	viper.Set("account.custom.user", "admin")

	weight, err := secureAction(newSecureTestContext(t, map[string]string{"pass": "new-secret"}))
	if err != nil {
		t.Fatalf("secureAction returned error: %v", err)
	}
	if weight != PASS {
		t.Fatalf("weight = %d, want %d", weight, PASS)
	}
	if got := viper.GetString("account.custom.user"); got != "admin" {
		t.Fatalf("account.custom.user = %q, want admin", got)
	}
	if err := security.VerifyPassword(viper.GetString("account.custom.pass_hash"), "new-secret"); err != nil {
		t.Fatalf("bcrypt hash did not verify new password: %v", err)
	}
}

func TestSecureActionUsernameOnlyPreservesExistingPasswordHash(t *testing.T) {
	resetSecureTestViper(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", hash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)

	weight, err := secureAction(newSecureTestContext(t, map[string]string{"user": "root"}))
	if err != nil {
		t.Fatalf("secureAction returned error: %v", err)
	}
	if weight != USER {
		t.Fatalf("weight = %d, want %d", weight, USER)
	}
	if got := viper.GetString("account.custom.user"); got != "root" {
		t.Fatalf("account.custom.user = %q, want root", got)
	}
	if got := viper.GetString("account.custom.pass_hash"); got != hash {
		t.Fatalf("account.custom.pass_hash changed")
	}
	if err := security.VerifyPassword(viper.GetString("account.custom.pass_hash"), "secret"); err != nil {
		t.Fatalf("bcrypt hash no longer verifies password: %v", err)
	}
}

func TestHandleSecureRejectsLegacyOrMissingCredentials(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "legacy md5",
			setup: func() {
				viper.Set("account.custom.user", "21232f297a57a5a743894a0e4a801fc3")
				viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")
			},
		},
		{
			name:  "missing credentials",
			setup: func() {},
		},
		{
			name: "invalid bcrypt hash",
			setup: func() {
				viper.Set("account.custom.user", "admin")
				viper.Set("account.custom.pass_hash", "bad")
				viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
			},
		},
		{
			name: "legacy residue with bcrypt credentials",
			setup: func() {
				hash, err := security.HashPassword("secret")
				if err != nil {
					t.Fatalf("HashPassword returned error: %v", err)
				}
				viper.Set("account.custom.user", "admin")
				viper.Set("account.custom.pass_hash", hash)
				viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
				viper.Set("account.custom.pass", "legacy-md5")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSecureTestViper(t)
			tt.setup()

			err := Handle(SECU)
			if err == nil {
				t.Fatal("Handle accepted secure mode without bcrypt credentials")
			}
		})
	}
}

func TestSecureActionReturnsPasswordTooLong(t *testing.T) {
	resetSecureTestViper(t)
	viper.Set("account.custom.user", "admin")

	_, err := secureAction(newSecureTestContext(t, map[string]string{"pass": strings.Repeat("a", 73)}))
	if !errors.Is(err, security.ErrPasswordTooLong) {
		t.Fatalf("secureAction returned %v, want %v", err, security.ErrPasswordTooLong)
	}
}

func TestSecureActionDoesNotHalfWriteSecureOnPasswordTooLong(t *testing.T) {
	resetSecureTestViper(t)
	viper.Set("account.secure", false)
	viper.Set("account.custom.user", "admin")

	_, err := secureAction(newSecureTestContext(t, map[string]string{
		"secure": "true",
		"pass":   strings.Repeat("a", 73),
	}))
	if !errors.Is(err, security.ErrPasswordTooLong) {
		t.Fatalf("secureAction returned %v, want %v", err, security.ErrPasswordTooLong)
	}
	if got := viper.GetBool("account.secure"); got {
		t.Fatal("account.secure was set after password hash error")
	}
	if got := viper.GetString("account.custom.pass_hash"); got != "" {
		t.Fatalf("account.custom.pass_hash = %q, want empty", got)
	}
}

func TestSecureActionUsernameOnlyRejectsInvalidCredentialsWhenAlreadySecure(t *testing.T) {
	resetSecureTestViper(t)
	viper.Set("account.secure", true)
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", "bad")
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)

	_, err := secureAction(newSecureTestContext(t, map[string]string{"user": "root"}))
	if err == nil {
		t.Fatal("secureAction accepted username change while effective secure credentials were invalid")
	}
	if got := viper.GetString("account.custom.user"); got != "admin" {
		t.Fatalf("account.custom.user = %q, want admin", got)
	}
}

func TestSecureActionPasswordOnlyRepairsInvalidHashWhenAlreadySecure(t *testing.T) {
	resetSecureTestViper(t)
	viper.Set("account.secure", true)
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", "bad")
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)

	weight, err := secureAction(newSecureTestContext(t, map[string]string{"pass": "new-secret"}))
	if err != nil {
		t.Fatalf("secureAction returned error: %v", err)
	}
	if weight != PASS {
		t.Fatalf("weight = %d, want %d", weight, PASS)
	}
	if err := security.VerifyPassword(viper.GetString("account.custom.pass_hash"), "new-secret"); err != nil {
		t.Fatalf("new bcrypt hash did not verify password: %v", err)
	}
	if got := viper.GetString("account.custom.pass_hash_algo"); got != security.HashAlgoBcrypt {
		t.Fatalf("account.custom.pass_hash_algo = %q, want %q", got, security.HashAlgoBcrypt)
	}
}

func TestSecureActionUsernameOnlyRejectsLegacyResidueWhenAlreadySecure(t *testing.T) {
	resetSecureTestViper(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	viper.Set("account.secure", true)
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", hash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
	viper.Set("account.custom.pass", "legacy-md5")

	_, err = secureAction(newSecureTestContext(t, map[string]string{"user": "root"}))
	if err == nil {
		t.Fatal("secureAction accepted username change with legacy credential residue")
	}
	if got := viper.GetString("account.custom.user"); got != "admin" {
		t.Fatalf("account.custom.user = %q, want admin", got)
	}
}

func TestSecureActionSecureFalseAllowsInvalidCredentials(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "invalid hash",
			setup: func() {
				viper.Set("account.secure", true)
				viper.Set("account.custom.user", "admin")
				viper.Set("account.custom.pass_hash", "bad")
				viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
			},
		},
		{
			name: "legacy residue",
			setup: func() {
				hash, err := security.HashPassword("secret")
				if err != nil {
					t.Fatalf("HashPassword returned error: %v", err)
				}
				viper.Set("account.secure", true)
				viper.Set("account.custom.user", "admin")
				viper.Set("account.custom.pass_hash", hash)
				viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
				viper.Set("account.custom.pass", "legacy-md5")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSecureTestViper(t)
			tt.setup()

			weight, err := secureAction(newSecureTestContext(t, map[string]string{"secure": "false"}))
			if err != nil {
				t.Fatalf("secureAction returned error: %v", err)
			}
			if weight != 0 {
				t.Fatalf("weight = %d, want 0", weight)
			}
			if got := viper.GetBool("account.secure"); got {
				t.Fatal("account.secure was not disabled")
			}
		})
	}
}

func TestSecureCommandWritesSecureFalseToConfig(t *testing.T) {
	path := setupConfigTestFile(t, `
account:
  secure: true
  custom:
    user: admin
    pass_hash: bad
    pass_hash_algo: bcrypt
`)

	err := secureCmd().Action(newSecureTestContext(t, map[string]string{"secure": "false"}))
	if err == nil {
		t.Fatal("secure command returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("secure command returned %v, want exit code 0", err)
	}
	if got := viper.GetBool("account.secure"); got {
		t.Fatal("account.secure was not disabled in memory")
	}

	viper.Reset()
	viper.SetConfigFile(path)
	viper.SetConfigType("yml")
	if readErr := viper.ReadInConfig(); readErr != nil {
		t.Fatalf("ReadInConfig after secure command: %v", readErr)
	}
	if got := viper.GetBool("account.secure"); got {
		content, _ := os.ReadFile(path)
		t.Fatalf("account.secure was not disabled in config file:\n%s", string(content))
	}
}

func TestSecureCommandWithoutFlagsOnlyQueriesStatus(t *testing.T) {
	path := setupConfigTestFile(t, `
account:
  secure: false
`)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile config.yml before command: %v", err)
	}

	err = secureCmd().Action(newSecureTestContext(t, nil))
	if err == nil {
		t.Fatal("secure command returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("secure command returned %v, want exit code 0", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile config.yml after command: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("secure command without flags rewrote config file:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

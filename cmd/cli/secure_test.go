package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/tcp404/OneTiny/internal/config"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/urfave/cli/v2"
)

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

func newSecureStore(t *testing.T, content string) *config.Store {
	t.Helper()
	path := t.TempDir() + "/config.yml"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile config.yml: %v", err)
	}
	store := config.NewStore(path)
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load config: %v", err)
	}
	return store
}

func TestSecureActionWritesBcryptCredentials(t *testing.T) {
	store := newSecureStore(t, `
account:
  secure: false
  custom:
    pass: legacy-md5
`)

	weight, err := secureAction(store, newSecureTestContext(t, map[string]string{
		"user": "admin",
		"pass": "secret",
	}))
	if err != nil {
		t.Fatalf("secureAction returned error: %v", err)
	}
	if weight != USER|PASS {
		t.Fatalf("weight = %d, want %d", weight, USER|PASS)
	}
	cfg := store.Current()
	if cfg.Username != "admin" {
		t.Fatalf("username = %q, want admin", cfg.Username)
	}
	if cfg.PasswordHash == "" {
		t.Fatal("password hash is empty")
	}
	if cfg.PasswordHash == "secret" {
		t.Fatal("password stored in plaintext")
	}
	if cfg.PasswordHashAlgo != security.HashAlgoBcrypt {
		t.Fatalf("hash algo = %q, want %q", cfg.PasswordHashAlgo, security.HashAlgoBcrypt)
	}
	if cfg.LegacyPassword != "" {
		t.Fatalf("legacy password = %q, want empty", cfg.LegacyPassword)
	}
	if err := security.VerifyPassword(cfg.PasswordHash, "secret"); err != nil {
		t.Fatalf("bcrypt hash did not verify password: %v", err)
	}
}

func TestSecureActionReturnsPasswordTooLong(t *testing.T) {
	store := newSecureStore(t, `
account:
  secure: false
  custom:
    user: admin
`)

	_, err := secureAction(store, newSecureTestContext(t, map[string]string{"pass": strings.Repeat("a", 73)}))
	if !errors.Is(err, security.ErrPasswordTooLong) {
		t.Fatalf("secureAction returned %v, want %v", err, security.ErrPasswordTooLong)
	}
	if cfg := store.Current(); cfg.PasswordHash != "" {
		t.Fatalf("password hash = %q, want empty after failed hash", cfg.PasswordHash)
	}
}

func TestHandleSecureRejectsMissingCredentials(t *testing.T) {
	err := Handle(SECU, security.CredentialConfig{})
	if err == nil {
		t.Fatal("Handle accepted secure mode without credentials")
	}
}

func TestSecureCommandWithoutFlagsOnlyQueriesStatus(t *testing.T) {
	store := newSecureStore(t, `
account:
  secure: false
`)
	before, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("ReadFile before command: %v", err)
	}

	err = secureCmd(store).Action(newSecureTestContext(t, nil))
	if err == nil {
		t.Fatal("secure command returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("secure command returned %v, want exit code 0", err)
	}

	after, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("ReadFile after command: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("secure command without flags rewrote config file:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/tcp404/OneTiny/internal/config"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/urfave/cli/v2"
)

func newConfigTestContext(t *testing.T, values map[string]string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("config", flag.ContinueOnError)
	set.Int("port", 0, "")
	set.Bool("allow", false, "")
	set.Int("max", 0, "")
	set.String("road", "", "")
	set.Bool("secure", false, "")
	set.Int("scratch-max-items", 0, "")
	set.String("scratch-max-item-size", "", "")
	for name, value := range values {
		if err := set.Set(name, value); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}

func newConfigCommandStore(t *testing.T, content string) *config.Store {
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

func TestConfigActionSecureTrueRejectsInvalidCredentialsWithoutWritingSecure(t *testing.T) {
	store := newConfigCommandStore(t, `
account:
  secure: false
  custom:
    user: admin
    pass_hash: bad
    pass_hash_algo: bcrypt
`)

	err := configAction(store, newConfigTestContext(t, map[string]string{"secure": "true"}))
	if err == nil {
		t.Fatal("configAction accepted secure=true with invalid credentials")
	}
	content, readErr := os.ReadFile(store.Path())
	if readErr != nil {
		t.Fatalf("ReadFile config.yml: %v", readErr)
	}
	if strings.Contains(string(content), "secure: true") {
		t.Fatalf("account.secure was written to file:\n%s", content)
	}
}

func TestConfigActionSecureTrueAllowsValidBcryptCredentials(t *testing.T) {
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	store := newConfigCommandStore(t, `
account:
  secure: false
  custom:
    user: admin
    pass_hash: `+hash+`
    pass_hash_algo: bcrypt
`)

	err = configAction(store, newConfigTestContext(t, map[string]string{"secure": "true"}))
	if err == nil {
		t.Fatal("configAction returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("configAction returned %v, want exit code 0", err)
	}
	if got := store.Current().IsSecure; !got {
		t.Fatal("account.secure was not enabled")
	}
}

func TestConfigActionSecureFalseAllowsInvalidCredentials(t *testing.T) {
	store := newConfigCommandStore(t, `
account:
  secure: true
  custom:
    user: admin
    pass_hash: bad
    pass_hash_algo: bcrypt
`)

	err := configAction(store, newConfigTestContext(t, map[string]string{"secure": "false"}))
	if err == nil {
		t.Fatal("configAction returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("configAction returned %v, want exit code 0", err)
	}
	if got := store.Current().IsSecure; got {
		t.Fatal("account.secure was not disabled")
	}
}

func TestConfigActionWritesScratchFlags(t *testing.T) {
	store := newConfigCommandStore(t, `
scratch:
  max_items: 500
  max_item_size: 10MB
`)

	err := configAction(store, newConfigTestContext(t, map[string]string{
		"scratch-max-items":     "33",
		"scratch-max-item-size": "3MB",
	}))
	if err == nil {
		t.Fatal("configAction returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("configAction returned %v, want exit code 0", err)
	}
	if store.Current().ScratchMaxItems != 33 || store.Current().ScratchMaxItemSize != "3MB" {
		t.Fatalf("stored scratch = %d %q", store.Current().ScratchMaxItems, store.Current().ScratchMaxItemSize)
	}
}

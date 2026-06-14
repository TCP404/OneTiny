package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TCP404/OneTiny-cli/internal/security"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func newConfigTestContext(t *testing.T, values map[string]string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("config", flag.ContinueOnError)
	set.Bool("secure", false, "")
	for name, value := range values {
		if err := set.Set(name, value); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}

func setupConfigTestFile(t *testing.T, content string) string {
	t.Helper()
	resetSecureTestViper(t)
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile config.yml: %v", err)
	}
	viper.SetConfigFile(path)
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}
	return path
}

func TestConfigActionSecureTrueRejectsInvalidCredentialsWithoutWritingSecure(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{
			name: "legacy credentials",
			config: `
account:
  secure: false
  custom:
    user: 21232f297a57a5a743894a0e4a801fc3
    pass: 21232f297a57a5a743894a0e4a801fc3
`,
		},
		{
			name: "missing credentials",
			config: `
account:
  secure: false
`,
		},
		{
			name: "invalid bcrypt hash",
			config: `
account:
  secure: false
  custom:
    user: admin
    pass_hash: bad
    pass_hash_algo: bcrypt
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := setupConfigTestFile(t, tt.config)

			err := configAction(newConfigTestContext(t, map[string]string{"secure": "true"}))
			if err == nil {
				t.Fatal("configAction accepted secure=true with invalid credentials")
			}
			if got := viper.GetBool("account.secure"); got {
				t.Fatal("account.secure was set in memory")
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("ReadFile config.yml: %v", readErr)
			}
			if strings.Contains(string(content), "secure: true") {
				t.Fatalf("account.secure was written to file:\n%s", content)
			}
		})
	}
}

func TestConfigActionSecureTrueAllowsValidBcryptCredentials(t *testing.T) {
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	setupConfigTestFile(t, `
account:
  secure: false
  custom:
    user: admin
    pass_hash: `+hash+`
    pass_hash_algo: bcrypt
`)

	err = configAction(newConfigTestContext(t, map[string]string{"secure": "true"}))
	if err == nil {
		t.Fatal("configAction returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("configAction returned %v, want exit code 0", err)
	}
	if got := viper.GetBool("account.secure"); !got {
		t.Fatal("account.secure was not enabled")
	}
}

func TestConfigActionSecureFalseAllowsInvalidCredentials(t *testing.T) {
	setupConfigTestFile(t, `
account:
  secure: true
  custom:
    user: admin
    pass_hash: bad
    pass_hash_algo: bcrypt
`)

	err := configAction(newConfigTestContext(t, map[string]string{"secure": "false"}))
	if err == nil {
		t.Fatal("configAction returned nil, want cli exit")
	}
	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() != 0 {
		t.Fatalf("configAction returned %v, want exit code 0", err)
	}
	if got := viper.GetBool("account.secure"); got {
		t.Fatal("account.secure was not disabled")
	}
}

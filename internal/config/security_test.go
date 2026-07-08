package config

import (
	"testing"

	"github.com/tcp404/OneTiny/internal/security"
)

func TestCredentialConfigFromConfigReadsBcryptFields(t *testing.T) {
	cfg := Config{
		Username:         "admin",
		PasswordHash:     validBcryptHash,
		PasswordHashAlgo: security.HashAlgoBcrypt,
	}

	creds := CredentialConfigFromConfig(cfg)
	if creds.Username != "admin" {
		t.Fatalf("username = %q, want admin", creds.Username)
	}
	if creds.PasswordHash == "" {
		t.Fatal("missing password hash")
	}
	if creds.HashAlgo != security.HashAlgoBcrypt {
		t.Fatalf("hash algo = %q, want bcrypt", creds.HashAlgo)
	}
}

func TestValidateSecureConfigForRejectsLegacyWhenEffectiveSecure(t *testing.T) {
	store := newTestStore(t, `
account:
  secure: false
  custom:
    user: 21232f297a57a5a743894a0e4a801fc3
    pass: 21232f297a57a5a743894a0e4a801fc3
`)
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if err := store.ValidateSecureConfigFor(true); err == nil {
		t.Fatal("ValidateSecureConfigFor accepted legacy MD5 with effective secure mode")
	}
}

func TestValidateSecureConfigForAllowsUnprotectedLegacyConfig(t *testing.T) {
	store := newTestStore(t, `
account:
  secure: false
  custom:
    pass: 21232f297a57a5a743894a0e4a801fc3
`)
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if err := store.ValidateSecureConfigFor(false); err != nil {
		t.Fatalf("unprotected legacy config rejected: %v", err)
	}
}

func TestPatchSecurityWritesBcryptAndClearsLegacy(t *testing.T) {
	store := newTestStore(t, `
account:
  secure: true
  custom:
    pass: legacy-md5
`)
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	username := "admin"
	passwordHash := validBcryptHash
	got, err := store.PatchSecurity(SecurityPatch{
		Username:     &username,
		PasswordHash: &passwordHash,
	})
	if err != nil {
		t.Fatalf("PatchSecurity returned error: %v", err)
	}
	if got.Username != "admin" {
		t.Fatalf("Username = %q, want admin", got.Username)
	}
	if got.PasswordHash != validBcryptHash {
		t.Fatalf("PasswordHash = %q, want %q", got.PasswordHash, validBcryptHash)
	}
	if got.PasswordHashAlgo != security.HashAlgoBcrypt {
		t.Fatalf("PasswordHashAlgo = %q, want %q", got.PasswordHashAlgo, security.HashAlgoBcrypt)
	}
	if got.LegacyPassword != "" {
		t.Fatalf("LegacyPassword = %q, want empty", got.LegacyPassword)
	}
	if err := store.ValidateSecureConfigFor(true); err != nil {
		t.Fatalf("ValidateSecureConfigFor rejected PatchSecurity output: %v", err)
	}
}

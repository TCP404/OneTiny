package security

import (
	"errors"
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "correct horse battery staple" {
		t.Fatal("HashPassword returned the plain password")
	}
	if err := VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword rejected valid password: %v", err)
	}
	if err := VerifyPassword(hash, "wrong"); err == nil {
		t.Fatal("VerifyPassword accepted invalid password")
	}
}

func TestCredentialConfigValidation(t *testing.T) {
	valid := CredentialConfig{
		Username:     "admin",
		PasswordHash: "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi47OhMKn8IW0YFDCw5Ac1TYT2RP1xG",
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := valid.ValidateForSecureMode(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	legacy := CredentialConfig{
		Username:  "21232f297a57a5a743894a0e4a801fc3",
		LegacyMD5: "21232f297a57a5a743894a0e4a801fc3",
	}
	if err := legacy.ValidateForSecureMode(); !errors.Is(err, ErrLegacyMD5Config) {
		t.Fatalf("legacy MD5 config returned %v, want %v", err, ErrLegacyMD5Config)
	}

	missing := CredentialConfig{}
	if err := missing.ValidateForSecureMode(); !errors.Is(err, ErrMissingCredentials) {
		t.Fatalf("missing credentials returned %v, want %v", err, ErrMissingCredentials)
	}

	unsupportedHash := CredentialConfig{
		Username:     "admin",
		PasswordHash: valid.PasswordHash,
		HashAlgo:     "argon2",
	}
	if err := unsupportedHash.ValidateForSecureMode(); !errors.Is(err, ErrUnsupportedHash) {
		t.Fatalf("unsupported hash returned %v, want %v", err, ErrUnsupportedHash)
	}

	invalidHash := CredentialConfig{
		Username:     "admin",
		PasswordHash: "bad",
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := invalidHash.ValidateForSecureMode(); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("invalid bcrypt hash returned %v, want %v", err, ErrInvalidPasswordHash)
	}

	tamperedPayload := CredentialConfig{
		Username:     "admin",
		PasswordHash: valid.PasswordHash[:len(valid.PasswordHash)-1] + "!",
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := tamperedPayload.ValidateForSecureMode(); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("tampered bcrypt payload returned %v, want %v", err, ErrInvalidPasswordHash)
	}

	invalidTail := CredentialConfig{
		Username:     "admin",
		PasswordHash: valid.PasswordHash[:len(valid.PasswordHash)-1] + "H",
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := invalidTail.ValidateForSecureMode(); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("invalid bcrypt tail returned %v, want %v", err, ErrInvalidPasswordHash)
	}

	invalidSaltTail := CredentialConfig{
		Username:     "admin",
		PasswordHash: valid.PasswordHash[:28] + "B" + valid.PasswordHash[29:],
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := invalidSaltTail.ValidateForSecureMode(); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("invalid bcrypt salt tail returned %v, want %v", err, ErrInvalidPasswordHash)
	}

	excessiveCost := CredentialConfig{
		Username:     "admin",
		PasswordHash: valid.PasswordHash[:4] + "31" + valid.PasswordHash[6:],
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := excessiveCost.ValidateForSecureMode(); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("excessive bcrypt cost returned %v, want %v", err, ErrInvalidPasswordHash)
	}

	corruptPayload := CredentialConfig{
		Username:     "admin",
		PasswordHash: "$2a$10$!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!",
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := corruptPayload.ValidateForSecureMode(); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("corrupt bcrypt payload returned %v, want %v", err, ErrInvalidPasswordHash)
	}
}

func TestPasswordLengthLimit(t *testing.T) {
	if _, err := HashPassword(strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("HashPassword returned %v, want %v", err, ErrPasswordTooLong)
	}

	hash, err := HashPassword(strings.Repeat("a", 72))
	if err != nil {
		t.Fatalf("HashPassword returned error for 72-byte password: %v", err)
	}
	if err := VerifyPassword(hash, strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("VerifyPassword returned %v, want %v", err, ErrPasswordTooLong)
	}
}

func TestCredentialConfigState(t *testing.T) {
	config := CredentialConfig{
		Username:     " admin ",
		PasswordHash: " hash ",
		HashAlgo:     HashAlgoBcrypt,
	}
	if !config.IsConfigured() {
		t.Fatal("bcrypt credential config was not treated as configured")
	}

	legacy := CredentialConfig{LegacyMD5: " 21232f297a57a5a743894a0e4a801fc3 "}
	if !legacy.HasLegacyMD5() {
		t.Fatal("legacy MD5 config was not detected")
	}
}

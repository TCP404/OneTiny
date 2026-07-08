package verify

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/security"
)

func resetUPSTestViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestUPSVerifierUsesBcryptCredentialFields(t *testing.T) {
	resetUPSTestViper(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", hash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)

	if err := NewUPSVerifier(uint8(SECU)).Handle(); err != nil {
		t.Fatalf("Handle rejected bcrypt credentials: %v", err)
	}
}

func TestUPSVerifierRejectsLegacyCredentials(t *testing.T) {
	resetUPSTestViper(t)
	viper.Set("account.custom.user", "21232f297a57a5a743894a0e4a801fc3")
	viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")

	if err := NewUPSVerifier(uint8(SECU)).Handle(); err == nil {
		t.Fatal("Handle accepted legacy credentials")
	}
}

func TestUPSVerifierRejectsInvalidBcryptHash(t *testing.T) {
	resetUPSTestViper(t)
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", "bad")
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)

	if err := NewUPSVerifier(uint8(SECU)).Handle(); err == nil {
		t.Fatal("Handle accepted invalid bcrypt hash")
	}
}

func TestUPSVerifierRejectsLegacyResidue(t *testing.T) {
	resetUPSTestViper(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", hash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
	viper.Set("account.custom.pass", "legacy-md5")

	if err := NewUPSVerifier(uint8(SECU)).Handle(); err == nil {
		t.Fatal("Handle accepted bcrypt credentials with legacy residue")
	}
}

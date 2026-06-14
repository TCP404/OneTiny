package conf

import (
	"github.com/TCP404/OneTiny-cli/internal/security"
	"github.com/spf13/viper"
)

func CredentialConfigFromViper() security.CredentialConfig {
	return security.CredentialConfig{
		Username:     viper.GetString("account.custom.user"),
		PasswordHash: viper.GetString("account.custom.pass_hash"),
		HashAlgo:     viper.GetString("account.custom.pass_hash_algo"),
		LegacyMD5:    viper.GetString("account.custom.pass"),
	}
}

func ValidateSecureConfig() error {
	return ValidateSecureConfigFor(viper.GetBool("account.secure"))
}

func ValidateSecureConfigFor(secure bool) error {
	if !secure {
		return nil
	}
	return CredentialConfigFromViper().ValidateForSecureMode()
}

func SetCredentialConfig(username, passwordHash string) {
	viper.Set("account.custom.user", username)
	viper.Set("account.custom.pass_hash", passwordHash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
	viper.Set("account.custom.pass", "")
}

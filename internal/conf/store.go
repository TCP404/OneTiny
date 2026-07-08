package conf

import (
	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/security"
)

type SecurityPatch struct {
	IsSecure     *bool
	Username     *string
	PasswordHash *string
}

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

func CredentialConfigFromConfig(cfg Config) security.CredentialConfig {
	return security.CredentialConfig{
		Username:     cfg.Username,
		PasswordHash: cfg.PasswordHash,
		HashAlgo:     cfg.PasswordHashAlgo,
		LegacyMD5:    cfg.LegacyPassword,
	}
}

func SaveSecurityPatch(patch SecurityPatch) (Config, error) {
	rollback := captureViperKeys(
		"account.secure",
		"account.custom.user",
		"account.custom.pass_hash",
		"account.custom.pass_hash_algo",
		"account.custom.pass",
	)
	originalConfig := currentConfig

	ensureCurrentConfigFromViper()
	setConfigValues(currentConfig)
	if patch.IsSecure != nil {
		viper.Set("account.secure", *patch.IsSecure)
	}
	if patch.Username != nil {
		viper.Set("account.custom.user", *patch.Username)
	}
	if patch.PasswordHash != nil {
		viper.Set("account.custom.pass_hash", *patch.PasswordHash)
		viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
		viper.Set("account.custom.pass", "")
	}
	if err := ValidateSecureConfig(); err != nil {
		rollback.restore()
		currentConfig = originalConfig
		return Config{}, err
	}
	if err := writeCurrentViperConfigAtomic(); err != nil {
		rollback.restore()
		currentConfig = originalConfig
		return Config{}, err
	}
	cfg, err := LoadConfig()
	if err != nil {
		rollback.restore()
		currentConfig = originalConfig
		return Config{}, err
	}
	return cfg, nil
}

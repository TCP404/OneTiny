package config

import (
	"github.com/tcp404/OneTiny/internal/security"
)

type SecurityPatch struct {
	IsSecure     *bool
	Username     *string
	PasswordHash *string
}

func CredentialConfigFromConfig(cfg Config) security.CredentialConfig {
	return security.CredentialConfig{
		Username:     cfg.Username,
		PasswordHash: cfg.PasswordHash,
		HashAlgo:     cfg.PasswordHashAlgo,
		LegacyMD5:    cfg.LegacyPassword,
	}
}

func (s *Store) ValidateSecureConfigFor(secure bool) error {
	cfg, err := s.ensureCurrent()
	if err != nil {
		return err
	}
	return validateSecureConfigFor(cfg, secure)
}

func (s *Store) PatchSecurity(patch SecurityPatch) (Config, error) {
	cfg, err := s.ensureCurrent()
	if err != nil {
		return Config{}, err
	}
	if patch.IsSecure != nil {
		cfg.IsSecure = *patch.IsSecure
	}
	if patch.Username != nil {
		cfg.Username = *patch.Username
	}
	if patch.PasswordHash != nil {
		cfg.PasswordHash = *patch.PasswordHash
		cfg.PasswordHashAlgo = security.HashAlgoBcrypt
		cfg.LegacyPassword = ""
	}
	if err := validateSecureConfigFor(cfg, cfg.IsSecure); err != nil {
		return Config{}, err
	}
	if err := s.writeConfig(cfg); err != nil {
		return Config{}, err
	}
	return s.Load()
}

func validateSecureConfigFor(cfg Config, secure bool) error {
	if !secure {
		return nil
	}
	return CredentialConfigFromConfig(cfg).ValidateForSecureMode()
}

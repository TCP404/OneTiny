package security

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	HashAlgoBcrypt        = "bcrypt"
	MaxAcceptedBcryptCost = bcrypt.DefaultCost + 4
	MaxPasswordBytes      = 72
)

var (
	ErrMissingCredentials  = errors.New("开启访问登录需先设置帐号密码")
	ErrLegacyMD5Config     = errors.New("检测到旧版 MD5 账号密码配置，请重新设置账号密码")
	ErrUnsupportedHash     = errors.New("不支持的密码哈希算法")
	ErrInvalidPasswordHash = errors.New("密码哈希无效，请重新设置账号密码")
	ErrPasswordTooLong     = errors.New("密码长度不能超过 72 字节")
)

type CredentialConfig struct {
	Username     string
	PasswordHash string
	HashAlgo     string
	LegacyMD5    string
}

func HashPassword(password string) (string, error) {
	if len([]byte(password)) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) error {
	if len([]byte(password)) > MaxPasswordBytes {
		return ErrPasswordTooLong
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func isValidBcryptHash(hash string) bool {
	if !hasValidBcryptStructure(hash) {
		return false
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil || cost > MaxAcceptedBcryptCost {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("onetiny-validation-probe"))
	return err == nil || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword)
}

func hasValidBcryptStructure(hash string) bool {
	if len(hash) != 60 {
		return false
	}
	if hash[0] != '$' || hash[1] != '2' || hash[3] != '$' || hash[6] != '$' {
		return false
	}
	if hash[2] != 'a' && hash[2] != 'b' && hash[2] != 'y' {
		return false
	}
	if !isDigit(hash[4]) || !isDigit(hash[5]) {
		return false
	}
	for i := 7; i < len(hash); i++ {
		if !isBcryptBase64(hash[i]) {
			return false
		}
	}
	if bcryptBase64Index(hash[59])%4 != 0 {
		return false
	}
	if bcryptBase64Index(hash[28])%16 != 0 {
		return false
	}
	return true
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isBcryptBase64(b byte) bool {
	return bcryptBase64Index(b) != -1
}

func bcryptBase64Index(b byte) int {
	switch {
	case b == '.':
		return 0
	case b == '/':
		return 1
	case b >= 'A' && b <= 'Z':
		return int(b-'A') + 2
	case b >= 'a' && b <= 'z':
		return int(b-'a') + 28
	case b >= '0' && b <= '9':
		return int(b-'0') + 54
	default:
		return -1
	}
}

// IsConfigured only checks field presence; use ValidateForSecureMode before enabling secure mode.
func (c CredentialConfig) IsConfigured() bool {
	return strings.TrimSpace(c.Username) != "" &&
		strings.TrimSpace(c.PasswordHash) != "" &&
		c.HashAlgo == HashAlgoBcrypt
}

func (c CredentialConfig) HasLegacyMD5() bool {
	return strings.TrimSpace(c.LegacyMD5) != ""
}

func (c CredentialConfig) ValidateForSecureMode() error {
	if c.HasLegacyMD5() {
		return ErrLegacyMD5Config
	}
	if strings.TrimSpace(c.Username) == "" || strings.TrimSpace(c.PasswordHash) == "" {
		return ErrMissingCredentials
	}
	if c.HashAlgo != HashAlgoBcrypt {
		return ErrUnsupportedHash
	}
	if !isValidBcryptHash(c.PasswordHash) {
		return ErrInvalidPasswordHash
	}
	return nil
}

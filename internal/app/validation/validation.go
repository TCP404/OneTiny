package validation

import (
	"errors"
	"os"
	stdruntime "runtime"

	"github.com/fatih/color"
	"github.com/tcp404/OneTiny/internal/security"
)

func ValidatePort(port int) error {
	switch stdruntime.GOOS {
	case "linux", "darwin":
		if port < 1024 || port > 65535 {
			return errors.New(color.RedString("不可以设置系统预留端口 %d, 您可以设置的范围在 [ 1024 ~ 65535 ] 之间。", port))
		}
	case "windows":
		if port < 5001 || port > 65535 {
			return errors.New(color.RedString("不可以设置系统预留端口 %d, 您可以设置的范围在 [ 5001 ~ 65535 ] 之间。", port))
		}
	}
	return nil
}

func ValidatePath(rootPath string) error {
	if _, err := os.Stat(rootPath); err != nil {
		if !os.IsExist(err) {
			return errors.New(color.RedString("无法设置您指定的共享路径, 请检查给出的路径是否有问题：%s", rootPath))
		}
	}
	return nil
}

func ValidateSecureTransition(weight uint8, credentials security.CredentialConfig) error {
	switch ups(weight) {
	case user | pass | secure:
		return validateCredentialsForSecureMode(credentials)
	case user | pass:
		return nil
	case user | secure:
		return validateCredentialsForSecureMode(credentials)
	case user:
		if credentials.IsConfigured() {
			return nil
		}
		return errors.New("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
	case pass | secure:
		return validateCredentialsForSecureMode(credentials)
	case pass:
		if credentials.Username != "" {
			return nil
		}
		return errors.New("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
	case secure:
		return validateCredentialsForSecureMode(credentials)
	case 0:
		return nil
	default:
		return errors.New("设置失败～")
	}
}

type ups uint8

const (
	user   ups = 1 << 0
	pass   ups = 1 << 1
	secure ups = 1 << 2
)

func validateCredentialsForSecureMode(credentials security.CredentialConfig) error {
	if err := credentials.ValidateForSecureMode(); err != nil {
		if errors.Is(err, security.ErrMissingCredentials) {
			return errors.New("开启访问登录需先设置帐号密码，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
		}
		return err
	}
	return nil
}

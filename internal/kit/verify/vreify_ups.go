package verify

import (
	"errors"

	"github.com/TCP404/OneTiny-cli/pkg/container"
	"github.com/TCP404/eutil"
	"github.com/spf13/viper"
)

// ups 是 User、Pass、Secure 三个单词的首字母合并，
// 用于设置 帐号、密码、访问登录 三个配置项的权值
// 用法和 unix 系统中文件权限的 rwx 相同
type ups uint8

// USER 表示 已设置账户
// PASS 表示 已设置密码
// SECU 表示 已开启访问登录
const (
	USER ups = 1 << 0 // The weight of USER is 1
	PASS ups = 1 << 1 // The weight of PASS is 2
	SECU ups = 1 << 2 // The weight of SECU is 4
)

type upsVerifier struct {
	weight ups
}

var _ container.Handler = (*upsVerifier)(nil)

func NewUPSVerifier(weight uint8) *upsVerifier { return &upsVerifier{weight: ups(weight)} }

// 检查UPS
// 000 (0) 表示 用户仅仅执行了命令 `onetiny sec`，返回访问登录开启状态
// 001 (1) 表示 开启访问登录
// 010 (2) 表示 设置密码
// 011 (3) 表示 开启访问登录，并设置密码
// 100 (4) 表示 设置用户名
// 101 (5) 表示 开启访问登录，并设置帐户名
// 110 (6) 表示 设置账户、密码
// 111 (7) 表示 开启访问登录，并设置账户、密码
//
// 设置规则：
// 开启访问登录时，需配置文件中已设置帐号密码
// 设置密码时，需配置文件中已设置账户
// 设置账户时，需配置文件中已设置密码
func (u *upsVerifier) Handle() error {
	switch u.weight {
	case USER | PASS | SECU:
		// 111 开启访问登录，并设置账户和密码
		return nil
	case USER | PASS:
		// 110 设置账户和密码
		return nil
	case USER | SECU:
		// 101 开启访问登录，并设置用户名，穿透下去检查是否有密码
		fallthrough
	case USER:
		// 100 设置用户名，需配置文件中有密码
		return eutil.If(viper.GetString("account.custom.pass") != "", nil, errors.New("未设置密码"))
	case PASS | SECU:
		// 011 开启访问登录，并设置密码，穿透下去检查是否有帐户名
		fallthrough
	case PASS:
		// 010 设置密码，需配置文件中有账户名
		return eutil.If(viper.GetString("account.custom.user") != "", nil, errors.New("未设置帐号"))
	case SECU:
		// 001 开启访问登录
		return eutil.If(viper.GetString("account.custom.user") != "" && viper.GetString("account.custom.pass") != "", nil, errors.New("未设置帐号和密码"))
	case 0:
		// 000 打印当前是否开启访问登录
		return nil
	default:
		return errors.New("设置失败～")
	}
}

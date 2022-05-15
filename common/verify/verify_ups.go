package verify

import (
	"errors"
	"oneTiny/common"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
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

func Register(c *cli.Context) (ups, error) {
	var setSECU, setUSER, setPASS ups = 0, 0, 0

	// 当填写了 -s 选项并且 -s 的值为 true 时才设置
	if is, s := c.IsSet("secure"), c.Bool("secure"); is {
		if s {
			setSECU = SECU
		}
		viper.Set("account.secure", s)
	}
	// 当填写了 -u 选项并且 -u 的值不为 空 时才设置
	if is, u := c.IsSet("user"), c.String("user"); is && u != "" {
		setUSER = USER
		viper.Set("account.custom.user", common.MD5(u))
	}
	// 当填写了 -p 选项并且 -p 的值不为 空 时才设置
	if is, p := c.IsSet("pass"), c.String("pass"); is && p != "" {
		setPASS = PASS
		viper.Set("account.custom.pass", common.MD5(p))
	}

	weight := setSECU | setUSER | setPASS
	if !verifyUPS(weight) {
		return weight, errors.New("设置失败~")
	}
	if err := viper.WriteConfig(); err != nil {
		return weight, errors.New("配置文件写入失败～")
	}
	return weight, nil
}

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
func verifyUPS(weight ups) bool {
	switch weight {
	case USER | PASS | SECU:
		// 111 开启访问登录，并设置账户和密码
		return true
	case USER | PASS:
		// 110 设置账户和密码
		return true
	case USER | SECU:
		// 101 开启访问登录，并设置用户名，穿透下去检查是否有密码
		fallthrough
	case USER:
		// 100 设置用户名，需配置文件中有密码
		return viper.GetString("account.custom.pass") != ""
	case PASS | SECU:
		// 011 开启访问登录，并设置密码，穿透下去检查是否有帐户名
		fallthrough
	case PASS:
		// 010 设置密码，需配置文件中有账户名
		return viper.GetString("account.custom.user") != ""
	case SECU:
		// 001 开启访问登录
		return viper.GetString("account.custom.user") != "" && viper.GetString("account.custom.pass") != ""
	case 0:
		// 000 打印当前是否开启访问登录
		return true
	default:
		return false
	}
}

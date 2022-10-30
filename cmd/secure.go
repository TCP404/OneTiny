package cmd

import (
	// "github.com/TCP404/OneTiny-cli/common/verify"

	"errors"

	"github.com/TCP404/eutil"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func secureCmd() *cli.Command {
	return &cli.Command{
		Name:        "sec",
		Aliases:     []string{"s"},
		Usage:       "设置访问登录的账户和密码",
		UsageText:   "onetiny sec [OPTIONS]",
		Description: "使用 onetiny sec 命令可以设置访问登录的帐号密码。\n允许的命令形式如下：\n 注册并开启：\t onetiny sec -u=账户名 -p=密码 -s\n 注册/覆盖账户：onetiny sec -u=账户名 -p=密码\n 重设密码：\t\t onetiny sec -p=密码",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "设置访问登录的`账户`名",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "pass",
				Aliases:  []string{"p"},
				Usage:    "设置访问登录的`密码`",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "secure",
				Aliases:  []string{"s"},
				Usage:    "设置`开启`访问登录，效果同 onetiny -s 一样",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			weight, err := secureAction(c)
			if err != nil {
				return cli.Exit(color.RedString(err.Error()), 21)
			}
			if weight == 0 {
				return cli.Exit(color.GreenString("当前访问登录是否已开启: %t", viper.GetBool("account.secure")), 0)
			}
			if err := viper.WriteConfig(); err != nil {
				return cli.Exit(color.RedString("配置文件写入失败～"), 22)
			}
			return cli.Exit(color.GreenString("设置成功~"), 0)
		},
	}
}

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

func secureAction(c *cli.Context) (ups, error) {
	var weight ups
	// 当填写了 -s 选项并且 -s 的值为 true 时才设置
	if is, s := c.IsSet("secure"), c.Bool("secure"); is {
		if s {
			weight |= SECU
		}
		viper.Set("account.secure", s)
	}
	// 当填写了 -u 选项并且 -u 的值不为 空 时才设置
	if is, u := c.IsSet("user"), c.String("user"); is && u != "" {
		weight |= USER
		viper.Set("account.custom.user", eutil.MD5(u))
	}
	// 当填写了 -p 选项并且 -p 的值不为 空 时才设置
	if is, p := c.IsSet("pass"), c.String("pass"); is && p != "" {
		weight |= PASS
		viper.Set("account.custom.pass", eutil.MD5(p))
	}
	return weight, Handle(weight)
}

// Handle 检查UPS
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
func Handle(weight ups) error {
	switch weight {
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
		return eutil.If(viper.GetString("account.custom.pass") != "", nil, errors.New("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。"))
	case PASS | SECU:
		// 011 开启访问登录，并设置密码，穿透下去检查是否有帐户名
		fallthrough
	case PASS:
		// 010 设置密码，需配置文件中有账户名
		return eutil.If(viper.GetString("account.custom.user") != "", nil, errors.New("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。"))
	case SECU:
		// 001 开启访问登录
		return eutil.If(viper.GetString("account.custom.user") != "" && viper.GetString("account.custom.pass") != "", nil, errors.New("开启访问登录需先设置帐号密码，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。"))
	case 0:
		// 000 打印当前是否开启访问登录
		return nil
	default:
		return errors.New("设置失败～")
	}
}

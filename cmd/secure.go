package cmd

import (
	"oneTiny/common/verify"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var secureCmd = newSecureCmd()

func newSecureCmd() *cli.Command {
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
			weight, err := verify.Register(c)
			if err != nil {
				switch weight {
				case 1:
					color.Red("开启访问登录需先设置帐号密码，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
				case 2, 3:
					color.Red("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
				case 4, 5:
					color.Red("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
				}
				return cli.Exit(err.Error(), 21)
			}
			if weight == 0 {
				return cli.Exit(color.GreenString("当前访问登录是否已开启: %t", viper.GetBool("account.secure")), 0)
			}
			return cli.Exit(color.GreenString("设置成功~"), 0)
		},
	}
}

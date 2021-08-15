package start

//  onetiny
//  ├── -r
//  ├── -p
//  ├── -a
//  ├── -x
//  ├── -s
//  ├── config
//  │   ├── -r
//  │   ├── -p
//  │   ├── -a
//  │   ├── -x
//  │   └── -s
//  ├── sec
//  │   ├── -u
//  │   ├── -p
//  │   └── -s
//  └── update
//      └── -l

import (
	"fmt"
	"io"
	"log"
	"oneTiny/config"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func initCLI() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "打印版本信息",
	}
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("当前版本: ", c.App.Version)
		os.Exit(0)
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "打印帮助信息",
	}
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		cli.HelpPrinterCustom(w, templ, data, nil)
		os.Exit(0)
	}
	cli.ErrWriter = config.Output
}

// RunCLI 函数作为程序入口，主要负责处理命令和 flag
func RunCLI() {
	initCLI()
	globalFlag := []cli.Flag{
		&cli.PathFlag{
			Name:        "road",
			Aliases:     []string{"r"},
			Usage:       "指定对外开放的目录`路径`",
			Value:       config.RootPath,
			Required:    false,
			DefaultText: config.RootPath,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "指定开放的`端口`",
			Value:       config.Port,
			Required:    false,
			DefaultText: strconv.Itoa(config.Port),
		},
		&cli.BoolFlag{
			Name:        "allow",
			Aliases:     []string{"a"},
			Usage:       "指定`是否`允许访问者上传",
			Value:       config.IsAllowUpload,
			Required:    false,
			DefaultText: strconv.FormatBool(config.IsAllowUpload),
		},
		&cli.IntFlag{
			Name:        "max",
			Aliases:     []string{"x"},
			Usage:       "指定允许访问的`深度`，默认仅限访问共享目录",
			Value:       int(config.MaxLevel),
			Required:    false,
			DefaultText: "0",
		},
		&cli.BoolFlag{
			Name:        "secure",
			Aliases:     []string{"s"},
			Usage:       "指定是否开启访问登录",
			Value:       config.IsSecure,
			Required:    false,
			DefaultText: strconv.FormatBool(config.IsSecure),
		},
	}

	secCmd := &cli.Command{
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
			rec, err := config.Register(c)
			if err != nil {
				switch rec {
				case 1:
					color.Red("开启访问登录需先设置帐号密码，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
				case 2, 3:
					color.Red("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
				case 4, 5:
					color.Red("未找到您的帐号，请使用 `onetiny sec -u=帐号 -p=密码` 进行设置。")
				}
				return err
			}
			if rec == 0 {
				return cli.Exit(color.GreenString("当前访问登录是否已开启: %t", viper.GetBool("account.secure")), 0)
			}
			return cli.Exit(color.GreenString("设置成功~"), 0)
		},
	}

	updateCmd := &cli.Command{
		Name:    "update",
		Aliases: []string{"u", "up"},
		Usage:   "更新 OneTiny 到最新版",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "list",
				Aliases:     []string{"l"},
				Usage:       "列出远程服务器上所有可用版本",
				Required:    false,
				DefaultText: "false",
			},
		},
	}

	configCmd := &cli.Command{
		Name:        "config",
		Aliases:     []string{"c", "cf", "cfg", "conf"},
		Usage:       "设置默认配置",
		UsageText:   "onetiny config [OPTIONS]",
		Description: "使用 onetiny config 命令可以将设置写入配置文件。\n使用方式与 onetiny 命令相同，仅多了一个 config 关键字，如：\n  onetiny config -p 10240  可以将端口设置为 10240 写入配置\n  onetiny config -a false  可以设置不允许访问者上传并写入配置",
		Flags:       globalFlag,
		Action: func(c *cli.Context) error {
			if err := config.Set(c); err != nil {
				return cli.Exit(err.Error(), 11)
			}
			return cli.Exit(color.GreenString("配置成功~"), 0)
		},
	}

	app := &cli.App{
		Name:            "OneTiny",
		Usage:           "一个用于局域网内共享文件的FTP程序",
		UsageText:       "onetiny [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS] [参数...]",
		Version:         config.VERSION,
		Flags:           globalFlag,
		Authors:         []*cli.Author{{Name: "Boii", Email: "i@tcp404.com"}},
		Commands:        []*cli.Command{updateCmd, configCmd, secCmd},
		CommandNotFound: func(c *cli.Context, s string) { cli.ShowAppHelpAndExit(c, 10) },
		Writer:          config.Output,
		ErrWriter:       config.Output,
		Action: func(c *cli.Context) error {
			config.Port = c.Int("port")
			config.RootPath = c.String("road")
			config.MaxLevel = uint8(c.Int("max"))
			config.IsAllowUpload = c.Bool("allow")
			config.IsSecure = c.Bool("secure")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(color.RedString("%v", err))
	}
}

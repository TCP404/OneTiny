package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/constant"
	"github.com/TCP404/OneTiny-cli/internal/kit/verify"
	"github.com/TCP404/OneTiny-cli/pkg/container"

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
	cli.HelpPrinter = func(w io.Writer, temp string, data interface{}) {
		cli.HelpPrinterCustom(w, temp, data, nil)
		os.Exit(0)
	}
	cli.ErrWriter = conf.Config.Output
}

// CLI 函数作为程序入口，主要负责处理命令和 flag
func CLI() *cli.App {
	initCLI()

	return &cli.App{
		Name:            "OneTiny",
		Usage:           "一个用于局域网内共享文件的FTP程序",
		UsageText:       "onetiny [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS] [参数...]",
		Version:         constant.VERSION,
		Flags:           newGlobalFlag(),
		Authors:         []*cli.Author{{Name: "Boii", Email: "i@tcp404.com"}},
		Commands:        []*cli.Command{updateCmd(), configCmd(), secureCmd()},
		CommandNotFound: func(c *cli.Context, s string) { cli.ShowAppHelpAndExit(c, 10) },
		Writer:          conf.Config.Output,
		ErrWriter:       conf.Config.Output,
		After:           afterRootAction,
		Action:          rootAction,
	}
}

func afterRootAction(c *cli.Context) error {
	return container.NewHandleChain().
		AddToHead(verify.NewPortVerifier(conf.Config.Port)).
		AddToHead(verify.NewPathVerifier(conf.Config.RootPath)).
		Iterator()
}

func rootAction(c *cli.Context) error {
	conf.Config.Port = c.Int("port")
	conf.Config.MaxLevel = uint8(c.Int("max"))
	conf.Config.IsAllowUpload = c.Bool("allow")
	conf.Config.IsSecure = c.Bool("secure")
	if c.IsSet("road") {
		road := c.Path("road")
		if road[0] == '.' {
			pwd, _ := os.Getwd()
			road = filepath.Join(pwd, road)
		}
		conf.Config.RootPath = road
	}
	// 开启登录的时候查一下是否有设置账号密码
	// if c.IsSet("secure") {

	// }

	return nil
}

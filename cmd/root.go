package cmd

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
	"oneTiny/common/define"
	"oneTiny/common/config"
	"os"

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
func RunCLI() *cli.App {
	initCLI()

	return &cli.App{
		Name:            "OneTiny",
		Usage:           "一个用于局域网内共享文件的FTP程序",
		UsageText:       "onetiny [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS] [参数...]",
		Version:         define.VERSION,
		Flags:            newGlobalFlag(),
		Authors:         []*cli.Author{{Name: "Boii", Email: "i@tcp404.com"}},
		Commands:        []*cli.Command{updateCmd, configCmd, secureCmd},
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

}

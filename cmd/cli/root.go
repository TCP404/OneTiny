package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tcp404/OneTiny/internal/app/validation"
	"github.com/tcp404/OneTiny/internal/config"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/version"

	"github.com/urfave/cli/v2"
)

func initCLI(output io.Writer) {
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
	cli.ErrWriter = output
}

// CLI 函数作为程序入口，主要负责处理命令和 flag
func CLI(store *config.Store, runtimeState *runtime.Runtime) *cli.App {
	defaults := runtimeState.Snapshot()
	initCLI(defaults.Output)

	return &cli.App{
		Name:            "OneTiny",
		Usage:           "一个用于局域网内共享文件的FTP程序",
		UsageText:       "onetiny [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS] [参数...]",
		Version:         version.Version,
		Flags:           newGlobalFlag(defaults),
		Authors:         []*cli.Author{{Name: "Boii", Email: "i@tcp404.com"}},
		Commands:        []*cli.Command{updateCmd(), configCmd(store, defaults), secureCmd(store)},
		CommandNotFound: func(c *cli.Context, s string) { cli.ShowAppHelpAndExit(c, 10) },
		Writer:          defaults.Output,
		ErrWriter:       defaults.Output,
		After: func(c *cli.Context) error {
			return afterRootAction(runtimeState)
		},
		Action: func(c *cli.Context) error {
			return rootAction(c, runtimeState)
		},
	}
}

func afterRootAction(runtimeState *runtime.Runtime) error {
	snapshot := runtimeState.Snapshot()
	if err := validation.ValidatePort(snapshot.Port); err != nil {
		return err
	}
	return validation.ValidatePath(snapshot.RootPath)
}

func rootAction(c *cli.Context, runtimeState *runtime.Runtime) error {
	port := c.Int("port")
	maxLevel := uint8(c.Int("max"))
	allowUpload := c.Bool("allow")
	secure := c.Bool("secure")
	patch := runtime.Patch{
		Port:          &port,
		MaxLevel:      &maxLevel,
		IsAllowUpload: &allowUpload,
		IsSecure:      &secure,
	}
	if c.IsSet("road") {
		road := c.Path("road")
		if road[0] == '.' {
			pwd, _ := os.Getwd()
			road = filepath.Join(pwd, road)
		}
		patch.RootPath = &road
	}
	runtimeState.Update(patch)
	// 开启登录的时候查一下是否有设置账号密码
	// if c.IsSet("secure") {

	// }

	return nil
}

package cmd

import "github.com/urfave/cli/v2"

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cli.Command {
	return &cli.Command{
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
}

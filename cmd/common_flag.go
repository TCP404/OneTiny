package cmd

import (
	"strconv"

	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/urfave/cli/v2"
)

func newGlobalFlag() []cli.Flag {
	return []cli.Flag{
		&cli.PathFlag{
			Name:        "road",
			Aliases:     []string{"r"},
			Usage:       "指定对外开放的目录`路径`",
			Value:       conf.Config.RootPath,
			Required:    false,
			DefaultText: conf.Config.RootPath,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "指定开放的`端口`",
			Value:       conf.Config.Port,
			Required:    false,
			DefaultText: strconv.Itoa(conf.Config.Port),
		},
		&cli.BoolFlag{
			Name:        "allow",
			Aliases:     []string{"a"},
			Usage:       "指定`是否`允许访问者上传",
			Value:       conf.Config.IsAllowUpload,
			Required:    false,
			DefaultText: strconv.FormatBool(conf.Config.IsAllowUpload),
		},
		&cli.IntFlag{
			Name:        "max",
			Aliases:     []string{"x"},
			Usage:       "指定允许访问的`深度`，默认仅限访问共享目录",
			Value:       int(conf.Config.MaxLevel),
			Required:    false,
			DefaultText: "0",
		},
		&cli.BoolFlag{
			Name:        "secure",
			Aliases:     []string{"s"},
			Usage:       "指定是否开启访问登录",
			Value:       conf.Config.IsSecure,
			Required:    false,
			DefaultText: strconv.FormatBool(conf.Config.IsSecure),
		},
	}
}

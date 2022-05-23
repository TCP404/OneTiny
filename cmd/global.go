package cmd

import (
	"strconv"

	"github.com/TCP404/OneTiny-cli/config"

	"github.com/urfave/cli/v2"
)

func newGlobalFlag() []cli.Flag {
	return []cli.Flag{
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
}

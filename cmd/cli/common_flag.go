package main

import (
	"strconv"

	"github.com/tcp404/OneTiny/internal/state"
	"github.com/urfave/cli/v2"
)

func newGlobalFlag(defaults state.ConfigSnapshot) []cli.Flag {
	return []cli.Flag{
		&cli.PathFlag{
			Name:        "road",
			Aliases:     []string{"r"},
			Usage:       "指定对外开放的目录`路径`",
			Value:       defaults.RootPath,
			Required:    false,
			DefaultText: defaults.RootPath,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "指定开放的`端口`",
			Value:       defaults.Port,
			Required:    false,
			DefaultText: strconv.Itoa(defaults.Port),
		},
		&cli.BoolFlag{
			Name:        "allow",
			Aliases:     []string{"a"},
			Usage:       "指定`是否`允许访问者上传",
			Value:       defaults.IsAllowUpload,
			Required:    false,
			DefaultText: strconv.FormatBool(defaults.IsAllowUpload),
		},
		&cli.IntFlag{
			Name:        "max",
			Aliases:     []string{"x"},
			Usage:       "指定允许访问的`深度`，默认仅限访问共享目录",
			Value:       int(defaults.MaxLevel),
			Required:    false,
			DefaultText: "0",
		},
		&cli.BoolFlag{
			Name:        "secure",
			Aliases:     []string{"s"},
			Usage:       "指定是否开启访问登录",
			Value:       defaults.IsSecure,
			Required:    false,
			DefaultText: strconv.FormatBool(defaults.IsSecure),
		},
	}
}

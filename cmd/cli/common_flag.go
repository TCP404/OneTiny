package main

import (
	"strconv"

	defaultcfg "github.com/tcp404/OneTiny/internal/defaults"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/urfave/cli/v2"
)

func newGlobalFlag(snapshot runtime.Snapshot) []cli.Flag {
	return []cli.Flag{
		&cli.PathFlag{
			Name:        "road",
			Aliases:     []string{"r"},
			Usage:       "指定对外开放的目录`路径`",
			Value:       snapshot.RootPath,
			Required:    false,
			DefaultText: snapshot.RootPath,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "指定开放的`端口`",
			Value:       snapshot.Port,
			Required:    false,
			DefaultText: strconv.Itoa(snapshot.Port),
		},
		&cli.BoolFlag{
			Name:        "allow",
			Aliases:     []string{"a"},
			Usage:       "指定`是否`允许访问者上传",
			Value:       snapshot.IsAllowUpload,
			Required:    false,
			DefaultText: strconv.FormatBool(snapshot.IsAllowUpload),
		},
		&cli.IntFlag{
			Name:        "max",
			Aliases:     []string{"x"},
			Usage:       "指定允许访问的`深度`，默认仅限访问共享目录",
			Value:       int(snapshot.MaxLevel),
			Required:    false,
			DefaultText: "0",
		},
		&cli.BoolFlag{
			Name:        "secure",
			Aliases:     []string{"s"},
			Usage:       "指定是否开启访问登录",
			Value:       snapshot.IsSecure,
			Required:    false,
			DefaultText: strconv.FormatBool(snapshot.IsSecure),
		},
		&cli.IntFlag{
			Name:        "scratch-max-items",
			Usage:       "指定临时列表最多保留的`条目数`",
			Value:       defaultcfg.ScratchMaxItems,
			Required:    false,
			DefaultText: strconv.Itoa(defaultcfg.ScratchMaxItems),
		},
		&cli.StringFlag{
			Name:        "scratch-max-item-size",
			Usage:       "指定临时列表单条内容大小上限，例如 `10MB`",
			Value:       defaultcfg.ScratchMaxItemSize,
			Required:    false,
			DefaultText: defaultcfg.ScratchMaxItemSize,
		},
	}
}

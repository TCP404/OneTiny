package main

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/kit/chain"
	"github.com/tcp404/OneTiny/internal/kit/verify"
	"github.com/tcp404/OneTiny/internal/state"
	"github.com/urfave/cli/v2"
)

func configCmd(defaults state.ConfigSnapshot) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c", "cf", "cfg", "conf"},
		Usage:       "设置默认配置",
		UsageText:   "onetiny config [OPTIONS]",
		Description: "使用 onetiny config 命令可以将设置写入配置文件。\n使用方式与 onetiny 命令相同，仅多了一个 config 关键字，如：\n  onetiny config -p 10240  可以将端口设置为 10240 写入配置\n  onetiny config -a false  可以设置不允许访问者上传并写入配置",
		Flags:       newGlobalFlag(defaults),
		Before:      beforeConfigAction,
		Action:      configAction,
	}
}

func beforeConfigAction(c *cli.Context) error {
	verificationChain := chain.NewHandleChain()
	if c.IsSet("port") {
		verificationChain.AddToHead(verify.NewPortVerifier(c.Int("port")))
	}
	if c.IsSet("road") {
		p := c.Path("road")
		if p[0] == '.' {
			curr, _ := os.Getwd()
			p = filepath.Join(curr, p)
		}
		verificationChain.AddToHead(verify.NewPathVerifier(p))
	}
	return verificationChain.Iterator()
}

func configAction(c *cli.Context) error {
	// tiny config -p=8080 -x=3
	if c.IsSet("secure") && c.Bool("secure") {
		if err := conf.ValidateSecureConfigFor(true); err != nil {
			return cli.Exit(color.RedString(err.Error()), 11)
		}
	}
	var patch conf.ConfigPatch
	if c.IsSet("port") {
		port := c.Int("port")
		patch.Port = &port
	}
	if c.IsSet("allow") {
		allow := c.Bool("allow")
		patch.IsAllowUpload = &allow
	}
	if c.IsSet("max") {
		maxLevel := uint8(c.Int("max"))
		patch.MaxLevel = &maxLevel
	}
	if c.IsSet("road") {
		p := c.Path("road")
		if p[0] == '.' {
			curr, _ := os.Getwd()
			p = filepath.Join(curr, p)
		}
		patch.RootPath = &p
	}
	if c.IsSet("secure") {
		secure := c.Bool("secure")
		patch.IsSecure = &secure
	}
	if _, err := conf.SavePatch(patch); err != nil {
		return cli.Exit(err.Error(), 11)
	}
	return cli.Exit(color.GreenString("配置成功~"), 0)
}

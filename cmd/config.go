package cmd

import (
	"os"
	"path/filepath"

	"github.com/TCP404/OneTiny-cli/internal/kit/verify"
	"github.com/TCP404/OneTiny-cli/pkg/container"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func configCmd() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c", "cf", "cfg", "conf"},
		Usage:       "设置默认配置",
		UsageText:   "onetiny config [OPTIONS]",
		Description: "使用 onetiny config 命令可以将设置写入配置文件。\n使用方式与 onetiny 命令相同，仅多了一个 config 关键字，如：\n  onetiny config -p 10240  可以将端口设置为 10240 写入配置\n  onetiny config -a false  可以设置不允许访问者上传并写入配置",
		Flags:       newGlobalFlag(),
		Before:      beforeConfigAction,
		Action:      configAction,
	}
}

func beforeConfigAction(c *cli.Context) error {
	chain := container.NewHandleChain()
	if c.IsSet("port") {
		chain.AddToHead(verify.NewPortVerifier(c.Int("port")))
	}
	if c.IsSet("road") {
		p := c.Path("road")
		if p[0] == '.' {
			curr, _ := os.Getwd()
			p = filepath.Join(curr, p)
		}
		chain.AddToHead(verify.NewPathVerifier(p))
	}
	return chain.Iterator()
}

func configAction(c *cli.Context) error {
	// tiny config -p=8080 -x=3
	if c.IsSet("port") {
		viper.Set("server.port", c.Int("port"))
	}
	if c.IsSet("allow") {
		viper.Set("server.allow_upload", c.Bool("allow"))
	}
	if c.IsSet("max") {
		viper.Set("server.max_level", c.Int("max"))
	}
	if c.IsSet("road") {
		p := c.Path("road")
		if p[0] == '.' {
			curr, _ := os.Getwd()
			p = filepath.Join(curr, p)
		}
		viper.Set("server.road", p)
	}
	if c.IsSet("secure") {
		viper.Set("account.secure", c.Bool("secure"))
	}
	if err := viper.WriteConfig(); err != nil {
		return cli.Exit(err.Error(), 11)
	}
	return cli.Exit(color.GreenString("配置成功~"), 0)
}

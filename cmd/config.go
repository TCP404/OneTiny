package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var configCmd = newConfigCmd()

func newConfigCmd() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c", "cf", "cfg", "conf"},
		Usage:       "设置默认配置",
		UsageText:   "onetiny config [OPTIONS]",
		Description: "使用 onetiny config 命令可以将设置写入配置文件。\n使用方式与 onetiny 命令相同，仅多了一个 config 关键字，如：\n  onetiny config -p 10240  可以将端口设置为 10240 写入配置\n  onetiny config -a false  可以设置不允许访问者上传并写入配置",
		Flags:        newGlobalFlag(),
		Action: func(c *cli.Context) error {
			if err := Set(c); err != nil {
				return cli.Exit(err.Error(), 11)
			}
			return cli.Exit(color.GreenString("配置成功~"), 0)
		},
	}
}

func Set(c *cli.Context) error {
	// tiny config -p=8080 -x=3
	// TODO 增加参数检查
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
		viper.Set("server.road", c.Path("road"))
	}
	if c.IsSet("secure") {
		viper.Set("account.secure", c.Bool("secure"))
	}
	return viper.WriteConfig()
}

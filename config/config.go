package config

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

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

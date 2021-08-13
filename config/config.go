// 配置优先级： flag > 配置文件 > 默认值
package config

import (
	"io"
	"log"
	"oneTiny/util"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const (
	VERSION    string = "v0.2.3"
	ROOT       string = "/"
	SEPARATORS string = "/"
)

var (
	Goos  string = runtime.GOOS // 程序所在的操作系统，默认值 linux
	IP    string = ip           // 本机局域网IP
	wd, _        = os.Getwd()
	ip, _        = util.GetIP()

	userCfgDir, _        = os.UserConfigDir()
	cfgPath       string = filepath.Join(userCfgDir, "tiny")
	cfgFile       string = filepath.Join(cfgPath, "config.yml")

	// 各个参数的原厂默认值
	RootPath      string = wd      // 共享目录的根路径，默认值：当前目录
	Port          int    = 9090    // 指定的服务端口，默认值 9090
	MaxLevel      uint8  = 0       // 允许访问的最大层级，默认值  0
	IsAllowUpload bool   = false   // 是否允许上传，默认值：否
	Username      string = "admin" // 访问登录的帐号
	Password      string = "admin" // 访问登录的密码
)

var (
	CurrPath string // 上传目录时的当前路径
	Output   io.Writer
)

func init() {
	loadUserConfig()

	Port = viper.GetInt("server.port")
	MaxLevel = uint8(viper.GetInt("server.max_level"))
	IsAllowUpload = viper.GetBool("server.allow_upload")
	RootPath = strings.TrimSuffix(RootPath, SEPARATORS)
	if Goos == "windows" {
		Output = color.Output
	} else {
		Output = os.Stderr
	}
}

func loadUserConfig() {
	viper.AddConfigPath(cfgPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

read:
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			createCfgFile()
			goto read
		case viper.ConfigParseError:
			log.Println(color.RedString("已找到「自定义配置文件」，但是解析失败！"))
		case viper.ConfigMarshalError:
			log.Println(color.RedString("已找到「自定义配置文件」，但是读取失败！"))
		}
	}
}

func setDefault() error {
	viper.Set("server.port", 9090)
	viper.Set("server.road", wd)
	viper.Set("server.allow_upload", false)
	viper.Set("server.max_level", 0)
	return viper.WriteConfig()
}

func createCfgFile() {
	log.Println(color.YellowString("未找到「自定义配置文件」, 正在创建中..."))
	_, err := os.Stat(cfgPath)
	if os.IsNotExist(err) {
		os.MkdirAll(cfgPath, os.ModePerm)
	}
	_, err = os.Create(cfgFile)
	if err != nil {
		log.Println(color.RedString("创建自定义配置文件失败！"))
	}
	if err := setDefault(); err == nil {
		log.Println(color.GreenString("创建成功，配置文件位于: %s", cfgFile))
	}
}

func Set(c *cli.Context) error {
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
		viper.Set("server.road", c.Path("road"))
	}
	return viper.WriteConfig()
}

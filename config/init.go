// 配置优先级： flag > 配置文件 > 默认值
package config

import (
	"io"
	"log"
	"oneTiny/core/util"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

const (
	VERSION    string = "v0.2.3"
	ROOT       string = "/"
	SEPARATORS string = "/"
)

var (
	Output     io.Writer = os.Stderr
	SessionVal string    = util.RandString(64)
	Goos       string    = runtime.GOOS // 程序所在的操作系统，默认值 linux
	IP         string    = ip           // 本机局域网IP
	wd, _                = os.Getwd()
	ip, _                = util.GetIP()

	userCfgDir, _        = os.UserConfigDir()
	cfgDir        string = filepath.Join(userCfgDir, "tiny")
	cfgFile       string = filepath.Join(cfgDir, "config.yml")

	// 各个参数的原厂默认值
	MaxLevel      uint8  = 0       // 允许访问的最大层级，默认值  0
	Port          int    = 9090    // 指定的服务端口，默认值 9090
	IsAllowUpload bool   = false   // 是否允许上传，默认值：否
	IsSecure      bool   = false   // 是否开启访问登录，默认值：否
	RootPath      string = wd      // 共享目录的根路径，默认值：当前目录
	Username      string = "admin" // 访问登录的帐号
	Password      string = "admin" // 访问登录的密码
)

func init() {
	loadUserConfig()

	MaxLevel = uint8(viper.GetInt("server.max_level"))
	Port = viper.GetInt("server.port")
	IsAllowUpload = viper.GetBool("server.allow_upload")
	IsSecure = viper.GetBool("account.secure")

	if Goos == "windows" {
		Output = color.Output
	}
}

// loadUserConfig 负责加载用户配置文件,如果文件不存在则创建并设置默认值
func loadUserConfig() {
	viper.AddConfigPath(cfgDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

read:
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Println(color.YellowString("未找到「自定义配置文件」, 正在创建中..."))
			createCfgFile()
			goto read
		case viper.ConfigParseError:
			log.Println(color.RedString("已找到「自定义配置文件」，但是解析失败！"))
		case viper.ConfigMarshalError:
			log.Println(color.RedString("已找到「自定义配置文件」，但是读取失败！"))
		}
	}
}

func createCfgFile() {
	_, err := os.Stat(cfgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(cfgDir, os.ModePerm)
	}
	_, err = os.Create(cfgFile)
	if err != nil {
		log.Println(color.RedString("创建自定义配置文件失败！"))
	}
	if err := setDefault(); err == nil {
		log.Println(color.GreenString("创建成功，配置文件位于: %s", cfgFile))
	}
}

func setDefault() error {
	viper.Set("server.port", 9090)
	viper.Set("server.road", wd)
	viper.Set("server.allow_upload", false)
	viper.Set("server.max_level", 0)
	viper.Set("account.default.user", util.MD5("admin"))
	viper.Set("account.default.pass", util.MD5("admin"))
	return viper.WriteConfig()
}

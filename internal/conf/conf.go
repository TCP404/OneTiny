package conf

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/TCP404/OneTiny-cli/internal/constant"
	"github.com/TCP404/eutil"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var Config = &config{
	Output: eutil.If[io.Writer](runtime.GOOS == "windows", color.Output, os.Stderr),
	OS:     runtime.GOOS, // 程序所在的操作系统，默认值 linux
	IP:     constant.IP,  // 本机局域网IP
	Pwd:    "/",

	RootPath:      constant.RootPath,      // 共享目录的根路径，默认值：当前目录
	MaxLevel:      constant.MaxLevel,      // 允许访问的最大层级，默认值  0
	Port:          constant.Port,          // 指定的服务端口，默认值 9090
	IsAllowUpload: constant.IsAllowUpload, // 是否允许上传，默认值：否
	IsSecure:      constant.IsSecure,      // 是否开启访问登录，默认值：否

	SessionVal: eutil.RandString(64),
	Username:   constant.Username, // 访问登录的帐号，默认值：admin
	Password:   constant.Password, // 访问登录的密码，默认值：admin
}

type config struct {
	Output io.Writer
	OS     string
	IP     string
	Pwd    string

	RootPath      string // 共享目录的根路径，默认值：当前目录
	MaxLevel      uint8  // 允许访问的最大层级，默认值  0
	Port          int    // 指定的服务端口，默认值 9090
	IsAllowUpload bool   // 是否允许上传，默认值：否
	IsSecure      bool   // 是否开启访问登录，默认值：否

	SessionVal string
	Username   string // 访问登录的帐号
	Password   string // 访问登录的密码
}

func LoadConfig() error {
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return errors.New("获取配置目录失败")
	}
	cfgDir := filepath.Join(userCfgDir, "tiny")
	cfgFile := filepath.Join(cfgDir, "config.yml")
	err = loadUserConfig(cfgDir, cfgFile)
	if err != nil {
		return err
	}

	Config.IP, err = eutil.GetLocalIP()
	if err != nil {
		log.Println(color.YellowString("获取不到本机的局域网IP"))
	}

	Config.RootPath = viper.GetString("server.road")
	if len(Config.RootPath) < 1 {
		wd, err := os.Getwd()
		if err != nil {
			return errors.New("获取不到共享路径")
		}
		Config.RootPath = wd
		Config.Pwd = wd
	}

	Config.MaxLevel = uint8(viper.GetInt("server.max_level"))
	Config.Port = viper.GetInt("server.port")
	Config.IsAllowUpload = viper.GetBool("server.allow_upload")
	Config.IsSecure = viper.GetBool("account.secure")
	Config.Username = viper.GetString("account.custom.user")
	Config.Password = viper.GetString("account.custom.pass")
	return nil
}

// loadUserConfig 负责加载用户配置文件,如果文件不存在则创建并设置默认值
func loadUserConfig(cfgDir, cfgFile string) error {
	viper.AddConfigPath(cfgDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

read:
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Println(color.YellowString("未找到「自定义配置文件」, 正在创建中..."))
			if err := createCfgFile(cfgDir, cfgFile); err != nil {
				return err
			}
			log.Println(color.GreenString("创建成功，配置文件位于: %s", cfgFile))
			goto read
		case viper.ConfigParseError:
			return errors.New("已找到「自定义配置文件」，但是解析失败！")
		case viper.ConfigMarshalError:
			return errors.New("已找到「自定义配置文件」，但是读取失败！")
		}
	}
	return nil
}

func createCfgFile(cfgDir, cfgFile string) error {
	_, err := os.Stat(cfgDir)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(cfgDir, os.ModePerm)
	}
	_, err = os.Create(cfgFile)
	if err != nil {
		return errors.New("创建自定义配置文件失败！")
	}
	if err := setDefault(); err != nil {
		log.Println(color.YellowString("设置默认配置失败！"))
	}
	return nil
}

func setDefault() error {
	viper.Set("server.port", constant.Port)
	viper.Set("server.allow_upload", constant.IsAllowUpload)
	viper.Set("server.max_level", constant.MaxLevel)
	return viper.WriteConfig()
}

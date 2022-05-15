package config

import (
	"errors"
	"io"
	"log"
	"oneTiny/common"
	"oneTiny/common/define"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

// 配置优先级： flag > 配置文件 > 默认值

var (
	Output     io.Writer = os.Stderr
	SessionVal string    = common.RandString(64)
	Goos       string    = runtime.GOOS // 程序所在的操作系统，默认值 linux
	IP         string    = define.IP    // 本机局域网IP

	// 各个参数的原厂默认值
	RootPath      string = define.RootPath      // 共享目录的根路径，默认值：当前目录
	MaxLevel      uint8  = define.MaxLevel      // 允许访问的最大层级，默认值  0
	Port          int    = define.Port          // 指定的服务端口，默认值 9090
	IsAllowUpload bool   = define.IsAllowUpload // 是否允许上传，默认值：否
	IsSecure      bool   = define.IsSecure      // 是否开启访问登录，默认值：否
	Username      string = define.Username      // 访问登录的帐号
	Password      string = define.Password      // 访问登录的密码
)

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

	IP, err = common.GetIP()
	if err != nil {
		log.Println(color.YellowString("获取不到本机的局域网IP"))
	}

	RootPath = viper.GetString("server.road")
	if len(RootPath) < 1 {
		RootPath, err = os.Getwd()
		if err != nil {
			return errors.New("获取不到共享路径")
		}
	}

	MaxLevel = uint8(viper.GetInt("server.max_level"))
	Port = viper.GetInt("server.port")
	IsAllowUpload = viper.GetBool("server.allow_upload")
	IsSecure = viper.GetBool("account.secure")

	if Goos == "windows" {
		Output = color.Output
	}
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
		os.MkdirAll(cfgDir, os.ModePerm)
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
	viper.Set("server.port", 9090)
	viper.Set("server.allow_upload", false)
	viper.Set("server.max_level", 0)
	return viper.WriteConfig()
}

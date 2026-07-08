package conf

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/constant"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RootPath      string // 共享目录的根路径，默认值：当前目录
	MaxLevel      uint8  // 允许访问的最大层级，默认值  0
	Port          int    // 指定的服务端口，默认值 9090
	IsAllowUpload bool   // 是否允许上传，默认值：否
	IsSecure      bool   // 是否开启访问登录，默认值：否

	Username         string // 访问登录的帐号
	PasswordHash     string // 访问登录的密码 hash
	PasswordHashAlgo string // 访问登录密码 hash 算法
	LegacyPassword   string // 旧版 MD5 密码，仅用于兼容读取和迁移
}

type ConfigPatch struct {
	RootPath      *string
	MaxLevel      *uint8
	Port          *int
	IsAllowUpload *bool
	IsSecure      *bool
}

var currentConfig Config
var writeConfigFile = atomicWriteFile

func Current() Config {
	return currentConfig
}

// UnsafeCurrentForTest exposes the in-memory config cache for legacy tests only.
// Production code must use LoadConfig, Current, SavePatch, or SaveSecurityPatch.
func UnsafeCurrentForTest() *Config {
	return &currentConfig
}

func SetWriteConfigFileForTest(fn func(path string, data []byte) error) func() {
	previous := writeConfigFile
	writeConfigFile = fn
	return func() {
		writeConfigFile = previous
	}
}

func RestoreInMemory(cfg Config) {
	setConfigValues(cfg)
	currentConfig = cfg
}

func RefreshCurrent() (Config, error) {
	cfg, err := configFromViper()
	if err != nil {
		return Config{}, err
	}
	currentConfig = cfg
	return cfg, nil
}

func ConfigDir() (string, error) {
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.New("获取配置目录失败")
	}
	return filepath.Join(userCfgDir, "tiny"), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

func LoadConfig() (Config, error) {
	cfgDir, err := ConfigDir()
	if err != nil {
		return Config{}, err
	}
	cfgFile, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	err = loadUserConfig(cfgDir, cfgFile)
	if err != nil {
		return Config{}, err
	}

	cfg, err := configFromViper()
	if err != nil {
		return Config{}, err
	}

	currentConfig = cfg
	return cfg, nil
}

func SavePatch(patch ConfigPatch) (Config, error) {
	rollback := captureViperKeys(
		"server.road",
		"server.port",
		"server.max_level",
		"server.allow_upload",
		"account.secure",
	)
	originalConfig := currentConfig

	ensureCurrentConfigFromViper()
	setConfigValues(currentConfig)
	if patch.IsSecure != nil && *patch.IsSecure {
		if err := ValidateSecureConfigFor(true); err != nil {
			return Config{}, err
		}
	}
	if patch.RootPath != nil {
		viper.Set("server.road", *patch.RootPath)
	}
	if patch.Port != nil {
		viper.Set("server.port", *patch.Port)
	}
	if patch.MaxLevel != nil {
		viper.Set("server.max_level", int(*patch.MaxLevel))
	}
	if patch.IsAllowUpload != nil {
		viper.Set("server.allow_upload", *patch.IsAllowUpload)
	}
	if patch.IsSecure != nil {
		viper.Set("account.secure", *patch.IsSecure)
	}
	if err := writeCurrentViperConfigAtomic(); err != nil {
		rollback.restore()
		currentConfig = originalConfig
		return Config{}, err
	}
	cfg, err := LoadConfig()
	if err != nil {
		rollback.restore()
		currentConfig = originalConfig
		return Config{}, err
	}
	return cfg, nil
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

func setConfigValues(cfg Config) {
	viper.Set("server.road", cfg.RootPath)
	viper.Set("server.port", cfg.Port)
	viper.Set("server.max_level", int(cfg.MaxLevel))
	viper.Set("server.allow_upload", cfg.IsAllowUpload)
	viper.Set("account.secure", cfg.IsSecure)
	viper.Set("account.custom.user", cfg.Username)
	viper.Set("account.custom.pass_hash", cfg.PasswordHash)
	viper.Set("account.custom.pass_hash_algo", cfg.PasswordHashAlgo)
	viper.Set("account.custom.pass", cfg.LegacyPassword)
}

func ensureCurrentConfigFromViper() {
	if currentConfig != (Config{}) {
		return
	}
	cfg, err := configFromViper()
	if err != nil {
		return
	}
	currentConfig = cfg
}

func configFromViper() (Config, error) {
	cfg := Config{
		RootPath:         viper.GetString("server.road"),
		MaxLevel:         uint8(viper.GetInt("server.max_level")),
		Port:             viper.GetInt("server.port"),
		IsAllowUpload:    viper.GetBool("server.allow_upload"),
		IsSecure:         viper.GetBool("account.secure"),
		Username:         viper.GetString("account.custom.user"),
		PasswordHash:     viper.GetString("account.custom.pass_hash"),
		PasswordHashAlgo: viper.GetString("account.custom.pass_hash_algo"),
		LegacyPassword:   viper.GetString("account.custom.pass"),
	}
	if len(cfg.RootPath) < 1 {
		wd, err := os.Getwd()
		if err != nil {
			return Config{}, errors.New("获取不到共享路径")
		}
		cfg.RootPath = wd
	}
	return cfg, nil
}

type viperRollback struct {
	values map[string]any
}

func captureViperKeys(keys ...string) viperRollback {
	values := make(map[string]any, len(keys))
	for _, key := range keys {
		values[key] = viper.Get(key)
	}
	return viperRollback{values: values}
}

func (r viperRollback) restore() {
	for key, value := range r.values {
		viper.Set(key, value)
	}
}

func writeCurrentViperConfigAtomic() error {
	path := viper.ConfigFileUsed()
	if strings.TrimSpace(path) == "" {
		var err error
		path, err = ConfigPath()
		if err != nil {
			return err
		}
	}
	data, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}
	return writeConfigFile(path, data)
}

func atomicWriteFile(path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("配置文件路径为空")
	}
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	removeTemp = false
	if dirFile, err := os.Open(dir); err == nil {
		_ = dirFile.Sync()
		_ = dirFile.Close()
	}
	return nil
}

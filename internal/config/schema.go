package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/defaults"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RootPath           string
	MaxLevel           uint8
	Port               int
	IsAllowUpload      bool
	IsSecure           bool
	ScratchMaxItems    int
	ScratchMaxItemSize string

	Username         string
	PasswordHash     string
	PasswordHashAlgo string
	LegacyPassword   string
}

type ConfigPatch struct {
	RootPath           *string
	MaxLevel           *uint8
	Port               *int
	IsAllowUpload      *bool
	IsSecure           *bool
	ScratchMaxItems    *int
	ScratchMaxItemSize *string
}

type Store struct {
	path      string
	v         *viper.Viper
	current   Config
	writeFile func(path string, data []byte) error
}

func DefaultDir() (string, error) {
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.New("获取配置目录失败")
	}
	return filepath.Join(userCfgDir, "tiny"), nil
}

func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

func NewStore(path string) *Store {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yml")
	return &Store{
		path:      path,
		v:         v,
		writeFile: atomicWriteFile,
	}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Current() Config {
	return s.current
}

func (s *Store) Load() (Config, error) {
	if strings.TrimSpace(s.path) == "" {
		return Config{}, errors.New("配置文件路径为空")
	}
	if err := s.ensureFile(); err != nil {
		return Config{}, err
	}
	if err := s.v.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigParseError:
			return Config{}, errors.New("已找到「自定义配置文件」，但是解析失败！")
		case viper.ConfigMarshalError:
			return Config{}, errors.New("已找到「自定义配置文件」，但是读取失败！")
		default:
			return Config{}, err
		}
	}
	cfg, err := s.configFromViper()
	if err != nil {
		return Config{}, err
	}
	s.current = cfg
	return cfg, nil
}

func (s *Store) Patch(patch ConfigPatch) (Config, error) {
	cfg, err := s.ensureCurrent()
	if err != nil {
		return Config{}, err
	}
	if patch.RootPath != nil {
		cfg.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		cfg.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		cfg.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		cfg.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		cfg.IsSecure = *patch.IsSecure
	}
	if patch.ScratchMaxItems != nil {
		cfg.ScratchMaxItems = *patch.ScratchMaxItems
	}
	if patch.ScratchMaxItemSize != nil {
		cfg.ScratchMaxItemSize = *patch.ScratchMaxItemSize
	}
	if err := validateScratchConfig(cfg); err != nil {
		return Config{}, err
	}
	if err := validateSecureConfigFor(cfg, cfg.IsSecure); err != nil {
		return Config{}, err
	}
	if err := s.writeConfig(cfg); err != nil {
		return Config{}, err
	}
	return s.Load()
}

func (s *Store) ensureCurrent() (Config, error) {
	if s.current != (Config{}) {
		return s.current, nil
	}
	return s.Load()
}

func (s *Store) ensureFile() error {
	if _, err := os.Stat(s.path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return s.writeConfig(defaultConfig())
}

func (s *Store) configFromViper() (Config, error) {
	cfg := Config{
		RootPath:        s.v.GetString("server.road"),
		MaxLevel:        uint8(s.v.GetInt("server.max_level")),
		Port:            s.v.GetInt("server.port"),
		IsAllowUpload:   s.v.GetBool("server.allow_upload"),
		IsSecure:        s.v.GetBool("account.secure"),
		ScratchMaxItems: s.v.GetInt("scratch.max_items"),
		ScratchMaxItemSize: strings.TrimSpace(
			s.v.GetString("scratch.max_item_size"),
		),
		Username:         s.v.GetString("account.custom.user"),
		PasswordHash:     s.v.GetString("account.custom.pass_hash"),
		PasswordHashAlgo: s.v.GetString("account.custom.pass_hash_algo"),
		LegacyPassword:   s.v.GetString("account.custom.pass"),
	}
	if cfg.Port == 0 {
		cfg.Port = defaults.Port
	}
	if strings.TrimSpace(cfg.RootPath) == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Config{}, errors.New("获取不到共享路径")
		}
		cfg.RootPath = wd
	}
	if !s.v.IsSet("scratch.max_items") {
		cfg.ScratchMaxItems = defaults.ScratchMaxItems
	}
	if !s.v.IsSet("scratch.max_item_size") {
		cfg.ScratchMaxItemSize = defaults.ScratchMaxItemSize
	}
	if err := validateScratchConfig(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *Store) writeConfig(cfg Config) error {
	data, err := yaml.Marshal(configSettings(cfg))
	if err != nil {
		return err
	}
	return s.writeFile(s.path, data)
}

func defaultConfig() Config {
	return Config{
		RootPath:           defaults.RootPath,
		MaxLevel:           defaults.MaxLevel,
		Port:               defaults.Port,
		IsAllowUpload:      defaults.IsAllowUpload,
		IsSecure:           defaults.IsSecure,
		ScratchMaxItems:    defaults.ScratchMaxItems,
		ScratchMaxItemSize: defaults.ScratchMaxItemSize,
	}
}

func configSettings(cfg Config) map[string]any {
	return map[string]any{
		"server": map[string]any{
			"road":         cfg.RootPath,
			"port":         cfg.Port,
			"max_level":    int(cfg.MaxLevel),
			"allow_upload": cfg.IsAllowUpload,
		},
		"account": map[string]any{
			"secure": cfg.IsSecure,
			"custom": map[string]any{
				"user":           cfg.Username,
				"pass_hash":      cfg.PasswordHash,
				"pass_hash_algo": cfg.PasswordHashAlgo,
				"pass":           cfg.LegacyPassword,
			},
		},
		"scratch": map[string]any{
			"max_items":     cfg.ScratchMaxItems,
			"max_item_size": cfg.ScratchMaxItemSize,
		},
	}
}

func validateScratchConfig(cfg Config) error {
	if cfg.ScratchMaxItems < 1 {
		return errors.New("临时列表容量必须大于 0")
	}
	sizeBytes, err := ParseByteSize(cfg.ScratchMaxItemSize)
	if err != nil || sizeBytes < 1 {
		return ErrInvalidByteSize
	}
	return nil
}

func atomicWriteFile(path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("配置文件路径为空")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
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

package app

import (
	"errors"
	"time"
)

var (
	ErrPasswordConfirmationMismatch = errors.New("两次输入的密码不一致")
	ErrPortRestartRequiresConfirm   = errors.New("修改端口需要确认并重启服务")
	ErrUsernameRequired             = errors.New("用户名不能为空")
	ErrPasswordRequired             = errors.New("密码不能为空")
	ErrInvalidExportPath            = errors.New("导出路径无效")
)

type ConfigDTO struct {
	RootPath      string `json:"rootPath"`
	Port          int    `json:"port"`
	MaxLevel      uint8  `json:"maxLevel"`
	IsAllowUpload bool   `json:"isAllowUpload"`
	IsSecure      bool   `json:"isSecure"`
}

type StatusDTO struct {
	Running             bool      `json:"running"`
	StateLabel          string    `json:"stateLabel"`
	Address             string    `json:"address"`
	Config              ConfigDTO `json:"config"`
	HasCredentials      bool      `json:"hasCredentials"`
	ConfigPath          string    `json:"configPath"`
	AccessLogPath       string    `json:"accessLogPath"`
	PortRestartRequired bool      `json:"portRestartRequired"`
	LastError           string    `json:"lastError"`
}

type ConfigPatchDTO struct {
	RootPath      *string `json:"rootPath,omitempty"`
	Port          *int    `json:"port,omitempty"`
	MaxLevel      *uint8  `json:"maxLevel,omitempty"`
	IsAllowUpload *bool   `json:"isAllowUpload,omitempty"`
	IsSecure      *bool   `json:"isSecure,omitempty"`
	RestartPort   bool    `json:"restartPort,omitempty"`
}

type CredentialPatchDTO struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	EnableSecure    bool   `json:"enableSecure"`
}

type LogFilterDTO struct {
	Event string     `json:"event,omitempty"`
	Since *time.Time `json:"since,omitempty"`
	Until *time.Time `json:"until,omitempty"`
}

type LogEntryDTO struct {
	Time     time.Time `json:"time"`
	ClientIP string    `json:"clientIP"`
	Method   string    `json:"method"`
	Event    string    `json:"event"`
	Path     string    `json:"path"`
	Status   int       `json:"status"`
	Result   string    `json:"result"`
}

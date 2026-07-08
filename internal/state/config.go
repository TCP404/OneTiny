package state

import (
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/tcp404/eutil"

	"github.com/tcp404/OneTiny/internal/conf"
)

const ContextKey = "runtimeConfigSnapshot"

type ConfigSnapshot struct {
	RootPath      string
	Port          int
	MaxLevel      uint8
	IsAllowUpload bool
	IsSecure      bool
	IP            string
	Username      string
	PasswordHash  string
	SessionVal    string
	Output        io.Writer
	OS            string
	Pwd           string
}

type ConfigPatch struct {
	RootPath      *string
	Port          *int
	MaxLevel      *uint8
	IsAllowUpload *bool
	IsSecure      *bool
	Username      *string
	PasswordHash  *string
	SessionVal    *string
}

type ProcessState struct {
	Output     io.Writer
	OS         string
	IP         string
	Pwd        string
	SessionVal string
}

type RuntimeConfig struct {
	mu       sync.RWMutex
	snapshot ConfigSnapshot
}

var current atomic.Pointer[RuntimeConfig]
var fallbackProcess = NewProcessState()

func NewRuntimeConfig(snapshot ConfigSnapshot) *RuntimeConfig {
	return &RuntimeConfig{snapshot: snapshot}
}

func NewProcessState() ProcessState {
	pwd, _ := os.Getwd()
	ip, _ := eutil.GetLocalIP()
	return ProcessState{
		Output:     eutil.If[io.Writer](runtime.GOOS == "windows", color.Output, os.Stderr),
		OS:         runtime.GOOS,
		IP:         ip,
		Pwd:        pwd,
		SessionVal: eutil.RandString(64),
	}
}

func SnapshotFromConfig(cfg conf.Config, process ProcessState) ConfigSnapshot {
	return ConfigSnapshot{
		RootPath:      cfg.RootPath,
		MaxLevel:      cfg.MaxLevel,
		Port:          cfg.Port,
		IsAllowUpload: cfg.IsAllowUpload,
		IsSecure:      cfg.IsSecure,
		Username:      cfg.Username,
		PasswordHash:  cfg.PasswordHash,
		Output:        process.Output,
		OS:            process.OS,
		IP:            process.IP,
		Pwd:           process.Pwd,
		SessionVal:    process.SessionVal,
	}
}

func SnapshotFromCurrentConfig() ConfigSnapshot {
	return SnapshotFromConfig(conf.Current(), fallbackProcess)
}

func ProcessStateFromSnapshot(snapshot ConfigSnapshot) ProcessState {
	return ProcessState{
		Output:     snapshot.Output,
		OS:         snapshot.OS,
		IP:         snapshot.IP,
		Pwd:        snapshot.Pwd,
		SessionVal: snapshot.SessionVal,
	}
}

func Current() *RuntimeConfig {
	return current.Load()
}

func SetCurrent(cfg *RuntimeConfig) {
	current.Store(cfg)
}

func (c *RuntimeConfig) Snapshot() ConfigSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.snapshot
}

func (c *RuntimeConfig) Update(patch ConfigPatch) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if patch.RootPath != nil {
		c.snapshot.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		c.snapshot.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		c.snapshot.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		c.snapshot.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		c.snapshot.IsSecure = *patch.IsSecure
	}
	if patch.Username != nil {
		c.snapshot.Username = *patch.Username
	}
	if patch.PasswordHash != nil {
		c.snapshot.PasswordHash = *patch.PasswordHash
	}
	if patch.SessionVal != nil {
		c.snapshot.SessionVal = *patch.SessionVal
	}
}

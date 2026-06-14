package runtimeconf

import (
	"sync"
	"sync/atomic"
)

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

type RuntimeConfig struct {
	mu       sync.RWMutex
	snapshot ConfigSnapshot
}

var current atomic.Pointer[RuntimeConfig]

func NewRuntimeConfig(snapshot ConfigSnapshot) *RuntimeConfig {
	return &RuntimeConfig{snapshot: snapshot}
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

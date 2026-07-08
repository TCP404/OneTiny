package control

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/tcp404/OneTiny/internal/server"
	"gopkg.in/yaml.v3"
)

type Controller struct {
	mu                  sync.Mutex
	cfg                 *runtimeconf.RuntimeConfig
	manager             *server.ServiceManager
	logger              *accesslog.Logger
	lastErr             string
	portRestartRequired bool
	pendingPort         *int
}

var (
	restartServiceWithSnapshot = func(manager *server.ServiceManager, snapshot runtimeconf.ConfigSnapshot, commit func() error) error {
		return manager.RestartWithSnapshot(snapshot, commit)
	}
	atomicWriteConfigFile = atomicWriteFile
)

func NewController() *Controller {
	cfg := runtimeconf.NewRuntimeConfig(snapshotFromConf())
	return &Controller{
		cfg:     cfg,
		manager: server.NewServiceManager(cfg),
		logger:  accesslog.New(accesslog.DefaultPath()),
	}
}

func (c *Controller) GetStatus() (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statusLocked(), nil
}

func (c *Controller) StartSharing() (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := conf.ValidateSecureConfigFor(c.cfg.Snapshot().IsSecure); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}
	if err := c.manager.Start(); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}
	c.lastErr = ""
	return c.statusLocked(), nil
}

func (c *Controller) StopSharing() (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.manager.Stop(); err != nil && !errors.Is(err, server.ErrServerNotRunning) {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}
	runtimeconf.SetCurrent(nil)
	if c.portRestartRequired || c.pendingPort != nil {
		port := conf.Config.Port
		if err := c.manager.ApplyRuntimeConfig(runtimeconf.ConfigPatch{Port: &port}); err != nil {
			c.lastErr = err.Error()
			return c.statusLocked(), err
		}
		c.pendingPort = nil
		c.portRestartRequired = false
	}
	c.lastErr = ""
	return c.statusLocked(), nil
}

func (c *Controller) UpdateConfig(patch ConfigPatchDTO) (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	active := c.cfg.Snapshot()
	running := c.manager.Running()
	targetPort, hasPortTarget := c.portTargetLocked(patch, active)
	confirmPort := running && patch.RestartPort && hasPortTarget
	portChanged := hasPortTarget && targetPort != active.Port

	persistPatch := patch
	if confirmPort {
		persistPatch.Port = &targetPort
	}
	if confirmPort && portChanged {
		targetSnapshot := snapshotWithPatch(active, persistPatch)
		if err := restartServiceWithSnapshot(c.manager, targetSnapshot, func() error {
			return persistConfigPatch(persistPatch)
		}); err != nil {
			c.resetPendingPortFailureLocked(active.Port)
			c.lastErr = err.Error()
			return c.statusLocked(), err
		}
		c.pendingPort = nil
		c.portRestartRequired = false
		c.lastErr = ""
		return c.statusLocked(), nil
	}

	if err := persistConfigPatch(persistPatch); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}

	runtimePatch := runtimePatchFromDTO(persistPatch)
	if running && hasPortTarget && !confirmPort {
		runtimePatch.Port = nil
	}
	if err := c.manager.ApplyRuntimeConfig(runtimePatch); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}

	if running && hasPortTarget && portChanged {
		c.pendingPort = intPtr(targetPort)
		c.portRestartRequired = true
		c.lastErr = ErrPortRestartRequiresConfirm.Error()
		return c.statusLocked(), nil
	}
	if hasPortTarget && (!running || !portChanged) {
		c.pendingPort = nil
		c.portRestartRequired = false
	}
	if confirmPort {
		c.pendingPort = nil
		c.portRestartRequired = false
	}

	c.lastErr = ""
	return c.statusLocked(), nil
}

func (c *Controller) SetCredentials(patch CredentialPatchDTO) (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	username := strings.TrimSpace(patch.Username)
	if username == "" {
		c.lastErr = ErrUsernameRequired.Error()
		return c.statusLocked(), ErrUsernameRequired
	}
	if strings.TrimSpace(patch.Password) == "" {
		c.lastErr = ErrPasswordRequired.Error()
		return c.statusLocked(), ErrPasswordRequired
	}
	if patch.Password != patch.ConfirmPassword {
		c.lastErr = ErrPasswordConfirmationMismatch.Error()
		return c.statusLocked(), ErrPasswordConfirmationMismatch
	}
	hash, err := security.HashPassword(patch.Password)
	if err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}
	credentials := security.CredentialConfig{
		Username:     username,
		PasswordHash: hash,
		HashAlgo:     security.HashAlgoBcrypt,
	}
	if err := credentials.ValidateForSecureMode(); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}

	rollback := captureViperKeys("account.secure", "account.custom.user", "account.custom.pass_hash", "account.custom.pass_hash_algo", "account.custom.pass")
	conf.SetCredentialConfig(username, hash)
	if patch.EnableSecure {
		viper.Set("account.secure", true)
	}
	if err := writeCurrentViperConfigAtomic(); err != nil {
		rollback.restore()
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}

	conf.Config.Username = username
	conf.Config.Password = hash
	if patch.EnableSecure {
		conf.Config.IsSecure = true
	}
	secure := conf.Config.IsSecure
	if err := c.manager.ApplyRuntimeConfig(runtimeconf.ConfigPatch{
		Username:     &username,
		PasswordHash: &hash,
		IsSecure:     &secure,
	}); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}

	c.lastErr = ""
	return c.statusLocked(), nil
}

func (c *Controller) GetLogs(filter LogFilterDTO) ([]LogEntryDTO, error) {
	events, err := c.logger.Read(accesslog.Filter{
		Event: filter.Event,
		Since: timeValue(filter.Since),
		Until: timeValue(filter.Until),
	})
	if err != nil {
		return nil, err
	}
	return logEntriesFromEvents(events), nil
}

func (c *Controller) ClearLogs() error {
	return c.logger.Clear()
}

func (c *Controller) ExportLogs(path string, filter LogFilterDTO) error {
	if strings.TrimSpace(path) == "" {
		return ErrInvalidExportPath
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return ErrInvalidExportPath
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	entries, err := c.GetLogs(filter)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"time", "client_ip", "method", "event", "path", "status", "result"}); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := writer.Write([]string{
			entry.Time.Format(time.RFC3339),
			entry.ClientIP,
			entry.Method,
			entry.Event,
			entry.Path,
			strconv.Itoa(entry.Status),
			entry.Result,
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func (c *Controller) statusLocked() StatusDTO {
	snapshot := c.cfg.Snapshot()
	configPath, _ := conf.ConfigPath()
	running := c.manager.Running()
	state := "未运行"
	if running {
		state = "运行中"
	}

	return StatusDTO{
		Running:             running,
		StateLabel:          state,
		Address:             addressFromSnapshot(snapshot, running),
		Config:              c.configDTOLocked(snapshot),
		HasCredentials:      hasCredentials(snapshot),
		ConfigPath:          configPath,
		AccessLogPath:       accesslog.DefaultPath(),
		PortRestartRequired: c.portRestartRequired,
		LastError:           c.lastErr,
	}
}

func (c *Controller) portTargetLocked(patch ConfigPatchDTO, active runtimeconf.ConfigSnapshot) (int, bool) {
	if patch.RestartPort && c.pendingPort != nil {
		return *c.pendingPort, true
	}
	if patch.RestartPort && c.portRestartRequired && conf.Config.Port != active.Port {
		return conf.Config.Port, true
	}
	if patch.Port != nil {
		return *patch.Port, true
	}
	return 0, false
}

func (c *Controller) configDTOLocked(snapshot runtimeconf.ConfigSnapshot) ConfigDTO {
	dto := configDTOFromSnapshot(snapshot)
	if c.pendingPort != nil {
		dto.Port = *c.pendingPort
	}
	return dto
}

func (c *Controller) resetPendingPortFailureLocked(activePort int) {
	viper.Set("server.port", activePort)
	conf.Config.Port = activePort
	c.pendingPort = nil
	c.portRestartRequired = true
}

func snapshotFromConf() runtimeconf.ConfigSnapshot {
	return runtimeconf.ConfigSnapshot{
		RootPath:      conf.Config.RootPath,
		Port:          conf.Config.Port,
		MaxLevel:      conf.Config.MaxLevel,
		IsAllowUpload: conf.Config.IsAllowUpload,
		IsSecure:      conf.Config.IsSecure,
		IP:            conf.Config.IP,
		Username:      conf.Config.Username,
		PasswordHash:  conf.Config.Password,
		SessionVal:    conf.Config.SessionVal,
	}
}

func persistConfigPatch(patch ConfigPatchDTO) error {
	rollback := captureViperKeys(
		"server.road",
		"server.port",
		"server.max_level",
		"server.allow_upload",
		"account.secure",
	)
	originalConfig := *conf.Config

	if patch.IsSecure != nil && *patch.IsSecure {
		if err := conf.ValidateSecureConfigFor(true); err != nil {
			return err
		}
	}
	if patch.RootPath != nil {
		viper.Set("server.road", *patch.RootPath)
		conf.Config.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		viper.Set("server.port", *patch.Port)
		conf.Config.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		viper.Set("server.max_level", int(*patch.MaxLevel))
		conf.Config.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		viper.Set("server.allow_upload", *patch.IsAllowUpload)
		conf.Config.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		viper.Set("account.secure", *patch.IsSecure)
		conf.Config.IsSecure = *patch.IsSecure
	}
	if err := writeCurrentViperConfigAtomic(); err != nil {
		rollback.restore()
		*conf.Config = originalConfig
		return err
	}
	return nil
}

func runtimePatchFromDTO(patch ConfigPatchDTO) runtimeconf.ConfigPatch {
	return runtimeconf.ConfigPatch{
		RootPath:      patch.RootPath,
		Port:          patch.Port,
		MaxLevel:      patch.MaxLevel,
		IsAllowUpload: patch.IsAllowUpload,
		IsSecure:      patch.IsSecure,
	}
}

func snapshotWithPatch(snapshot runtimeconf.ConfigSnapshot, patch ConfigPatchDTO) runtimeconf.ConfigSnapshot {
	if patch.RootPath != nil {
		snapshot.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		snapshot.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		snapshot.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		snapshot.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		snapshot.IsSecure = *patch.IsSecure
	}
	return snapshot
}

func configDTOFromSnapshot(snapshot runtimeconf.ConfigSnapshot) ConfigDTO {
	return ConfigDTO{
		RootPath:      snapshot.RootPath,
		Port:          snapshot.Port,
		MaxLevel:      snapshot.MaxLevel,
		IsAllowUpload: snapshot.IsAllowUpload,
		IsSecure:      snapshot.IsSecure,
	}
}

func addressFromSnapshot(snapshot runtimeconf.ConfigSnapshot, running bool) string {
	if !running {
		return ""
	}
	host := snapshot.IP
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%d", host, snapshot.Port)
}

func hasCredentials(snapshot runtimeconf.ConfigSnapshot) bool {
	return security.CredentialConfig{
		Username:     snapshot.Username,
		PasswordHash: snapshot.PasswordHash,
		HashAlgo:     security.HashAlgoBcrypt,
	}.IsConfigured()
}

func logEntriesFromEvents(events []accesslog.Event) []LogEntryDTO {
	entries := make([]LogEntryDTO, len(events))
	for i, event := range events {
		entries[i] = LogEntryDTO{
			Time:     event.Time,
			ClientIP: event.ClientIP,
			Method:   event.Method,
			Event:    event.Event,
			Path:     event.Path,
			Status:   event.Status,
			Result:   event.Result,
		}
	}
	return entries
}

func timeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func intPtr(value int) *int {
	return &value
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
		path, err = conf.ConfigPath()
		if err != nil {
			return err
		}
	}
	data, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}
	return atomicWriteConfigFile(path, data)
}

func atomicWriteFile(path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return ErrInvalidExportPath
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

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

	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/tcp404/OneTiny/internal/server"
	"github.com/tcp404/OneTiny/internal/state"
)

type Controller struct {
	mu                  sync.Mutex
	cfg                 *state.RuntimeConfig
	manager             *server.ServiceManager
	logger              *accesslog.Logger
	lastErr             string
	portRestartRequired bool
	pendingPort         *int
}

var (
	restartServiceWithSnapshot = func(manager *server.ServiceManager, snapshot state.ConfigSnapshot, commit func() error) error {
		return manager.RestartWithSnapshot(snapshot, commit)
	}
)

func NewController() *Controller {
	process := state.NewProcessState()
	process.IP = "127.0.0.1"
	if process.SessionVal == "" {
		process.SessionVal = "test-session"
	}
	return NewControllerWithState(state.NewRuntimeConfig(state.SnapshotFromConfig(conf.Current(), process)))
}

func NewControllerWithState(cfg *state.RuntimeConfig) *Controller {
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
	state.SetCurrent(nil)
	if c.portRestartRequired || c.pendingPort != nil {
		port := conf.Current().Port
		if err := c.manager.ApplyRuntimeConfig(state.ConfigPatch{Port: &port}); err != nil {
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
		savedSnapshot, err := persistConfigPatch(persistPatch, state.ProcessStateFromSnapshot(active))
		if err != nil {
			c.pendingPort = nil
			c.portRestartRequired = true
			c.lastErr = err.Error()
			return c.statusLocked(), err
		}
		if err := restartServiceWithSnapshot(c.manager, savedSnapshot, nil); err != nil {
			c.resetPendingPortFailureLocked(active.Port)
			c.lastErr = err.Error()
			return c.statusLocked(), err
		}
		c.pendingPort = nil
		c.portRestartRequired = false
		c.lastErr = ""
		return c.statusLocked(), nil
	}

	savedSnapshot, err := persistConfigPatch(persistPatch, state.ProcessStateFromSnapshot(active))
	if err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}

	runtimePatch := runtimePatchFromSnapshot(active, savedSnapshot)
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

	securityPatch := conf.SecurityPatch{
		Username:     &username,
		PasswordHash: &hash,
	}
	if patch.EnableSecure {
		securityPatch.IsSecure = &patch.EnableSecure
	}
	savedConfig, err := conf.SaveSecurityPatch(securityPatch)
	if err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}
	active := c.cfg.Snapshot()
	savedSnapshot := state.SnapshotFromConfig(savedConfig, state.ProcessStateFromSnapshot(active))
	if err := c.manager.ApplyRuntimeConfig(state.ConfigPatch{
		Username:     &savedSnapshot.Username,
		PasswordHash: &savedSnapshot.PasswordHash,
		IsSecure:     &savedSnapshot.IsSecure,
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

func (c *Controller) portTargetLocked(patch ConfigPatchDTO, active state.ConfigSnapshot) (int, bool) {
	if patch.RestartPort && c.pendingPort != nil {
		return *c.pendingPort, true
	}
	if savedPort := conf.Current().Port; patch.RestartPort && c.portRestartRequired && savedPort != active.Port {
		return savedPort, true
	}
	if patch.Port != nil {
		return *patch.Port, true
	}
	return 0, false
}

func (c *Controller) configDTOLocked(snapshot state.ConfigSnapshot) ConfigDTO {
	dto := configDTOFromSnapshot(snapshot)
	if c.pendingPort != nil {
		dto.Port = *c.pendingPort
	}
	return dto
}

func (c *Controller) resetPendingPortFailureLocked(activePort int) {
	if _, err := conf.SavePatch(conf.ConfigPatch{Port: &activePort}); err != nil {
		c.lastErr = err.Error()
	}
	c.pendingPort = nil
	c.portRestartRequired = true
}

func snapshotFromConf() state.ConfigSnapshot {
	return state.SnapshotFromConfig(conf.Current(), state.NewProcessState())
}

func persistConfigPatch(patch ConfigPatchDTO, process state.ProcessState) (state.ConfigSnapshot, error) {
	cfg, err := conf.SavePatch(confPatchFromDTO(patch))
	if err != nil {
		return state.ConfigSnapshot{}, err
	}
	return state.SnapshotFromConfig(cfg, process), nil
}

func confPatchFromDTO(patch ConfigPatchDTO) conf.ConfigPatch {
	return conf.ConfigPatch{
		RootPath:      patch.RootPath,
		Port:          patch.Port,
		MaxLevel:      patch.MaxLevel,
		IsAllowUpload: patch.IsAllowUpload,
		IsSecure:      patch.IsSecure,
	}
}

func runtimePatchFromSnapshot(old, next state.ConfigSnapshot) state.ConfigPatch {
	patch := state.ConfigPatch{}
	if old.RootPath != next.RootPath {
		patch.RootPath = &next.RootPath
	}
	if old.Port != next.Port {
		patch.Port = &next.Port
	}
	if old.MaxLevel != next.MaxLevel {
		patch.MaxLevel = &next.MaxLevel
	}
	if old.IsAllowUpload != next.IsAllowUpload {
		patch.IsAllowUpload = &next.IsAllowUpload
	}
	if old.IsSecure != next.IsSecure {
		patch.IsSecure = &next.IsSecure
	}
	return patch
}

func snapshotWithPatch(snapshot state.ConfigSnapshot, patch ConfigPatchDTO) state.ConfigSnapshot {
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

func configDTOFromSnapshot(snapshot state.ConfigSnapshot) ConfigDTO {
	return ConfigDTO{
		RootPath:      snapshot.RootPath,
		Port:          snapshot.Port,
		MaxLevel:      snapshot.MaxLevel,
		IsAllowUpload: snapshot.IsAllowUpload,
		IsSecure:      snapshot.IsSecure,
	}
}

func addressFromSnapshot(snapshot state.ConfigSnapshot, running bool) string {
	if !running {
		return ""
	}
	host := snapshot.IP
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%d", host, snapshot.Port)
}

func hasCredentials(snapshot state.ConfigSnapshot) bool {
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

# Wails Desktop Control Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first OneTiny desktop GUI control panel with Wails v3, tray behavior, hot runtime config updates, credential setup, and access log browsing/export.

**Architecture:** Keep the CLI/server behavior stable by adding a pure-Go `internal/control` facade first, then bind that facade to Wails. The desktop frontend talks only to Wails-bound service methods; the service delegates lifecycle, config, credentials, logs, and dialogs to backend components.

**Tech Stack:** Go 1.26, Gin, Viper, bcrypt, JSON Lines access logs, Wails v3, vanilla TypeScript/CSS frontend.

---

## External API Notes Checked

- Wails v3 uses `github.com/wailsapp/wails/v3/pkg/application`, `application.New`, service binding through `application.NewService`, and `app.Run()`.
- Wails v3 frontend Go calls are asynchronous promises.
- Wails v3 supports system tray menus, left/right click handlers, `AttachWindow`, and native dialogs through `app.Dialog`.
- Wails v3 window close behavior uses `RegisterHook(events.Common.WindowClosing, ...)`, `event.Cancel()`, and `window.Hide()` for hide-to-tray behavior.

## File Structure

- `internal/runtimeconf/config.go`: extend runtime snapshots with login credential fields so GUI credential changes affect new login attempts without restarting the server.
- `internal/server/server.go`: route CLI startup through `ServiceManager` helper path so CLI and GUI share the same lifecycle code.
- `internal/server/manager.go`: add `ApplyRuntimeConfig`, `Config`, and address helpers needed by GUI status.
- `internal/handle/secure/login.go`: read credentials from request runtime snapshot instead of only `conf.Config`.
- `internal/control/types.go`: DTOs returned to Wails frontend.
- `internal/control/controller.go`: pure-Go control panel backend with start/stop/update/credentials/logs/export.
- `internal/control/controller_test.go`: tests for hot update, restart-required port changes, credential setup, logs, and lifecycle.
- `internal/gui/service.go`: Wails-facing service methods that wrap `internal/control.Controller`.
- `internal/gui/dialogs.go`: Wails dialog adapter for choose directory, export target, confirmation, and opening config dir.
- `internal/gui/app.go`: Wails app/window/tray construction.
- `cmd/onetiny-gui/main.go`: desktop app entrypoint.
- `frontend/assets.go`: Go embed package for `frontend/dist` so the GUI entrypoint can import assets without illegal `..` embed paths.
- `frontend/index.html`: desktop app shell.
- `frontend/package.json`: Vite/TypeScript scripts.
- `frontend/tsconfig.json`: TypeScript config.
- `frontend/src/main.ts`: frontend state, binding calls, and event handlers.
- `frontend/src/styles.css`: production UI styling.
- `frontend/src/types.ts`: frontend DTO types mirrored from Go JSON.
- `README.md`: GUI usage note.

## Shared DTOs

Use these names consistently across backend and frontend:

```go
package control

import "time"

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
```

## Task 1: Runtime Credentials And Unified Lifecycle

**Files:**
- Modify: `internal/runtimeconf/config.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/manager.go`
- Modify: `internal/handle/secure/login.go`
- Modify: `internal/handle/secure/login_test.go`
- Modify: `internal/server/manager_test.go`

- [ ] **Step 1: Write failing login runtime credential test**

Add this test to `internal/handle/secure/login_test.go`:

```go
func TestLoginPostUsesRuntimeCredentials(t *testing.T) {
	logger := resetLoginTestConfig(t)
	hash, err := security.HashPassword("runtime-pass")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.Config.Username = "global-user"
	conf.Config.Password = "$2a$10$YJCMw3VjB9FlGm8zJbv8we8z0N1P6l4L7jXWaCOc3SNH0WcEjPzNe"
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		Username:     "runtime-user",
		PasswordHash: hash,
		SessionVal:   "runtime-session",
	}))

	result, rec := postLogin(t, "runtime-user", "runtime-pass")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if result["code"] != float64(1) {
		t.Fatalf("login result = %+v, want success", result)
	}
	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventLogin})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 || events[0].Result != accesslog.ResultSuccess {
		t.Fatalf("login events = %+v, want one success", events)
	}
}
```

- [ ] **Step 2: Run failing login test**

Run:

```bash
go test -count=1 ./internal/handle/secure -run TestLoginPostUsesRuntimeCredentials
```

Expected: FAIL because `ConfigSnapshot` has no credential fields or login ignores them.

- [ ] **Step 3: Extend runtime config snapshot**

In `internal/runtimeconf/config.go`, update the structs:

```go
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
```

Add the matching assignments in `RuntimeConfig.Update`.

- [ ] **Step 4: Populate runtime credentials from global config**

In `internal/server/server.go`, update `snapshotFromGlobalConfig`:

```go
func snapshotFromGlobalConfig() runtimeconf.ConfigSnapshot {
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
```

- [ ] **Step 5: Read login credentials from runtime snapshot**

In `internal/handle/secure/login.go`, replace the credential read in `LoginPost` with:

```go
func LoginPost(c *gin.Context) {
	cfg := loginSnapshot()
	if c.PostForm("username") == cfg.Username &&
		security.VerifyPassword(cfg.PasswordHash, c.PostForm("password")) == nil {
		session := sessions.Default(c)
		session.Set("login", cfg.SessionVal)
		session.Save()
		logLoginEvent(c, accesslog.ResultSuccess)
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "登录成功"})
		return
	}
	logLoginEvent(c, accesslog.ResultFailure)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "登录失败"})
}

type loginCredentialSnapshot struct {
	Username     string
	PasswordHash string
	SessionVal   string
}

func loginSnapshot() loginCredentialSnapshot {
	cfg := runtimeconf.Current()
	if cfg == nil {
		return loginCredentialSnapshot{
			Username:     conf.Config.Username,
			PasswordHash: conf.Config.Password,
			SessionVal:   conf.Config.SessionVal,
		}
	}
	snapshot := cfg.Snapshot()
	if snapshot.SessionVal == "" {
		snapshot.SessionVal = conf.Config.SessionVal
	}
	return loginCredentialSnapshot{
		Username:     snapshot.Username,
		PasswordHash: snapshot.PasswordHash,
		SessionVal:   snapshot.SessionVal,
	}
}
```

Add the missing import:

```go
"github.com/tcp404/OneTiny-cli/internal/runtimeconf"
```

- [ ] **Step 6: Add manager lifecycle helper tests**

In `internal/server/manager_test.go`, add:

```go
func TestServiceManagerApplyRuntimeConfig(t *testing.T) {
	cfg := runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: t.TempDir(),
		Port:     freeTestPort(t),
		MaxLevel: 0,
	})
	manager := NewServiceManager(cfg)

	nextRoot := t.TempDir()
	upload := true
	level := uint8(2)
	manager.ApplyRuntimeConfig(runtimeconf.ConfigPatch{
		RootPath:      &nextRoot,
		IsAllowUpload: &upload,
		MaxLevel:      &level,
	})

	got := manager.Status()
	if got.RootPath != nextRoot || !got.IsAllowUpload || got.MaxLevel != 2 {
		t.Fatalf("status = %+v, want updated runtime config", got)
	}
}
```

- [ ] **Step 7: Implement manager helpers**

In `internal/server/manager.go`, add:

```go
func (m *ServiceManager) ApplyRuntimeConfig(patch runtimeconf.ConfigPatch) error {
	if m.cfg == nil {
		return ErrRuntimeConfigRequired
	}
	m.cfg.Update(patch)
	return nil
}

func (m *ServiceManager) Config() *runtimeconf.RuntimeConfig {
	return m.cfg
}
```

- [ ] **Step 8: Route CLI startup through ServiceManager**

In `internal/server/server.go`, rewrite `RunCore` to use `ServiceManager` while preserving signal behavior:

```go
func RunCore() {
	cfg := runtimeconf.NewRuntimeConfig(snapshotFromGlobalConfig())
	manager := NewServiceManager(cfg)
	if err := manager.Start(); err != nil {
		log.Println(color.RedString(err.Error()))
		return
	}
	printInfo()

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)
	<-q
	if err := manager.Stop(); err != nil {
		log.Println(color.RedString(err.Error()))
	}
	fmt.Println(color.GreenString("\nbye~"))
	os.Exit(0)
}
```

Remove `initServer`, `run`, and `exit` only after confirming no code references them. Keep `setupEngine`, `snapshotFromGlobalConfig`, and `printInfo`.

- [ ] **Step 9: Verify task 1**

Run:

```bash
go test -count=1 ./internal/runtimeconf ./internal/server ./internal/handle/secure
go test -count=1 ./...
git diff --check
```

Expected: PASS and no diff check output.

- [ ] **Step 10: Commit task 1**

```bash
git add internal/runtimeconf internal/server internal/handle/secure
git commit -m "feat: share runtime credentials and lifecycle"
```

## Task 2: Pure Go Control Controller

**Files:**
- Create: `internal/control/types.go`
- Create: `internal/control/controller.go`
- Create: `internal/control/controller_test.go`
- Modify: `internal/conf/conf.go`

- [ ] **Step 1: Add config path helpers**

In `internal/conf/conf.go`, add:

```go
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
```

Use these helpers inside `LoadConfig` so CLI and GUI agree on paths:

```go
cfgDir, err := ConfigDir()
if err != nil {
	return err
}
cfgFile, err := ConfigPath()
if err != nil {
	return err
}
```

- [ ] **Step 2: Write failing controller lifecycle test**

Create `internal/control/controller_test.go`:

```go
package control

import (
	"testing"

	"github.com/tcp404/OneTiny-cli/internal/conf"
	"github.com/tcp404/OneTiny-cli/internal/runtimeconf"
)

func TestControllerStartStopAndStatus(t *testing.T) {
	resetControllerTest(t)
	port := freeControlTestPort(t)
	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.Port = port
	conf.Config.MaxLevel = 1
	conf.Config.IsAllowUpload = false
	conf.Config.IsSecure = false

	controller := NewController()
	status, err := controller.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if status.Running {
		t.Fatalf("initial status running = true, want false")
	}

	status, err = controller.StartSharing()
	if err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	if !status.Running || status.Address == "" {
		t.Fatalf("started status = %+v, want running with address", status)
	}

	status, err = controller.StopSharing()
	if err != nil {
		t.Fatalf("StopSharing returned error: %v", err)
	}
	if status.Running {
		t.Fatalf("stopped status running = true, want false")
	}
	runtimeconf.SetCurrent(nil)
}
```

- [ ] **Step 3: Write failing config update and credential tests**

Append:

```go
func TestControllerUpdateConfigHotAndPortRestart(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()
	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.Port = freeControlTestPort(t)

	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer controller.StopSharing()

	nextRoot := t.TempDir()
	allowUpload := true
	status, err := controller.UpdateConfig(ConfigPatchDTO{
		RootPath:      &nextRoot,
		IsAllowUpload: &allowUpload,
	})
	if err != nil {
		t.Fatalf("UpdateConfig returned error: %v", err)
	}
	if status.Config.RootPath != nextRoot || !status.Config.IsAllowUpload || status.PortRestartRequired {
		t.Fatalf("hot update status = %+v, want hot root/upload update", status)
	}

	nextPort := freeControlTestPort(t)
	status, err = controller.UpdateConfig(ConfigPatchDTO{Port: &nextPort})
	if err != nil {
		t.Fatalf("UpdateConfig port returned error: %v", err)
	}
	if !status.PortRestartRequired || status.Config.Port != nextPort {
		t.Fatalf("port update status = %+v, want restart required", status)
	}
}

func TestControllerSetCredentialsEnablesSecure(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()

	status, err := controller.SetCredentials(CredentialPatchDTO{
		Username:        "admin",
		Password:        "strong-password",
		ConfirmPassword: "strong-password",
		EnableSecure:    true,
	})
	if err != nil {
		t.Fatalf("SetCredentials returned error: %v", err)
	}
	if !status.HasCredentials || !status.Config.IsSecure {
		t.Fatalf("credential status = %+v, want configured and secure enabled", status)
	}
	if conf.Config.Username != "admin" {
		t.Fatalf("conf username = %q, want admin", conf.Config.Username)
	}
}
```

- [ ] **Step 4: Add controller DTOs**

Create `internal/control/types.go` using the shared DTOs from this plan. Add:

```go
var (
	ErrPasswordConfirmationMismatch = errors.New("两次输入的密码不一致")
	ErrPortRestartRequiresConfirm   = errors.New("修改端口需要确认并重启服务")
)
```

- [ ] **Step 5: Implement controller**

Create `internal/control/controller.go`:

```go
package control

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/tcp404/OneTiny-cli/internal/accesslog"
	"github.com/tcp404/OneTiny-cli/internal/conf"
	"github.com/tcp404/OneTiny-cli/internal/runtimeconf"
	"github.com/tcp404/OneTiny-cli/internal/security"
	"github.com/tcp404/OneTiny-cli/internal/server"
	"github.com/spf13/viper"
)

type Controller struct {
	mu      sync.Mutex
	cfg     *runtimeconf.RuntimeConfig
	manager *server.ServiceManager
	logger  *accesslog.Logger
	lastErr string
}

func NewController() *Controller {
	snapshot := snapshotFromConf()
	cfg := runtimeconf.NewRuntimeConfig(snapshot)
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
	c.lastErr = ""
	return c.statusLocked(), nil
}

func (c *Controller) UpdateConfig(patch ConfigPatchDTO) (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if patch.Port != nil && *patch.Port != c.cfg.Snapshot().Port && !patch.RestartPort {
		viper.Set("server.port", *patch.Port)
		conf.Config.Port = *patch.Port
		c.lastErr = ErrPortRestartRequiresConfirm.Error()
		status := c.statusLocked()
		status.PortRestartRequired = true
		return status, nil
	}
	if err := persistConfigPatch(patch); err != nil {
		c.lastErr = err.Error()
		return c.statusLocked(), err
	}
	c.cfg.Update(runtimePatchFromDTO(patch))
	if patch.Port != nil && patch.RestartPort {
		if err := c.manager.Restart(); err != nil {
			c.lastErr = err.Error()
			return c.statusLocked(), err
		}
	}
	c.lastErr = ""
	return c.statusLocked(), nil
}

func (c *Controller) SetCredentials(patch CredentialPatchDTO) (StatusDTO, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if patch.Password != patch.ConfirmPassword {
		return c.statusLocked(), ErrPasswordConfirmationMismatch
	}
	hash, err := security.HashPassword(patch.Password)
	if err != nil {
		return c.statusLocked(), err
	}
	conf.SetCredentialConfig(patch.Username, hash)
	if patch.EnableSecure {
		viper.Set("account.secure", true)
	}
	if err := viper.WriteConfig(); err != nil {
		return c.statusLocked(), err
	}
	conf.Config.Username = patch.Username
	conf.Config.Password = hash
	if patch.EnableSecure {
		conf.Config.IsSecure = true
	}
	c.cfg.Update(runtimeconf.ConfigPatch{
		Username:     &patch.Username,
		PasswordHash: &hash,
		IsSecure:     boolPtr(conf.Config.IsSecure),
	})
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
	defer writer.Flush()
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
	return writer.Error()
}

func (c *Controller) statusLocked() StatusDTO {
	snapshot := c.cfg.Snapshot()
	configPath, _ := conf.ConfigPath()
	state := "未运行"
	if c.manager.Running() {
		state = "运行中"
	}
	return StatusDTO{
		Running:       c.manager.Running(),
		StateLabel:    state,
		Address:       addressFromSnapshot(snapshot, c.manager.Running()),
		Config:        configDTOFromSnapshot(snapshot),
		HasCredentials: security.CredentialConfig{
			Username:     snapshot.Username,
			PasswordHash: snapshot.PasswordHash,
			HashAlgo:     security.HashAlgoBcrypt,
		}.IsConfigured(),
		ConfigPath:    configPath,
		AccessLogPath: accesslog.DefaultPath(),
		LastError:     c.lastErr,
	}
}
```

Add these helper functions in the same file:

```go
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
		if *patch.IsSecure {
			if err := conf.ValidateSecureConfigFor(true); err != nil {
				return err
			}
		}
		viper.Set("account.secure", *patch.IsSecure)
		conf.Config.IsSecure = *patch.IsSecure
	}
	return viper.WriteConfig()
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

func boolPtr(value bool) *bool {
	return &value
}
```

- [ ] **Step 6: Add test helpers**

In `internal/control/controller_test.go`, add:

```go
func resetControllerTest(t *testing.T) {
	t.Helper()
	originalConfig := *conf.Config
	originalViper := viper.GetViper()
	v := viper.New()
	v.SetConfigType("yml")
	v.SetConfigFile(filepath.Join(t.TempDir(), "config.yml"))
	viper.Reset()
	for _, key := range v.AllKeys() {
		viper.Set(key, v.Get(key))
	}
	t.Cleanup(func() {
		*conf.Config = originalConfig
		runtimeconf.SetCurrent(nil)
		viper.Reset()
		for _, key := range originalViper.AllKeys() {
			viper.Set(key, originalViper.Get(key))
		}
	})
}

func freeControlTestPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen test port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
```

If the viper restore approach conflicts with existing tests, replace it with the same isolation helper already used by `internal/conf` tests.

- [ ] **Step 7: Verify task 2**

Run:

```bash
go test -count=1 ./internal/control ./internal/conf ./internal/server
go test -count=1 ./...
git diff --check
```

Expected: PASS and no diff check output.

- [ ] **Step 8: Commit task 2**

```bash
git add internal/control internal/conf
git commit -m "feat: add GUI control controller"
```

## Task 3: Wails Shell And Service Binding

**Files:**
- Create: `cmd/onetiny-gui/main.go`
- Create: `internal/gui/app.go`
- Create: `internal/gui/service.go`
- Create: `internal/gui/dialogs.go`
- Create: `frontend/index.html`
- Create: `frontend/assets.go`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/types.ts`
- Create: `frontend/src/styles.css`
- Create: `frontend/package.json`
- Create: `frontend/tsconfig.json`
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add Wails dependency and verify CLI availability**

Run:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
wails3 doctor
go get github.com/wailsapp/wails/v3@latest
go mod tidy
```

Expected: `wails3 doctor` reports usable local prerequisites. If Node is missing, install Node before continuing; do not vendor generated dependency folders.

- [ ] **Step 2: Add minimal frontend scaffold**

Create `frontend/package.json`:

```json
{
  "name": "onetiny-control-panel",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite --host 127.0.0.1",
    "build": "tsc --noEmit && vite build",
    "preview": "vite preview --host 127.0.0.1"
  },
  "devDependencies": {
    "typescript": "^5.5.4",
    "vite": "^5.4.0"
  }
}
```

Create `frontend/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "Bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "strict": true
  },
  "include": ["src"]
}
```

Create `frontend/index.html`:

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>OneTiny 控制面板</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

- [ ] **Step 3: Add Wails service wrapper**

Create `internal/gui/service.go`:

```go
package gui

import "github.com/tcp404/OneTiny-cli/internal/control"

type Service struct {
	controller *control.Controller
	dialogs    DialogAdapter
}

func NewService(controller *control.Controller, dialogs DialogAdapter) *Service {
	return &Service{controller: controller, dialogs: dialogs}
}

func (s *Service) GetStatus() (control.StatusDTO, error) {
	return s.controller.GetStatus()
}

func (s *Service) StartSharing() (control.StatusDTO, error) {
	return s.controller.StartSharing()
}

func (s *Service) StopSharing() (control.StatusDTO, error) {
	return s.controller.StopSharing()
}

func (s *Service) UpdateConfig(patch control.ConfigPatchDTO) (control.StatusDTO, error) {
	return s.controller.UpdateConfig(patch)
}

func (s *Service) SetCredentials(patch control.CredentialPatchDTO) (control.StatusDTO, error) {
	return s.controller.SetCredentials(patch)
}

func (s *Service) GetLogs(filter control.LogFilterDTO) ([]control.LogEntryDTO, error) {
	return s.controller.GetLogs(filter)
}

func (s *Service) ClearLogs() error {
	return s.controller.ClearLogs()
}

func (s *Service) ChooseDirectory(current string) (string, error) {
	return s.dialogs.ChooseDirectory(current)
}

func (s *Service) ExportLogs(filter control.LogFilterDTO) (string, error) {
	path, err := s.dialogs.ChooseExportPath("onetiny-access-log.csv")
	if err != nil || path == "" {
		return "", err
	}
	if err := s.controller.ExportLogs(path, filter); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) OpenConfigDir() error {
	return s.dialogs.OpenConfigDir()
}
```

- [ ] **Step 4: Add dialog adapter**

Create `internal/gui/dialogs.go`:

```go
package gui

import (
	"os/exec"
	"runtime"

	"github.com/tcp404/OneTiny-cli/internal/conf"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DialogAdapter interface {
	ChooseDirectory(current string) (string, error)
	ChooseExportPath(defaultFilename string) (string, error)
	OpenConfigDir() error
}

type WailsDialogs struct {
	app *application.App
}

func NewWailsDialogs(app *application.App) *WailsDialogs {
	return &WailsDialogs{app: app}
}

func (d *WailsDialogs) ChooseDirectory(current string) (string, error) {
	return d.app.Dialog.OpenFile().
		SetTitle("选择共享目录").
		SetDefaultDirectory(current).
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
}

func (d *WailsDialogs) ChooseExportPath(defaultFilename string) (string, error) {
	return d.app.Dialog.SaveFile().
		SetTitle("导出访问日志").
		SetDefaultFilename(defaultFilename).
		AddFilter("CSV", "*.csv").
		PromptForSingleSelection()
}

func (d *WailsDialogs) OpenConfigDir() error {
	dir, err := conf.ConfigDir()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", dir).Start()
	case "windows":
		return exec.Command("explorer", dir).Start()
	default:
		return exec.Command("xdg-open", dir).Start()
	}
}
```

- [ ] **Step 5: Add Wails app construction**

Create `internal/gui/app.go`:

```go
package gui

import (
	"embed"

	"github.com/tcp404/OneTiny-cli/internal/control"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
)

func Run(assets embed.FS) error {
	app := application.New(application.Options{
		Name:        "OneTiny",
		Description: "局域网文件共享控制面板",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	controller := control.NewController()
	service := NewService(controller, NewWailsDialogs(app))
	app.RegisterService(application.NewService(service))

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "OneTiny",
		Width:  960,
		Height: 680,
		URL:    "/",
	})
	window.SetMinSize(820, 560)

	tray := app.SystemTray.New()
	tray.SetLabel("OneTiny")
	tray.SetIcon(icons.SystrayLight)
	tray.SetDarkModeIcon(icons.SystrayDark)
	tray.AttachWindow(window)

	menu := app.NewMenu()
	menu.Add("打开面板").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Restore()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(ctx *application.Context) {
		if status, _ := controller.GetStatus(); status.Running {
			_, _ = controller.StopSharing()
		}
		app.Quit()
	})
	tray.SetMenu(menu)

	registerHideOnClose(window)
	return app.Run()
}

func registerHideOnClose(window *application.WebviewWindow) {
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		window.Hide()
	})
}
```

The behavior must be: clicking the window close button cancels close and hides the window.

- [ ] **Step 6: Add GUI entrypoint**

Create `cmd/onetiny-gui/main.go`:

```go
package main

import (
	"log"

	"github.com/tcp404/OneTiny-cli/frontend"
	"github.com/tcp404/OneTiny-cli/internal/conf"
	"github.com/tcp404/OneTiny-cli/internal/gui"
)

func main() {
	if err := conf.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	if err := gui.Run(frontend.Assets); err != nil {
		log.Fatal(err)
	}
}
```

Create `frontend/assets.go`:

```go
package frontend

import "embed"

//go:embed all:dist
var Assets embed.FS
```

Build `frontend/dist` before running `go build ./cmd/onetiny-gui`.

- [ ] **Step 7: Add minimal frontend TypeScript**

Create `frontend/src/types.ts` matching the shared DTO JSON names:

```ts
export interface ConfigDTO {
  rootPath: string;
  port: number;
  maxLevel: number;
  isAllowUpload: boolean;
  isSecure: boolean;
}

export interface StatusDTO {
  running: boolean;
  stateLabel: string;
  address: string;
  config: ConfigDTO;
  hasCredentials: boolean;
  configPath: string;
  accessLogPath: string;
  portRestartRequired: boolean;
  lastError: string;
}

export interface ConfigPatchDTO {
  rootPath?: string;
  port?: number;
  maxLevel?: number;
  isAllowUpload?: boolean;
  isSecure?: boolean;
  restartPort?: boolean;
}

export interface CredentialPatchDTO {
  username: string;
  password: string;
  confirmPassword: string;
  enableSecure: boolean;
}

export interface LogFilterDTO {
  event?: string;
  since?: string;
  until?: string;
}

export interface LogEntryDTO {
  time: string;
  clientIP: string;
  method: string;
  event: string;
  path: string;
  status: number;
  result: string;
}
```

Create `frontend/src/main.ts` with a local browser preview service shim and real Wails bindings for desktop:

```ts
import "./styles.css";
import type { ConfigPatchDTO, CredentialPatchDTO, LogEntryDTO, LogFilterDTO, StatusDTO } from "./types";

type GuiService = {
  GetStatus(): Promise<StatusDTO>;
  StartSharing(): Promise<StatusDTO>;
  StopSharing(): Promise<StatusDTO>;
  UpdateConfig(patch: ConfigPatchDTO): Promise<StatusDTO>;
  SetCredentials(patch: CredentialPatchDTO): Promise<StatusDTO>;
  ChooseDirectory(current: string): Promise<string>;
  GetLogs(filter: LogFilterDTO): Promise<LogEntryDTO[]>;
  ClearLogs(): Promise<void>;
  ExportLogs(filter: LogFilterDTO): Promise<string>;
  OpenConfigDir(): Promise<void>;
};

declare global {
  interface Window {
    go?: { gui?: { Service?: GuiService } };
  }
}

const service = window.go?.gui?.Service ?? mockService();
let status: StatusDTO | null = null;
let activeTab: "control" | "security" | "logs" | "about" = "control";
let logFilter: LogFilterDTO = {};
let logs: LogEntryDTO[] = [];

document.querySelector<HTMLDivElement>("#app")!.innerHTML = `
  <main class="shell">
    <section class="topbar">
      <div>
        <span class="label">访问地址</span>
        <span id="state" class="state">未运行</span>
        <div id="address" class="address">-</div>
      </div>
      <div class="actions">
        <button id="copyAddress" type="button">复制地址</button>
        <button id="toggleShare" type="button">启动共享</button>
      </div>
    </section>
    <nav class="tabs">
      <button data-tab="control" class="active">控制面板</button>
      <button data-tab="security">安全设置</button>
      <button data-tab="logs">访问日志</button>
      <button data-tab="about">关于</button>
    </nav>
    <section id="content" class="content"></section>
  </main>
`;

async function refresh(): Promise<void> {
  status = await service.GetStatus();
  if (activeTab === "logs") {
    logs = await service.GetLogs(logFilter);
  }
  render();
}

function render(): void {
  if (!status) return;
  document.querySelector("#state")!.textContent = status.stateLabel;
  document.querySelector("#state")!.className = `state ${status.running ? "running" : ""}`;
  document.querySelector("#address")!.textContent = status.running && status.address ? status.address : "-";
  document.querySelector("#toggleShare")!.textContent = status.running ? "停止共享" : "启动共享";

  const content = document.querySelector("#content")!;
  if (activeTab === "control") content.innerHTML = controlView(status);
  if (activeTab === "security") content.innerHTML = securityView(status);
  if (activeTab === "logs") content.innerHTML = logsView(logs);
  if (activeTab === "about") content.innerHTML = aboutView(status);
  bindContentEvents();
}

function controlView(current: StatusDTO): string {
  return `<div class="preview-panel">共享目录：${escapeHtml(current.config.rootPath)}</div>`;
}

function securityView(current: StatusDTO): string {
  return `<div class="preview-panel">登录保护：${current.config.isSecure ? "已开启" : "未开启"}</div>`;
}

function logsView(entries: LogEntryDTO[]): string {
  return `<div class="preview-panel">访问日志 ${entries.length} 条</div>`;
}

function aboutView(current: StatusDTO): string {
  return `<div class="preview-panel">配置目录：${escapeHtml(current.configPath)}</div>`;
}

function bindContentEvents(): void {
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "\"": "&quot;",
    "'": "&#39;",
  })[char] ?? char);
}

function mockService(): GuiService {
  let current: StatusDTO = {
    running: false,
    stateLabel: "未运行",
    address: "",
    config: {
      rootPath: "/Users/demo/Downloads",
      port: 8192,
      maxLevel: 0,
      isAllowUpload: false,
      isSecure: false,
    },
    hasCredentials: false,
    configPath: "/Users/demo/Library/Application Support/tiny/config.yml",
    accessLogPath: "/Users/demo/Library/Application Support/tiny/access.log",
    portRestartRequired: false,
    lastError: "",
  };
  return {
    async GetStatus() { return current; },
    async StartSharing() {
      current = { ...current, running: true, stateLabel: "运行中", address: `http://127.0.0.1:${current.config.port}` };
      return current;
    },
    async StopSharing() {
      current = { ...current, running: false, stateLabel: "未运行", address: "" };
      return current;
    },
    async UpdateConfig(patch) {
      current = { ...current, config: { ...current.config, ...patch }, portRestartRequired: Boolean(patch.port && !patch.restartPort) };
      return current;
    },
    async SetCredentials(patch) {
      current = { ...current, hasCredentials: true, config: { ...current.config, isSecure: patch.enableSecure } };
      return current;
    },
    async ChooseDirectory() { return "/Users/demo/Public"; },
    async GetLogs() { return []; },
    async ClearLogs() {},
    async ExportLogs() { return "/tmp/onetiny-access-log.csv"; },
    async OpenConfigDir() {},
  };
}

void refresh();
```

Task 4 replaces the preview-only views with the full control panel behavior.

- [ ] **Step 8: Add frontend CSS**

Create `frontend/src/styles.css` with stable, non-card-nested desktop UI styling:

```css
:root {
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  color: #18212f;
  background: #f5f7f8;
}

body {
  margin: 0;
  min-width: 820px;
  min-height: 560px;
}

button,
input,
select {
  font: inherit;
}

.shell {
  min-height: 100vh;
  display: grid;
  grid-template-rows: auto auto 1fr;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 18px 24px;
  background: #ffffff;
  border-bottom: 1px solid #d9e0e4;
}

.label {
  font-size: 14px;
  font-weight: 700;
}

.state {
  margin-left: 8px;
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  background: #e6eaed;
  color: #52616d;
  font-size: 12px;
}

.state.running {
  background: #dff4ea;
  color: #0f6b43;
}

.address {
  margin-top: 6px;
  font-size: 20px;
  font-weight: 700;
  min-height: 28px;
}

.actions,
.form-row,
.log-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

button {
  min-height: 34px;
  border: 1px solid #b9c4ca;
  background: #ffffff;
  color: #18212f;
  border-radius: 6px;
  padding: 0 12px;
  cursor: pointer;
}

button.primary,
#toggleShare {
  background: #0f6b43;
  border-color: #0f6b43;
  color: #ffffff;
}

.tabs {
  display: flex;
  gap: 2px;
  padding: 0 24px;
  background: #eef2f4;
  border-bottom: 1px solid #d9e0e4;
}

.tabs button {
  border: 0;
  border-radius: 0;
  background: transparent;
  min-height: 42px;
}

.tabs button.active {
  box-shadow: inset 0 -3px #0f6b43;
  font-weight: 700;
}

.content {
  padding: 22px 24px;
  overflow: auto;
}

.preview-panel,
.panel-form {
  max-width: 760px;
}

.panel-form {
  display: grid;
  gap: 16px;
}

.panel-form label {
  display: grid;
  gap: 6px;
  font-size: 13px;
  font-weight: 700;
}

input,
select {
  min-height: 34px;
  border: 1px solid #b9c4ca;
  border-radius: 6px;
  padding: 0 10px;
  background: #ffffff;
}

.switch-row {
  grid-template-columns: auto 1fr;
  align-items: center;
}

.error {
  color: #b3261e;
  min-height: 20px;
}

dialog {
  border: 1px solid #c8d0d5;
  border-radius: 8px;
  padding: 18px;
}

.logs {
  width: 100%;
  border-collapse: collapse;
  margin-top: 14px;
  background: #ffffff;
}

.logs th,
.logs td {
  border-bottom: 1px solid #d9e0e4;
  padding: 9px 10px;
  text-align: left;
  font-size: 13px;
}

.empty {
  text-align: center;
  color: #6b7780;
}

@media (max-width: 720px) {
  body {
    min-width: 0;
  }
  .topbar {
    align-items: stretch;
    flex-direction: column;
  }
  .actions,
  .form-row,
  .log-toolbar {
    flex-wrap: wrap;
  }
}
```

- [ ] **Step 9: Build frontend and generate Wails bindings**

Run:

```bash
cd frontend
npm install
npm run build
cd ..
wails3 generate bindings
go test -count=1 ./internal/gui ./internal/control
go build ./cmd/onetiny-gui
git diff --check
```

Expected: frontend builds, bindings generated, GUI command builds. Commit generated bindings if Wails generates source files inside `frontend/bindings`.

- [ ] **Step 10: Commit task 3**

```bash
git add go.mod go.sum cmd/onetiny-gui internal/gui frontend
git commit -m "feat: scaffold Wails control panel"
```

## Task 4: Frontend Control Panel Interactions

**Files:**
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/styles.css`
- Modify: `frontend/src/types.ts`

- [ ] **Step 1: Implement top control behavior**

In `frontend/src/main.ts`, implement:

```ts
document.querySelector("#toggleShare")!.addEventListener("click", async () => {
  if (!status) return;
  status = status.running ? await service.StopSharing() : await service.StartSharing();
  render();
});

document.querySelector("#copyAddress")!.addEventListener("click", async () => {
  if (!status?.address) return;
  await navigator.clipboard.writeText(status.address);
});

document.querySelectorAll<HTMLButtonElement>(".tabs button").forEach((button) => {
  button.addEventListener("click", async () => {
    activeTab = button.dataset.tab as typeof activeTab;
    document.querySelectorAll(".tabs button").forEach((item) => item.classList.remove("active"));
    button.classList.add("active");
    await refresh();
  });
});
```

- [ ] **Step 2: Implement control view**

Add:

```ts
function controlView(current: StatusDTO): string {
  const cfg = current.config;
  return `
    <form id="controlForm" class="panel-form">
      <label>共享目录
        <div class="form-row">
          <input id="rootPath" value="${escapeAttr(cfg.rootPath)}" readonly />
          <button id="chooseDir" type="button">选择</button>
        </div>
      </label>
      <label>端口
        <input id="port" type="number" min="1" max="65535" value="${cfg.port}" />
      </label>
      <label>最大访问层级
        <input id="maxLevel" type="number" min="0" max="255" value="${cfg.maxLevel}" />
      </label>
      <label class="switch-row">
        <input id="allowUpload" type="checkbox" ${cfg.isAllowUpload ? "checked" : ""} />
        <span>允许上传</span>
      </label>
      <label class="switch-row">
        <input id="secure" type="checkbox" ${cfg.isSecure ? "checked" : ""} />
        <span>登录保护</span>
      </label>
      <button id="openCredentials" type="button">账号设置</button>
      <p id="formError" class="error"></p>
    </form>
    <dialog id="credentialDialog"></dialog>
  `;
}
```

In `bindContentEvents`, wire:

```ts
document.querySelector("#chooseDir")?.addEventListener("click", async () => {
  if (!status) return;
  const selected = await service.ChooseDirectory(status.config.rootPath);
  if (selected) {
    status = await service.UpdateConfig({ rootPath: selected });
    render();
  }
});

document.querySelector("#allowUpload")?.addEventListener("change", async (event) => {
  status = await service.UpdateConfig({ isAllowUpload: (event.target as HTMLInputElement).checked });
  render();
});

document.querySelector("#maxLevel")?.addEventListener("change", async (event) => {
  status = await service.UpdateConfig({ maxLevel: Number((event.target as HTMLInputElement).value) });
  render();
});

document.querySelector("#port")?.addEventListener("change", async (event) => {
  const port = Number((event.target as HTMLInputElement).value);
  if (!status) return;
  if (status.running && !confirm("修改端口需要重启共享服务，是否继续？")) {
    render();
    return;
  }
  status = await service.UpdateConfig({ port, restartPort: status.running });
  render();
});
```

- [ ] **Step 3: Implement login switch credential prompt**

Add:

```ts
document.querySelector("#secure")?.addEventListener("change", async (event) => {
  const enabled = (event.target as HTMLInputElement).checked;
  if (enabled && !status?.hasCredentials) {
    openCredentialDialog(true);
    render();
    return;
  }
  status = await service.UpdateConfig({ isSecure: enabled });
  render();
});

document.querySelector("#openCredentials")?.addEventListener("click", () => openCredentialDialog(status?.config.isSecure ?? false));
```

Implement `openCredentialDialog(enableSecure: boolean)` as a native `<dialog>` containing username, password, confirm password, Cancel and Save. On save:

```ts
status = await service.SetCredentials({
  username,
  password,
  confirmPassword,
  enableSecure,
});
dialog.close();
render();
```

If password and confirmation differ, show an inline `.error` and do not call backend.

- [ ] **Step 4: Implement security/about tabs**

`securityView` must use the same login switch and credential dialog behavior as control view. `aboutView` must show version text, config path, log path, and an `打开配置目录` button wired to `service.OpenConfigDir()`.

- [ ] **Step 5: Verify with local browser preview**

Run:

```bash
cd frontend
npm run build
npm run preview -- --port 4173
```

Open `http://127.0.0.1:4173` in the in-app browser and verify no layout overlap at 960x680 and 390x844. The local preview uses `mockService`, so service actions should update the mock state.

- [ ] **Step 6: Verify desktop build**

Run:

```bash
wails3 dev
```

Manually verify:
- GUI opens visible.
- Starting sharing changes state and address.
- Closing window hides to tray.
- Tray left click opens the window.
- Tray right click menu shows Open and Quit.
- Non-port config changes apply without restart.
- Port change while running asks for confirmation and restarts.

- [ ] **Step 7: Commit task 4**

```bash
git add frontend
git commit -m "feat: implement control panel interactions"
```

## Task 5: Access Logs Tab

**Files:**
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/styles.css`
- Modify: `internal/control/controller_test.go`

- [ ] **Step 1: Add controller log export test**

In `internal/control/controller_test.go`, add:

```go
func TestControllerLogsFilterClearAndExport(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()
	if err := controller.logger.Write(accesslog.Event{
		Event:    accesslog.EventLogin,
		Path:     "/login",
		Status:   200,
		Result:   accesslog.ResultSuccess,
		ClientIP: "127.0.0.1",
		Method:   "POST",
	}); err != nil {
		t.Fatalf("Write login event: %v", err)
	}
	logs, err := controller.GetLogs(LogFilterDTO{Event: accesslog.EventLogin})
	if err != nil {
		t.Fatalf("GetLogs returned error: %v", err)
	}
	if len(logs) != 1 || logs[0].Event != accesslog.EventLogin {
		t.Fatalf("logs = %+v, want one login event", logs)
	}
	exportPath := filepath.Join(t.TempDir(), "access.csv")
	if err := controller.ExportLogs(exportPath, LogFilterDTO{}); err != nil {
		t.Fatalf("ExportLogs returned error: %v", err)
	}
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(data), "event") || !strings.Contains(string(data), "login") {
		t.Fatalf("csv export = %q, want header and login row", string(data))
	}
	if err := controller.ClearLogs(); err != nil {
		t.Fatalf("ClearLogs returned error: %v", err)
	}
	afterClear, err := controller.GetLogs(LogFilterDTO{})
	if err != nil {
		t.Fatalf("GetLogs after clear returned error: %v", err)
	}
	if len(afterClear) != 0 {
		t.Fatalf("after clear logs = %+v, want empty", afterClear)
	}
}
```

- [ ] **Step 2: Implement logs view**

In `frontend/src/main.ts`, implement:

```ts
function logsView(entries: LogEntryDTO[]): string {
  const rows = entries.map((entry) => `
    <tr>
      <td>${formatTime(entry.time)}</td>
      <td>${escapeHtml(entry.clientIP)}</td>
      <td>${escapeHtml(entry.event)}</td>
      <td>${escapeHtml(entry.path)}</td>
      <td><span class="result ${entry.result}">${escapeHtml(entry.result)}</span></td>
    </tr>
  `).join("");
  return `
    <div class="log-toolbar">
      <select id="logEvent">
        <option value="">全部事件</option>
        <option value="access">access</option>
        <option value="download">download</option>
        <option value="upload">upload</option>
        <option value="login">login</option>
        <option value="reject">reject</option>
        <option value="error">error</option>
      </select>
      <input id="logSince" type="datetime-local" />
      <input id="logUntil" type="datetime-local" />
      <button id="refreshLogs" type="button">刷新</button>
      <button id="clearLogs" type="button">清空</button>
      <button id="exportLogs" type="button">导出 CSV</button>
    </div>
    <table class="logs">
      <thead><tr><th>时间</th><th>客户端 IP</th><th>事件</th><th>路径</th><th>结果</th></tr></thead>
      <tbody>${rows || `<tr><td colspan="5" class="empty">暂无访问日志</td></tr>`}</tbody>
    </table>
  `;
}
```

Wire the toolbar in `bindContentEvents`:

```ts
document.querySelector("#refreshLogs")?.addEventListener("click", async () => {
  logFilter = readLogFilter();
  logs = await service.GetLogs(logFilter);
  render();
});

document.querySelector("#clearLogs")?.addEventListener("click", async () => {
  if (!confirm("确定清空访问日志？")) return;
  await service.ClearLogs();
  logs = [];
  render();
});

document.querySelector("#exportLogs")?.addEventListener("click", async () => {
  await service.ExportLogs(logFilter);
});
```

- [ ] **Step 3: Verify logs**

Run:

```bash
go test -count=1 ./internal/control ./internal/accesslog
cd frontend && npm run build && cd ..
go test -count=1 ./...
git diff --check
```

Expected: PASS and no diff check output.

- [ ] **Step 4: Commit task 5**

```bash
git add internal/control frontend
git commit -m "feat: add access logs panel"
```

## Task 6: Desktop Packaging And Documentation

**Files:**
- Modify: `README.md`
- Modify: `.gitignore`
- Create or Modify: `wails.json`

- [ ] **Step 1: Add ignore rules**

Ensure `.gitignore` contains:

```gitignore
frontend/node_modules/
bin/
build/bin/
```

Keep `frontend/dist` out of `.gitignore` because `frontend/assets.go` embeds it for desktop builds. Commit `frontend/dist` only after `npm run build` produces deterministic output.

- [ ] **Step 2: Add Wails project config**

Create `wails.json` and keep commands aligned with `frontend/package.json`:

```json
{
  "name": "OneTiny",
  "outputfilename": "OneTiny",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev"
}
```

- [ ] **Step 3: Document GUI usage**

Add to `README.md` under usage docs:

```markdown
### GUI 控制面板

桌面版控制面板用于本机管理共享服务。打开 GUI 后不会自动开始共享，点击“启动共享”后才会启动 HTTP 服务。

控制面板支持：

- 查看访问地址和运行状态。
- 启动、停止共享服务。
- 修改共享目录、上传开关、登录开关和最大访问层级，并在服务运行中立即生效。
- 修改端口时弹出确认，确认后重启共享服务。
- 未配置账号密码时，开启登录保护会先要求设置账号密码。
- 关闭窗口时隐藏到托盘；托盘左键打开面板，右键菜单包含“打开面板”和“退出”。
- 查看、筛选、清空和导出访问日志。
```

- [ ] **Step 4: Full verification**

Run:

```bash
go test -count=1 ./...
go test -race -count=1 ./internal/control ./internal/runtimeconf ./internal/server ./internal/server/middleware ./internal/handle/core ./internal/handle/secure ./internal/accesslog ./internal/security ./internal/conf ./cmd
cd frontend && npm run build && cd ..
go build ./...
wails3 build
git diff --stat
git diff --check
git status --short
```

Expected:
- Go tests pass.
- Frontend build passes.
- `go build ./...` passes.
- `wails3 build` produces a desktop binary.
- Diff covers GUI integration, docs, and Wails/frontend config.
- `git diff --check` has no output.

- [ ] **Step 5: Manual desktop QA**

Run the built app on the current OS and verify:

- Initial window is visible and service is stopped.
- `启动共享` starts service and displays `http://IP:port`.
- `复制地址` places the address on clipboard.
- Shared directory change affects the next browser request.
- Upload switch affects upload handler without restart.
- Login switch opens credential dialog when credentials are missing.
- Saving username/password writes bcrypt config and enables login when requested.
- Port change while running asks confirmation and restarts.
- Access logs show access/login/download/upload rows.
- Close button hides to tray instead of exiting.
- Tray left click opens/focuses the panel.
- Tray right-click menu has `打开面板` and `退出`.
- Exiting while service is running stops the service before app quit.

- [ ] **Step 6: Commit task 6**

```bash
git add README.md .gitignore wails.json frontend cmd/onetiny-gui internal/gui go.mod go.sum
git commit -m "docs: add GUI usage and desktop build config"
```

## Final Review

- [ ] **Step 1: Final full test**

Run:

```bash
go test -count=1 ./...
go build ./...
cd frontend && npm run build && cd ..
wails3 build
git diff --check
git status --short
```

Expected: all commands pass, diff check clean, status clean after final commit.

- [ ] **Step 2: Final code review**

Dispatch a final reviewer with:

```text
Review Wails desktop control panel implementation against docs/superpowers/specs/2026-06-13-gui-control-panel-design.md and docs/superpowers/plans/2026-06-13-wails-desktop-control-panel.md.
Focus on lifecycle correctness, runtime config hot updates, credential handling, Wails tray/window behavior, frontend UX regressions, and packaging risks.
```

- [ ] **Step 3: Fix Critical/Important review issues**

Apply reviewer feedback if it is technically correct. Re-run:

```bash
go test -count=1 ./...
go build ./...
cd frontend && npm run build && cd ..
wails3 build
```

- [ ] **Step 4: Commit final fixes**

```bash
git add .
git commit -m "fix: finalize desktop control panel"
```

Only run this commit if final review required code changes.

## Self-Review

- Spec coverage: covers GUI start/stop, hot config, port restart confirmation, credential setup, security tab, access logs, about tab, tray close/open/quit behavior, docs, and packaging.
- Known risks handled in plan: `RunCore` lifecycle split is closed in Task 1; runtime credential hot update is closed in Task 1; Wails v3 API drift is isolated to `internal/gui`.
- Plan scan: implementation markers are concrete. The Wails close hook is represented by concrete `RegisterHook(events.Common.WindowClosing, ...)` behavior in Task 3.

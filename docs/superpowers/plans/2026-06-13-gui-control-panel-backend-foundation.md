# GUI Control Panel Backend Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the backend foundation required by the OneTiny desktop GUI: bcrypt credentials, shared config validation, service lifecycle control, runtime hot updates, and persisted access logs.

**Architecture:** This plan does not implement Wails UI yet. It first extracts backend services that both CLI and future GUI can call. Existing Gin routes and templates remain, but server startup becomes controllable through a `ServiceManager` instead of only `server.RunCore()`.

**Tech Stack:** Go 1.26, Gin, Viper, urfave/cli, bcrypt from `golang.org/x/crypto/bcrypt`, JSON Lines access logs.

---

## Scope Check

The approved GUI spec covers several subsystems: backend service control, credential migration, access logging, and Wails desktop UI. This plan intentionally implements only the backend foundation. A later plan should cover the Wails v3 desktop shell and frontend pages after this backend is tested.

## File Structure

- Create `internal/security/credentials.go`: bcrypt credential hashing, verification, and legacy MD5 detection helpers.
- Create `internal/security/credentials_test.go`: unit tests for bcrypt and legacy config detection.
- Create `internal/conf/store.go`: config validation and credential config read/write helpers around the existing Viper-backed config.
- Create `internal/conf/store_test.go`: config validation tests using temporary config files.
- Modify `internal/conf/conf.go`: load new bcrypt fields and reject invalid protected legacy credentials.
- Modify `cmd/secure.go`: CLI `sec` writes bcrypt credentials through `internal/security`.
- Modify `internal/handle/secure/login.go`: Web login verifies bcrypt credentials.
- Create `internal/runtimeconf/config.go`: thread-safe runtime config snapshot for hot updates.
- Create `internal/server/manager.go`: start, stop, restart, status, and hot-update Gin server manager.
- Modify `internal/server/server.go`: keep current `RunCore()` behavior by delegating to `ServiceManager`.
- Modify `internal/server/middleware/check-login.go`: read login switch from runtime config.
- Modify `internal/server/middleware/check-level.go`: read root path and max level from runtime config.
- Modify `internal/handle/core/download.go`: read root path and upload setting through runtime config.
- Create `internal/accesslog/logger.go`: JSON Lines log writer and reader.
- Create `internal/accesslog/logger_test.go`: log write/read/filter tests.
- Create `internal/server/middleware/access_log.go`: request access log middleware and shared logger instance.
- Modify `internal/server/middleware/setup.go`: attach access log middleware.
- Modify `internal/handle/core/upload.go`, `internal/handle/core/download.go`, and `internal/handle/secure/login.go`: record upload/download/login events.

## Task 1: CredentialService With bcrypt

**Files:**
- Create: `internal/security/credentials.go`
- Create: `internal/security/credentials_test.go`
- Modify: `go.mod`

- [ ] **Step 1: Write failing bcrypt tests**

Create `internal/security/credentials_test.go`:

```go
package security

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "correct horse battery staple" {
		t.Fatal("HashPassword returned the plain password")
	}
	if err := VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword rejected valid password: %v", err)
	}
	if err := VerifyPassword(hash, "wrong"); err == nil {
		t.Fatal("VerifyPassword accepted invalid password")
	}
}

func TestCredentialConfigValidation(t *testing.T) {
	valid := CredentialConfig{
		Username:     "admin",
		PasswordHash: "$2a$10$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		HashAlgo:     HashAlgoBcrypt,
	}
	if err := valid.ValidateForSecureMode(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	legacy := CredentialConfig{
		Username:  "21232f297a57a5a743894a0e4a801fc3",
		LegacyMD5: "21232f297a57a5a743894a0e4a801fc3",
	}
	if err := legacy.ValidateForSecureMode(); err == nil {
		t.Fatal("legacy MD5 config was accepted in secure mode")
	}

	missing := CredentialConfig{}
	if err := missing.ValidateForSecureMode(); err == nil {
		t.Fatal("missing credentials were accepted in secure mode")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/security
```

Expected: FAIL because package `internal/security` and the referenced symbols do not exist.

- [ ] **Step 3: Implement bcrypt credential helpers**

Create `internal/security/credentials.go`:

```go
package security

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const HashAlgoBcrypt = "bcrypt"

var (
	ErrMissingCredentials = errors.New("开启访问登录需先设置帐号密码")
	ErrLegacyMD5Config    = errors.New("检测到旧版 MD5 账号密码配置，请重新设置账号密码")
	ErrUnsupportedHash    = errors.New("不支持的密码哈希算法")
)

type CredentialConfig struct {
	Username     string
	PasswordHash string
	HashAlgo     string
	LegacyMD5    string
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (c CredentialConfig) IsConfigured() bool {
	return strings.TrimSpace(c.Username) != "" &&
		strings.TrimSpace(c.PasswordHash) != "" &&
		c.HashAlgo == HashAlgoBcrypt
}

func (c CredentialConfig) HasLegacyMD5() bool {
	return strings.TrimSpace(c.LegacyMD5) != ""
}

func (c CredentialConfig) ValidateForSecureMode() error {
	if c.HasLegacyMD5() {
		return ErrLegacyMD5Config
	}
	if strings.TrimSpace(c.Username) == "" || strings.TrimSpace(c.PasswordHash) == "" {
		return ErrMissingCredentials
	}
	if c.HashAlgo != HashAlgoBcrypt {
		return ErrUnsupportedHash
	}
	return nil
}
```

- [ ] **Step 4: Add bcrypt dependency**

Run:

```bash
go get golang.org/x/crypto/bcrypt
go mod tidy
```

Expected: `go.mod` keeps `golang.org/x/crypto` available and `go.sum` is updated if needed.

- [ ] **Step 5: Verify credential tests pass**

Run:

```bash
go test ./internal/security
```

Expected: PASS.

- [ ] **Step 6: Commit credential service**

```bash
git add go.mod go.sum internal/security/credentials.go internal/security/credentials_test.go
git commit -m "feat: add bcrypt credential service"
```

## Task 2: Shared ConfigStore Validation

**Files:**
- Create: `internal/conf/store.go`
- Create: `internal/conf/store_test.go`
- Modify: `internal/conf/conf.go`

- [ ] **Step 1: Write failing config tests**

Create `internal/conf/store_test.go`:

```go
package conf

import (
	"testing"

	"github.com/spf13/viper"
)

func TestCredentialConfigFromViperReadsBcryptFields(t *testing.T) {
	viper.Reset()
	viper.Set("account.custom.user", "admin")
	viper.Set("account.custom.pass_hash", "$2a$10$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("account.custom.pass_hash_algo", "bcrypt")

	creds := CredentialConfigFromViper()
	if creds.Username != "admin" {
		t.Fatalf("username = %q, want admin", creds.Username)
	}
	if creds.PasswordHash == "" {
		t.Fatal("missing password hash")
	}
	if creds.HashAlgo != "bcrypt" {
		t.Fatalf("hash algo = %q, want bcrypt", creds.HashAlgo)
	}
}

func TestValidateSecureConfigRejectsLegacyMD5(t *testing.T) {
	viper.Reset()
	viper.Set("account.secure", true)
	viper.Set("account.custom.user", "21232f297a57a5a743894a0e4a801fc3")
	viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")

	if err := ValidateSecureConfig(); err == nil {
		t.Fatal("ValidateSecureConfig accepted legacy MD5 secure config")
	}
}

func TestValidateSecureConfigAllowsUnprotectedLegacyConfig(t *testing.T) {
	viper.Reset()
	viper.Set("account.secure", false)
	viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")

	if err := ValidateSecureConfig(); err != nil {
		t.Fatalf("unprotected legacy config rejected: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/conf
```

Expected: FAIL because `CredentialConfigFromViper` and `ValidateSecureConfig` do not exist.

- [ ] **Step 3: Implement config helpers**

Create `internal/conf/store.go`:

```go
package conf

import (
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/spf13/viper"
)

func CredentialConfigFromViper() security.CredentialConfig {
	return security.CredentialConfig{
		Username:     viper.GetString("account.custom.user"),
		PasswordHash: viper.GetString("account.custom.pass_hash"),
		HashAlgo:     viper.GetString("account.custom.pass_hash_algo"),
		LegacyMD5:    viper.GetString("account.custom.pass"),
	}
}

func ValidateSecureConfig() error {
	if !viper.GetBool("account.secure") {
		return nil
	}
	return CredentialConfigFromViper().ValidateForSecureMode()
}

func SetCredentialConfig(username, passwordHash string) {
	viper.Set("account.custom.user", username)
	viper.Set("account.custom.pass_hash", passwordHash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
	viper.Set("account.custom.pass", "")
}
```

- [ ] **Step 4: Wire bcrypt fields into LoadConfig**

Modify `internal/conf/conf.go` inside `LoadConfig()` after existing Viper reads:

```go
	Config.IsSecure = viper.GetBool("account.secure")
	Config.Username = viper.GetString("account.custom.user")
	Config.Password = viper.GetString("account.custom.pass_hash")
	return nil
```

Remove the old final line that assigns `Config.Password` from `account.custom.pass`.

Do not call `ValidateSecureConfig()` from `LoadConfig()`: `LoadConfig()` runs before CLI command parsing, and legacy secure configs must still allow users to run `onetiny sec` to repair credentials. Final secure validation belongs at the service-start path after CLI flags and config are merged.

- [ ] **Step 5: Verify config tests pass**

Run:

```bash
go test ./internal/conf
```

Expected: PASS.

- [ ] **Step 6: Commit config validation**

```bash
git add internal/conf/conf.go internal/conf/store.go internal/conf/store_test.go
git commit -m "feat: validate secure config"
```

## Task 3: CLI And Web Login Use bcrypt

**Files:**
- Modify: `cmd/secure.go`
- Modify: `internal/handle/secure/login.go`
- Test: `cmd/secure_test.go`

- [ ] **Step 1: Write failing CLI credential test**

Create `cmd/secure_test.go`:

```go
package cmd

import (
	"flag"
	"testing"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func TestSecureActionWritesBcryptCredentials(t *testing.T) {
	viper.Reset()
	set := flagSet(t, map[string]string{
		"user":   "admin",
		"pass":   "secret",
		"secure": "true",
	})
	ctx := cli.NewContext(&cli.App{}, set, nil)

	weight, err := secureAction(ctx)
	if err != nil {
		t.Fatalf("secureAction returned error: %v", err)
	}
	if weight != USER|PASS|SECU {
		t.Fatalf("weight = %d, want %d", weight, USER|PASS|SECU)
	}
	if viper.GetString("account.custom.user") != "admin" {
		t.Fatalf("username was not stored as plain username")
	}
	if viper.GetString("account.custom.pass") != "" {
		t.Fatalf("legacy MD5 pass field should be cleared")
	}
	if viper.GetString("account.custom.pass_hash_algo") != security.HashAlgoBcrypt {
		t.Fatalf("missing bcrypt algo")
	}
	if err := security.VerifyPassword(viper.GetString("account.custom.pass_hash"), "secret"); err != nil {
		t.Fatalf("stored hash does not verify: %v", err)
	}

	conf.Config.Username = viper.GetString("account.custom.user")
	conf.Config.Password = viper.GetString("account.custom.pass_hash")
}
```

Also add the helper in the same file:

```go
func flagSet(t *testing.T, values map[string]string) *flag.FlagSet {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("user", "", "")
	fs.String("pass", "", "")
	fs.Bool("secure", false, "")
	for name, value := range values {
		if err := fs.Set(name, value); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}
	return fs
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./cmd -run TestSecureActionWritesBcryptCredentials
```

Expected: FAIL because `secureAction` still writes MD5 fields.

- [ ] **Step 3: Update CLI secure action**

Modify `cmd/secure.go`:

```go
import (
	"errors"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)
```

Replace the user/password handling inside `secureAction`:

```go
	if is, u := c.IsSet("user"), c.String("user"); is && u != "" {
		weight |= USER
		viper.Set("account.custom.user", u)
	}
	if is, p := c.IsSet("pass"), c.String("pass"); is && p != "" {
		weight |= PASS
		hash, err := security.HashPassword(p)
		if err != nil {
			return weight, err
		}
		conf.SetCredentialConfig(viper.GetString("account.custom.user"), hash)
	}
```

Update `Handle` credential checks to use new fields:

```go
func hasConfiguredCredentials() bool {
	return viper.GetString("account.custom.user") != "" &&
		viper.GetString("account.custom.pass_hash") != "" &&
		viper.GetString("account.custom.pass_hash_algo") == security.HashAlgoBcrypt
}
```

Replace old `viper.GetString("account.custom.pass")` checks with `hasConfiguredCredentials()`.

After CLI flags and config are merged, validate secure credentials before starting the server. This final validation must use the effective secure setting, including `onetiny --secure`; enabling secure with missing, legacy MD5, unsupported, or invalid bcrypt credentials should fail before the HTTP server starts.

- [ ] **Step 4: Update Web login verification**

Modify `internal/handle/secure/login.go`:

```go
import (
	"net/http"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)
```

Replace MD5 comparison with:

```go
	if c.PostForm("username") == conf.Config.Username &&
		security.VerifyPassword(conf.Config.Password, c.PostForm("password")) == nil {
		session := sessions.Default(c)
		session.Set("login", conf.Config.SessionVal)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "登录成功"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "登录失败"})
```

- [ ] **Step 5: Verify CLI and login compile**

Run:

```bash
go test ./cmd ./internal/handle/secure
```

Expected: PASS.

- [ ] **Step 6: Commit bcrypt CLI and login**

```bash
git add cmd/secure.go cmd/secure_test.go internal/handle/secure/login.go
git commit -m "feat: use bcrypt for login credentials"
```

## Task 4: Runtime Config And ServiceManager

**Files:**
- Create: `internal/runtimeconf/config.go`
- Create: `internal/runtimeconf/config_test.go`
- Create: `internal/server/manager.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Write failing runtime config test**

Create `internal/runtimeconf/config_test.go`:

```go
package runtimeconf

import "testing"

func TestRuntimeConfigUpdate(t *testing.T) {
	cfg := NewRuntimeConfig(ConfigSnapshot{
		RootPath:      "/tmp/a",
		Port:          8192,
		MaxLevel:      1,
		IsAllowUpload: false,
		IsSecure:      false,
	})

	cfg.Update(ConfigPatch{RootPath: ptrString("/tmp/b"), IsAllowUpload: ptrBool(true)})
	got := cfg.Snapshot()
	if got.RootPath != "/tmp/b" {
		t.Fatalf("RootPath = %q, want /tmp/b", got.RootPath)
	}
	if !got.IsAllowUpload {
		t.Fatal("IsAllowUpload was not updated")
	}
	if got.Port != 8192 {
		t.Fatalf("Port changed unexpectedly: %d", got.Port)
	}
}

func ptrString(v string) *string { return &v }
func ptrBool(v bool) *bool { return &v }
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/runtimeconf -run TestRuntimeConfigUpdate
```

Expected: FAIL because package `internal/runtimeconf` does not exist.

- [ ] **Step 3: Implement runtime config in an independent package**

Create `internal/runtimeconf/config.go`:

```go
package runtimeconf

import "sync"

type ConfigSnapshot struct {
	RootPath      string
	Port          int
	MaxLevel      uint8
	IsAllowUpload bool
	IsSecure      bool
	IP            string
}

type ConfigPatch struct {
	RootPath      *string
	Port          *int
	MaxLevel      *uint8
	IsAllowUpload *bool
	IsSecure      *bool
}

type RuntimeConfig struct {
	mu  sync.RWMutex
	cfg ConfigSnapshot
}

func NewRuntimeConfig(cfg ConfigSnapshot) *RuntimeConfig {
	return &RuntimeConfig{cfg: cfg}
}

func (r *RuntimeConfig) Snapshot() ConfigSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg
}

func (r *RuntimeConfig) Update(patch ConfigPatch) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if patch.RootPath != nil {
		r.cfg.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		r.cfg.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		r.cfg.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		r.cfg.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		r.cfg.IsSecure = *patch.IsSecure
	}
}

var Current = NewRuntimeConfig(ConfigSnapshot{})

func SetCurrent(runtime *RuntimeConfig) {
	Current = runtime
}
```

- [ ] **Step 4: Implement service manager**

Create `internal/server/manager.go`:

```go
package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tcp404/OneTiny/internal/runtimeconf"
	"github.com/gin-gonic/gin"
)

var ErrServerAlreadyRunning = errors.New("服务已在运行")
var ErrServerNotRunning = errors.New("服务未运行")

type ServiceManager struct {
	mu      sync.Mutex
	runtime *runtimeconf.RuntimeConfig
	server  *http.Server
}

func NewServiceManager(runtime *runtimeconf.RuntimeConfig) *ServiceManager {
	return &ServiceManager{runtime: runtime}
}

func (m *ServiceManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.server != nil {
		return ErrServerAlreadyRunning
	}
	cfg := m.runtime.Snapshot()
	r := gin.New()
	setupEngine(r)
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: r,
	}
	m.server = srv
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.mu.Lock()
			if m.server == srv {
				m.server = nil
			}
			m.mu.Unlock()
		}
	}()
	return nil
}

func (m *ServiceManager) Stop() error {
	m.mu.Lock()
	srv := m.server
	if srv == nil {
		m.mu.Unlock()
		return ErrServerNotRunning
	}
	m.server = nil
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func (m *ServiceManager) Restart() error {
	if err := m.Stop(); err != nil && !errors.Is(err, ErrServerNotRunning) {
		return err
	}
	return m.Start()
}

func (m *ServiceManager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.server != nil
}

func (m *ServiceManager) Status() runtimeconf.ConfigSnapshot {
	return m.runtime.Snapshot()
}
```

- [ ] **Step 5: Extract shared Gin setup and seed runtime config**

Modify `internal/server/server.go`:

```go
func snapshotFromGlobalConfig() runtimeconf.ConfigSnapshot {
	return runtimeconf.ConfigSnapshot{
		RootPath:      conf.Config.RootPath,
		Port:          conf.Config.Port,
		MaxLevel:      conf.Config.MaxLevel,
		IsAllowUpload: conf.Config.IsAllowUpload,
		IsSecure:      conf.Config.IsSecure,
		IP:            conf.Config.IP,
	}
}

func setupEngine(r *gin.Engine) {
	middleware.Setup(r)
	routes.Setup(r)
}

func initServer() *http.Server {
	gin.SetMode(gin.ReleaseMode)
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(snapshotFromGlobalConfig()))
	r := gin.New()
	setupEngine(r)
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Config.Port),
		Handler: r,
	}
	return s
}
```

Add `github.com/tcp404/OneTiny/internal/runtimeconf` to the imports.

- [ ] **Step 6: Verify runtime and server tests pass**

Run:

```bash
go test ./internal/runtimeconf ./internal/server
```

Expected: PASS.

- [ ] **Step 7: Commit service manager**

```bash
git add internal/runtimeconf internal/server/manager.go internal/server/server.go
git commit -m "feat: add service manager"
```

## Task 5: Wire RuntimeConfig Into Middleware And Handlers

**Files:**
- Modify: `internal/server/middleware/check-login.go`
- Modify: `internal/server/middleware/check-level.go`
- Modify: `internal/handle/core/download.go`
- Modify: `internal/handle/core/upload.go`

- [ ] **Step 1: Update login middleware**

Modify `internal/server/middleware/check-login.go`:

```go
import (
	"net/http"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)
```

Replace the secure check with:

```go
	if !runtimeconf.Current.Snapshot().IsSecure {
		return
	}
```

- [ ] **Step 2: Update level middleware path lookups**

Modify `internal/server/middleware/check-level.go` to read a snapshot once:

```go
func CheckLevel(c *gin.Context) {
	filePath := strings.TrimPrefix(c.Param("filename"), constant.FileGroupPrefix)
	c.Set("filename", filePath)

	cfg := runtimeconf.Current.Snapshot()
	isD := isDir(cfg.RootPath, filePath)
	c.Set("isDirectory", isD)
	isFile := !isD
	if isOverLevel(cfg.RootPath, cfg.MaxLevel, filePath, isFile, c.Query("action") == "dl") {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
	}
}

func isDir(rootPath, filePath string) bool {
	if filePath == constant.ROOT {
		return true
	}
	fInfo, _ := os.Stat(path.Join(rootPath, filePath))
	return fInfo != nil && fInfo.IsDir()
}

func isOverLevel(rootPath string, maxLevel uint8, relPath string, isFile bool, isDl bool) bool {
	rel, _ := filepath.Rel(rootPath, filepath.Join(rootPath, relPath))
	i := strings.Split(rel, string(filepath.Separator))
	level := len(i)
	if i[0] == "." {
		level = 0
	}
	if isFile || isDl {
		level--
	}
	return level > int(maxLevel)
}
```

Add `github.com/tcp404/OneTiny/internal/runtimeconf` to imports and remove `internal/conf`.

- [ ] **Step 3: Update download handler root path**

In `internal/handle/core/download.go`, inside `Downloader`:

```go
	cfg := runtimeconf.Current.Snapshot()
	road := c.GetString("filename")
	a := &agent{
		abs:    filepath.Join(cfg.RootPath, road),
		rel:    road,
		action: c.Query("action"),
		isDir:  c.GetBool("isDirectory"),
	}
```

Update `readDir`, `getFileInfos`, and directory zip path to use `runtimeconf.Current.Snapshot().RootPath` instead of `conf.Config.RootPath`.

- [ ] **Step 4: Update upload handler**

In `internal/handle/core/upload.go`, read runtime config:

```go
	cfg := runtimeconf.Current.Snapshot()
	if !cfg.IsAllowUpload {
		handle.ErrorHandle(c, "当前未开启上传")
		return
	}
```

Use `cfg.RootPath` for destination path.

- [ ] **Step 5: Verify all packages compile**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit runtime wiring**

```bash
git add internal/runtimeconf internal/server internal/server/middleware internal/handle/core
git commit -m "feat: apply runtime config to handlers"
```

## Task 6: AccessLogger And Middleware

**Files:**
- Create: `internal/accesslog/logger.go`
- Create: `internal/accesslog/logger_test.go`
- Modify: `internal/server/middleware/setup.go`
- Modify: `internal/handle/secure/login.go`
- Modify: `internal/handle/core/download.go`
- Modify: `internal/handle/core/upload.go`

- [ ] **Step 1: Write failing access log tests**

Create `internal/accesslog/logger_test.go`:

```go
package accesslog

import (
	"path/filepath"
	"testing"
)

func TestLoggerWritesAndReadsJSONLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	logger := New(path)
	err := logger.Write(Event{
		ClientIP: "192.168.1.23",
		Method:   "GET",
		Event:    EventDownload,
		Path:     "/file/a.txt",
		Status:   200,
		Result:   ResultSuccess,
	})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	events, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].Event != EventDownload {
		t.Fatalf("event = %q, want %q", events[0].Event, EventDownload)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/accesslog
```

Expected: FAIL because package does not exist.

- [ ] **Step 3: Implement JSON Lines logger**

Create `internal/accesslog/logger.go`:

```go
package accesslog

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	EventAccess   = "access"
	EventDownload = "download"
	EventUpload   = "upload"
	EventLogin    = "login"
	EventReject   = "reject"
	EventError    = "error"

	ResultSuccess = "success"
	ResultFailure = "failure"
	ResultReject  = "reject"
)

type Event struct {
	Time     time.Time `json:"time"`
	ClientIP string    `json:"client_ip"`
	Method   string    `json:"method"`
	Event    string    `json:"event"`
	Path     string    `json:"path"`
	Status   int       `json:"status"`
	Result   string    `json:"result"`
}

type Filter struct {
	Event string
	Since time.Time
	Until time.Time
}

type Logger struct {
	path string
}

func DefaultPath() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", "access.log")
	}
	return filepath.Join(cfgDir, "tiny", "access.log")
}

func New(path string) *Logger {
	return &Logger{path: path}
}

func (l *Logger) Write(event Event) error {
	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	if err := os.MkdirAll(filepath.Dir(l.path), os.ModePerm); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(event)
}

func (l *Logger) Read(filter Filter) ([]Event, error) {
	f, err := os.Open(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if filter.Event != "" && event.Event != filter.Event {
			continue
		}
		if !filter.Since.IsZero() && event.Time.Before(filter.Since) {
			continue
		}
		if !filter.Until.IsZero() && event.Time.After(filter.Until) {
			continue
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

func (l *Logger) Clear() error {
	if err := os.MkdirAll(filepath.Dir(l.path), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(l.path, nil, 0o644)
}
```

- [ ] **Step 4: Add middleware helper**

In `internal/server/middleware/setup.go`, add middleware after logger setup:

```go
func Setup(r *gin.Engine) *gin.Engine {
	r.Use(Logger(), gin.Recovery())
	r.Use(AccessLog())
	r.Use(enableCookieSession())
	return r
}
```

Create `internal/server/middleware/access_log.go`:

```go
package middleware

import (
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/gin-gonic/gin"
)

var AccessLogger = accesslog.New(accesslog.DefaultPath())

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		event := accesslog.Event{
			ClientIP: c.ClientIP(),
			Method:   c.Request.Method,
			Event:    accesslog.EventAccess,
			Path:     c.Request.URL.Path,
			Status:   c.Writer.Status(),
			Result:   accesslog.ResultSuccess,
		}
		if c.Writer.Status() >= 400 {
			event.Result = accesslog.ResultFailure
		}
		_ = AccessLogger.Write(event)
	}
}
```

- [ ] **Step 5: Record login success and failure events**

Modify `internal/handle/secure/login.go` to write semantic login logs:

```go
func LoginPost(c *gin.Context) {
	success := c.PostForm("username") == conf.Config.Username &&
		security.VerifyPassword(conf.Config.Password, c.PostForm("password")) == nil
	result := accesslog.ResultFailure
	code := 0
	message := "登录失败"
	if success {
		session := sessions.Default(c)
		session.Set("login", conf.Config.SessionVal)
		session.Save()
		result = accesslog.ResultSuccess
		code = 1
		message = "登录成功"
	}
	_ = middleware.AccessLogger.Write(accesslog.Event{
		ClientIP: c.ClientIP(),
		Method:   c.Request.Method,
		Event:    accesslog.EventLogin,
		Path:     c.Request.URL.Path,
		Status:   http.StatusOK,
		Result:   result,
	})
	c.JSON(http.StatusOK, gin.H{"code": code, "message": message})
}
```

Add imports:

```go
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/server/middleware"
```

- [ ] **Step 6: Record download events**

In `internal/handle/core/download.go`, at the start of `func (a *agent) file(c *gin.Context)` after the file opens successfully:

```go
	_ = middleware.AccessLogger.Write(accesslog.Event{
		ClientIP: c.ClientIP(),
		Method:   c.Request.Method,
		Event:    accesslog.EventDownload,
		Path:     c.Request.URL.Path,
		Status:   http.StatusOK,
		Result:   accesslog.ResultSuccess,
	})
```

Add imports:

```go
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/server/middleware"
```

- [ ] **Step 7: Record upload success and rejection events**

In `internal/handle/core/upload.go`, when upload is disabled:

```go
	_ = middleware.AccessLogger.Write(accesslog.Event{
		ClientIP: c.ClientIP(),
		Method:   c.Request.Method,
		Event:    accesslog.EventUpload,
		Path:     c.Request.URL.Path,
		Status:   http.StatusForbidden,
		Result:   accesslog.ResultReject,
	})
```

After a successful upload response:

```go
	_ = middleware.AccessLogger.Write(accesslog.Event{
		ClientIP: c.ClientIP(),
		Method:   c.Request.Method,
		Event:    accesslog.EventUpload,
		Path:     c.Request.URL.Path,
		Status:   http.StatusOK,
		Result:   accesslog.ResultSuccess,
	})
```

Add imports:

```go
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/server/middleware"
```

- [ ] **Step 8: Verify logging tests and compile**

Run:

```bash
go test ./internal/accesslog ./internal/server/middleware ./internal/handle/secure ./internal/handle/core
```

Expected: PASS.

- [ ] **Step 9: Commit access logging**

```bash
git add internal/accesslog internal/server/middleware internal/handle/secure internal/handle/core
git commit -m "feat: add persisted access logs"
```

## Task 7: Full Verification And Documentation Note

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add migration note to README**

Add a short section under usage docs:

```markdown
### 登录配置迁移说明

新版本不再使用 MD5 保存访问登录密码。`onetiny sec` 会写入 bcrypt 密码哈希。

如果旧配置中已经开启访问登录，并且仍然存在旧版 MD5 密码字段，服务会拒绝启动并提示重新设置账号密码：

```bash
onetiny sec -u=你的账号 -p=你的密码 -s
```
```

- [ ] **Step 2: Run full test suite**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Build all packages**

Run:

```bash
go build ./...
```

Expected: PASS with no output.

- [ ] **Step 4: Inspect diff**

Run:

```bash
git diff --stat
git diff --check
```

Expected: diff covers only backend foundation changes and README; `git diff --check` has no output.

- [ ] **Step 5: Commit verification/docs**

```bash
git add README.md
git commit -m "docs: document credential migration"
```

## Handoff To Next Plan

After this backend plan is complete and tests pass, write the next plan for Wails desktop integration. That plan should use these backend methods as stable boundaries:

- service status
- start/stop/restart
- config update patch
- directory chooser adapter
- credential setup
- log read/filter/clear/export

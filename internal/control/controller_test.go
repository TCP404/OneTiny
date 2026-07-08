package control

import (
	"encoding/csv"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/tcp404/OneTiny/internal/server"
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
	conf.Config.IP = "127.0.0.1"

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
	if runtimeconf.Current() != nil {
		t.Fatalf("runtimeconf.Current() = %p, want nil after stop", runtimeconf.Current())
	}
}

func TestControllerUpdateConfigHotAndPortRestart(t *testing.T) {
	resetControllerTest(t)
	root := t.TempDir()
	conf.Config.RootPath = root
	port := freeControlTestPort(t)
	conf.Config.Port = port
	controller := NewController()

	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()

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
	if got := controller.manager.Status(); got.RootPath != nextRoot || !got.IsAllowUpload {
		t.Fatalf("runtime status = %+v, want hot update applied", got)
	}

	nextPort := freeControlTestPort(t)
	status, err = controller.UpdateConfig(ConfigPatchDTO{Port: &nextPort})
	if err != nil {
		t.Fatalf("UpdateConfig port returned error: %v", err)
	}
	if !status.PortRestartRequired || status.Config.Port != nextPort {
		t.Fatalf("port update status = %+v, want restart required", status)
	}
	if got := controller.manager.Status().Port; got != port {
		t.Fatalf("runtime port = %d, want active old port %d", got, port)
	}
}

func TestControllerStoppedPortUpdateAppliesBeforeStart(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	newPort := freeControlTestPort(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	controller := NewController()

	status, err := controller.UpdateConfig(ConfigPatchDTO{Port: &newPort})
	if err != nil {
		t.Fatalf("UpdateConfig stopped port returned error: %v", err)
	}
	if status.PortRestartRequired {
		t.Fatalf("PortRestartRequired = true, want false while stopped")
	}
	if status.Config.Port != newPort {
		t.Fatalf("status port = %d, want new port %d", status.Config.Port, newPort)
	}
	if got := controller.manager.Status().Port; got != newPort {
		t.Fatalf("runtime port = %d, want new port %d", got, newPort)
	}

	status, err = controller.StartSharing()
	if err != nil {
		t.Fatalf("StartSharing after stopped port update returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()
	if status.Address != "http://127.0.0.1:"+strconv.Itoa(newPort) {
		t.Fatalf("started address = %q, want new port %d", status.Address, newPort)
	}
	if !waitTCPReachable("127.0.0.1", newPort, true) {
		t.Fatalf("new port %d did not become reachable", newPort)
	}
	if waitTCPReachable("127.0.0.1", oldPort, false) {
		t.Fatalf("old port %d should not become reachable", oldPort)
	}
}

func TestControllerPendingPortKeepsActiveAddressUntilConfirmed(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	newPort := freeControlTestPort(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	controller := NewController()

	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()
	if !waitTCPReachable("127.0.0.1", oldPort, true) {
		t.Fatalf("old port %d did not become reachable", oldPort)
	}

	status, err := controller.UpdateConfig(ConfigPatchDTO{Port: &newPort})
	if err != nil {
		t.Fatalf("UpdateConfig pending port returned error: %v", err)
	}
	if !status.PortRestartRequired {
		t.Fatalf("PortRestartRequired = false, want true")
	}
	if status.Config.Port != newPort {
		t.Fatalf("status Config.Port = %d, want pending port %d", status.Config.Port, newPort)
	}
	if status.Address != "http://127.0.0.1:"+strconv.Itoa(oldPort) {
		t.Fatalf("status Address = %q, want old active port %d", status.Address, oldPort)
	}
	if got := controller.manager.Status().Port; got != oldPort {
		t.Fatalf("active runtime port = %d, want old port %d", got, oldPort)
	}
	if !waitTCPReachable("127.0.0.1", oldPort, true) {
		t.Fatalf("old port %d should remain reachable", oldPort)
	}
	if waitTCPReachable("127.0.0.1", newPort, false) {
		t.Fatalf("new pending port %d should not be reachable before restart", newPort)
	}
}

func TestControllerStopClearsPendingPortAndNextStartUsesSavedPort(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	newPort := freeControlTestPort(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	controller := NewController()

	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	if !waitTCPReachable("127.0.0.1", oldPort, true) {
		t.Fatalf("old port %d did not become reachable", oldPort)
	}
	status, err := controller.UpdateConfig(ConfigPatchDTO{Port: &newPort})
	if err != nil {
		t.Fatalf("UpdateConfig pending port returned error: %v", err)
	}
	if !status.PortRestartRequired {
		t.Fatalf("PortRestartRequired = false, want true before stop")
	}
	if got := controller.manager.Status().Port; got != oldPort {
		t.Fatalf("runtime port before stop = %d, want old port %d", got, oldPort)
	}

	status, err = controller.StopSharing()
	if err != nil {
		t.Fatalf("StopSharing returned error: %v", err)
	}
	if status.PortRestartRequired {
		t.Fatalf("PortRestartRequired = true, want false after stop")
	}
	if status.Config.Port != newPort {
		t.Fatalf("status port after stop = %d, want saved port %d", status.Config.Port, newPort)
	}
	if got := controller.manager.Status().Port; got != newPort {
		t.Fatalf("runtime port after stop = %d, want saved port %d", got, newPort)
	}

	status, err = controller.StartSharing()
	if err != nil {
		t.Fatalf("StartSharing after pending stop returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()
	if status.Address != "http://127.0.0.1:"+strconv.Itoa(newPort) {
		t.Fatalf("started address = %q, want new port %d", status.Address, newPort)
	}
	if !waitTCPReachable("127.0.0.1", newPort, true) {
		t.Fatalf("new port %d did not become reachable", newPort)
	}
	if waitTCPReachable("127.0.0.1", oldPort, false) {
		t.Fatalf("old port %d should not remain reachable", oldPort)
	}
}

func TestControllerPendingPortConfirmRestartsEvenWhenConfAlreadyHasPort(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	newPort := freeControlTestPort(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	controller := NewController()

	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()
	if _, err := controller.UpdateConfig(ConfigPatchDTO{Port: &newPort}); err != nil {
		t.Fatalf("UpdateConfig pending port returned error: %v", err)
	}
	if conf.Config.Port != newPort {
		t.Fatalf("conf.Config.Port = %d, want pending port persisted %d", conf.Config.Port, newPort)
	}

	status, err := controller.UpdateConfig(ConfigPatchDTO{
		Port:        &newPort,
		RestartPort: true,
	})
	if err != nil {
		t.Fatalf("UpdateConfig confirmed port returned error: %v", err)
	}
	if status.PortRestartRequired {
		t.Fatalf("PortRestartRequired = true, want false after confirmed restart")
	}
	if status.Config.Port != newPort || status.Address != "http://127.0.0.1:"+strconv.Itoa(newPort) {
		t.Fatalf("confirmed status = %+v, want active new port %d", status, newPort)
	}
	if got := controller.manager.Status().Port; got != newPort {
		t.Fatalf("active runtime port = %d, want new port %d", got, newPort)
	}
	if !waitTCPReachable("127.0.0.1", newPort, true) {
		t.Fatalf("new port %d did not become reachable", newPort)
	}
	if waitTCPReachable("127.0.0.1", oldPort, false) {
		t.Fatalf("old port %d should not remain reachable after restart", oldPort)
	}
}

func TestControllerConfirmedPortFailureKeepsOldServiceRunning(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	blockedPort := freeControlTestPort(t)
	blocker, err := net.Listen("tcp", ":"+strconv.Itoa(blockedPort))
	if err != nil {
		t.Fatalf("block new port: %v", err)
	}
	defer blocker.Close()

	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	controller := NewController()
	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()
	if _, err := controller.UpdateConfig(ConfigPatchDTO{Port: &blockedPort}); err != nil {
		t.Fatalf("UpdateConfig pending port returned error: %v", err)
	}

	status, err := controller.UpdateConfig(ConfigPatchDTO{
		Port:        &blockedPort,
		RestartPort: true,
	})
	if err == nil {
		t.Fatal("confirmed blocked port returned nil error, want bind failure")
	}
	if !status.Running || !status.PortRestartRequired {
		t.Fatalf("failed restart status = %+v, want running with restart still required", status)
	}
	if status.Address != "http://127.0.0.1:"+strconv.Itoa(oldPort) {
		t.Fatalf("status Address = %q, want old active port %d", status.Address, oldPort)
	}
	if status.Config.Port != oldPort {
		t.Fatalf("status Config.Port = %d, want old active port %d after failed restart", status.Config.Port, oldPort)
	}
	if got := controller.manager.Status().Port; got != oldPort {
		t.Fatalf("active runtime port = %d, want old port %d", got, oldPort)
	}
	if !waitTCPReachable("127.0.0.1", oldPort, true) {
		t.Fatalf("old port %d should remain reachable after failed restart", oldPort)
	}
}

func TestControllerConfirmedPortSwitchFailureDoesNotCommitConfigOrRuntime(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	newPort := freeControlTestPort(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	viper.Set("server.port", oldPort)
	if err := os.WriteFile(viper.ConfigFileUsed(), []byte("server:\n  port: "+strconv.Itoa(oldPort)+"\n"), 0o600); err != nil {
		t.Fatalf("write original config: %v", err)
	}
	originalContent, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		t.Fatalf("read original config: %v", err)
	}
	controller := NewController()
	controller.pendingPort = &newPort
	controller.portRestartRequired = true
	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()

	switchErr := errors.New("switch failed after bind")
	previousSwitch := restartServiceWithSnapshot
	restartServiceWithSnapshot = func(manager *server.ServiceManager, snapshot runtimeconf.ConfigSnapshot, commit func() error) error {
		return switchErr
	}
	t.Cleanup(func() {
		restartServiceWithSnapshot = previousSwitch
	})

	status, err := controller.UpdateConfig(ConfigPatchDTO{
		Port:        &newPort,
		RestartPort: true,
	})
	if !errors.Is(err, switchErr) {
		t.Fatalf("UpdateConfig error = %v, want %v", err, switchErr)
	}
	if !status.Running || !status.PortRestartRequired {
		t.Fatalf("failed switch status = %+v, want running with restart required", status)
	}
	if status.Config.Port != oldPort || status.Address != "http://127.0.0.1:"+strconv.Itoa(oldPort) {
		t.Fatalf("failed switch status = %+v, want old active port %d", status, oldPort)
	}
	if conf.Config.Port != oldPort {
		t.Fatalf("conf.Config.Port = %d, want old port %d", conf.Config.Port, oldPort)
	}
	if got := controller.manager.Status().Port; got != oldPort {
		t.Fatalf("runtime port = %d, want old port %d", got, oldPort)
	}
	content, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	if string(content) != string(originalContent) {
		t.Fatalf("config file changed after failed switch:\n%s", string(content))
	}
}

func TestControllerConfirmedPortCommitFailureDoesNotRewriteConfigFile(t *testing.T) {
	resetControllerTest(t)
	oldPort := freeControlTestPort(t)
	newPort := freeControlTestPort(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = oldPort
	conf.Config.IP = "127.0.0.1"
	viper.Set("server.port", oldPort)
	originalContent := []byte("# keep exact bytes\nserver:\n  port: " + strconv.Itoa(oldPort) + "\n")
	if err := os.WriteFile(viper.ConfigFileUsed(), originalContent, 0o600); err != nil {
		t.Fatalf("write original config: %v", err)
	}
	controller := NewController()
	controller.pendingPort = &newPort
	controller.portRestartRequired = true

	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	defer func() { _, _ = controller.StopSharing() }()

	writeErr := errors.New("commit atomic write failed")
	writeCalls := 0
	previousWrite := atomicWriteConfigFile
	atomicWriteConfigFile = func(path string, data []byte) error {
		writeCalls++
		if writeCalls == 1 {
			return writeErr
		}
		if err := os.WriteFile(path, []byte("rollback rewrote file\n"), 0o600); err != nil {
			t.Fatalf("write rollback marker: %v", err)
		}
		return nil
	}
	t.Cleanup(func() {
		atomicWriteConfigFile = previousWrite
	})

	status, err := controller.UpdateConfig(ConfigPatchDTO{RestartPort: true})
	if !errors.Is(err, writeErr) {
		t.Fatalf("UpdateConfig error = %v, want %v", err, writeErr)
	}
	if writeCalls != 1 {
		t.Fatalf("atomicWriteConfigFile calls = %d, want only failed commit write", writeCalls)
	}
	if !status.Running || !status.PortRestartRequired {
		t.Fatalf("failed commit status = %+v, want running with restart required", status)
	}
	if status.Config.Port != oldPort || status.Address != "http://127.0.0.1:"+strconv.Itoa(oldPort) {
		t.Fatalf("failed commit status = %+v, want old active port %d", status, oldPort)
	}
	if conf.Config.Port != oldPort {
		t.Fatalf("conf.Config.Port = %d, want old port %d", conf.Config.Port, oldPort)
	}
	if got := controller.manager.Status().Port; got != oldPort {
		t.Fatalf("runtime port = %d, want old port %d", got, oldPort)
	}
	content, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	if string(content) != string(originalContent) {
		t.Fatalf("config file changed after failed commit:\n%s", string(content))
	}
}

func TestControllerStartSharingRejectsInvalidSecureConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr error
	}{
		{
			name: "legacy md5",
			setup: func() {
				viper.Set("account.custom.user", "admin")
				viper.Set("account.custom.pass", "21232f297a57a5a743894a0e4a801fc3")
			},
			wantErr: security.ErrLegacyMD5Config,
		},
		{
			name: "missing credentials",
			setup: func() {
				viper.Set("account.custom.user", "")
			},
			wantErr: security.ErrMissingCredentials,
		},
		{
			name: "invalid bcrypt",
			setup: func() {
				viper.Set("account.custom.user", "admin")
				viper.Set("account.custom.pass_hash", "not-a-bcrypt-hash")
				viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
			},
			wantErr: security.ErrInvalidPasswordHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetControllerTest(t)
			conf.Config.RootPath = t.TempDir()
			conf.Config.Port = freeControlTestPort(t)
			conf.Config.IsSecure = true
			viper.Set("account.secure", true)
			tt.setup()
			controller := NewController()

			status, err := controller.StartSharing()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("StartSharing error = %v, want %v", err, tt.wantErr)
			}
			if status.Running {
				t.Fatalf("status running = true, want false")
			}
			if controller.manager.Running() {
				t.Fatal("manager running = true, want false")
			}
			if status.LastError == "" {
				t.Fatalf("LastError empty, want validation error")
			}
		})
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
	if conf.Config.Password == "" || conf.Config.Password == "strong-password" {
		t.Fatalf("conf password hash = %q, want non-plaintext bcrypt hash", conf.Config.Password)
	}
	if err := security.VerifyPassword(conf.Config.Password, "strong-password"); err != nil {
		t.Fatalf("stored password hash did not verify: %v", err)
	}
	if got := controller.manager.Status(); got.Username != "admin" || got.PasswordHash != conf.Config.Password || !got.IsSecure {
		t.Fatalf("runtime status = %+v, want credential update", got)
	}
}

func TestControllerSetCredentialsRejectsBlankUsernameAndPassword(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "blank username", username: "  ", password: "strong-password"},
		{name: "blank password", username: "admin", password: "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetControllerTest(t)
			controller := NewController()

			_, err := controller.SetCredentials(CredentialPatchDTO{
				Username:        tt.username,
				Password:        tt.password,
				ConfirmPassword: tt.password,
				EnableSecure:    true,
			})
			if err == nil {
				t.Fatal("SetCredentials returned nil error, want validation failure")
			}
			if conf.Config.Username != "" || conf.Config.Password != "" {
				t.Fatalf("conf credentials = %q/%q, want unchanged empty", conf.Config.Username, conf.Config.Password)
			}
			if got := controller.manager.Status(); got.Username != "" || got.PasswordHash != "" || got.IsSecure {
				t.Fatalf("runtime status = %+v, want unchanged credentials", got)
			}
		})
	}
}

func TestControllerSetCredentialsRejectsConfirmationMismatch(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()

	_, err := controller.SetCredentials(CredentialPatchDTO{
		Username:        "admin",
		Password:        "one",
		ConfirmPassword: "two",
		EnableSecure:    true,
	})
	if !errors.Is(err, ErrPasswordConfirmationMismatch) {
		t.Fatalf("SetCredentials error = %v, want ErrPasswordConfirmationMismatch", err)
	}
	if got := viper.GetString("account.custom.pass_hash"); got != "" {
		t.Fatalf("account.custom.pass_hash = %q, want empty after rejected credentials", got)
	}
}

func TestControllerUpdateConfigRollsBackWhenWriteConfigFails(t *testing.T) {
	resetControllerTest(t)
	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.Port = 9090
	viper.Set("server.road", root)
	viper.Set("server.port", 9090)
	controller := NewController()

	brokenFile := filepath.Join(t.TempDir(), "missing", "config.yml")
	viper.SetConfigFile(brokenFile)
	nextRoot := t.TempDir()
	nextPort := 10090
	_, err := controller.UpdateConfig(ConfigPatchDTO{
		RootPath: &nextRoot,
		Port:     &nextPort,
	})
	if err == nil {
		t.Fatal("UpdateConfig returned nil error, want WriteConfig failure")
	}
	if conf.Config.RootPath != root || conf.Config.Port != 9090 {
		t.Fatalf("conf config = %+v, want original root/port", conf.Config)
	}
	if got := controller.manager.Status(); got.RootPath != root || got.Port != 9090 {
		t.Fatalf("runtime status = %+v, want original root/port", got)
	}
	if got := viper.GetString("server.road"); got != root {
		t.Fatalf("viper server.road = %q, want %q", got, root)
	}
	if got := viper.GetInt("server.port"); got != 9090 {
		t.Fatalf("viper server.port = %d, want 9090", got)
	}
}

func TestControllerUpdateConfigAtomicWriteFailureKeepsConfigFileContent(t *testing.T) {
	resetControllerTest(t)
	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.Port = 9090
	viper.Set("server.road", root)
	viper.Set("server.port", 9090)
	originalContent := []byte("server:\n  road: " + root + "\n  port: 9090\n")
	if err := os.WriteFile(viper.ConfigFileUsed(), originalContent, 0o600); err != nil {
		t.Fatalf("write original config: %v", err)
	}
	controller := NewController()

	writeErr := errors.New("atomic write failed")
	previousWrite := atomicWriteConfigFile
	atomicWriteConfigFile = func(path string, data []byte) error {
		if err := os.WriteFile(path+".tmp", data, 0o600); err != nil {
			t.Fatalf("write temp probe: %v", err)
		}
		return writeErr
	}
	t.Cleanup(func() {
		atomicWriteConfigFile = previousWrite
	})

	nextRoot := t.TempDir()
	_, err := controller.UpdateConfig(ConfigPatchDTO{RootPath: &nextRoot})
	if !errors.Is(err, writeErr) {
		t.Fatalf("UpdateConfig error = %v, want %v", err, writeErr)
	}
	content, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	if string(content) != string(originalContent) {
		t.Fatalf("config file content changed after failed atomic write:\n%s", string(content))
	}
	if conf.Config.RootPath != root || controller.manager.Status().RootPath != root {
		t.Fatalf("state changed after failed atomic write: conf=%q runtime=%q", conf.Config.RootPath, controller.manager.Status().RootPath)
	}
}

func TestControllerSetCredentialsRollsBackWhenWriteConfigFails(t *testing.T) {
	resetControllerTest(t)
	hash, err := security.HashPassword("old-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.Config.Username = "old"
	conf.Config.Password = hash
	conf.Config.IsSecure = false
	viper.Set("account.secure", false)
	viper.Set("account.custom.user", "old")
	viper.Set("account.custom.pass_hash", hash)
	viper.Set("account.custom.pass_hash_algo", security.HashAlgoBcrypt)
	controller := NewController()

	brokenFile := filepath.Join(t.TempDir(), "missing", "config.yml")
	viper.SetConfigFile(brokenFile)
	_, err = controller.SetCredentials(CredentialPatchDTO{
		Username:        "admin",
		Password:        "new-password",
		ConfirmPassword: "new-password",
		EnableSecure:    true,
	})
	if err == nil {
		t.Fatal("SetCredentials returned nil error, want WriteConfig failure")
	}
	if conf.Config.Username != "old" || conf.Config.Password != hash || conf.Config.IsSecure {
		t.Fatalf("conf credentials = %+v, want original values", conf.Config)
	}
	if got := controller.manager.Status(); got.Username != "old" || got.PasswordHash != hash || got.IsSecure {
		t.Fatalf("runtime status = %+v, want original credentials", got)
	}
	if got := viper.GetString("account.custom.user"); got != "old" {
		t.Fatalf("viper account.custom.user = %q, want old", got)
	}
	if got := viper.GetString("account.custom.pass_hash"); got != hash {
		t.Fatalf("viper account.custom.pass_hash changed")
	}
	if got := viper.GetBool("account.secure"); got {
		t.Fatal("viper account.secure = true, want false")
	}
}

func TestControllerSetCredentialsAtomicWriteFailureKeepsConfigFileContent(t *testing.T) {
	resetControllerTest(t)
	originalContent := []byte("account:\n  secure: false\n")
	if err := os.WriteFile(viper.ConfigFileUsed(), originalContent, 0o600); err != nil {
		t.Fatalf("write original config: %v", err)
	}
	controller := NewController()

	writeErr := errors.New("atomic credential write failed")
	previousWrite := atomicWriteConfigFile
	atomicWriteConfigFile = func(path string, data []byte) error {
		if err := os.WriteFile(path+".tmp", data, 0o600); err != nil {
			t.Fatalf("write temp probe: %v", err)
		}
		return writeErr
	}
	t.Cleanup(func() {
		atomicWriteConfigFile = previousWrite
	})

	_, err := controller.SetCredentials(CredentialPatchDTO{
		Username:        "admin",
		Password:        "strong-password",
		ConfirmPassword: "strong-password",
		EnableSecure:    true,
	})
	if !errors.Is(err, writeErr) {
		t.Fatalf("SetCredentials error = %v, want %v", err, writeErr)
	}
	content, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	if string(content) != string(originalContent) {
		t.Fatalf("config file content changed after failed atomic credential write:\n%s", string(content))
	}
	if conf.Config.Username != "" || conf.Config.Password != "" || conf.Config.IsSecure {
		t.Fatalf("conf credentials changed after failed atomic write: %+v", conf.Config)
	}
	if got := controller.manager.Status(); got.Username != "" || got.PasswordHash != "" || got.IsSecure {
		t.Fatalf("runtime credentials changed after failed atomic write: %+v", got)
	}
}

func TestControllerLogsReadClearAndExportCSV(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()
	base := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)

	if err := controller.logger.Write(accesslog.Event{
		Time:     base,
		ClientIP: "127.0.0.1",
		Method:   "POST",
		Event:    accesslog.EventLogin,
		Path:     "/login",
		Status:   200,
		Result:   accesslog.ResultSuccess,
	}); err != nil {
		t.Fatalf("write login log: %v", err)
	}
	if err := controller.logger.Write(accesslog.Event{
		Time:   base.Add(time.Minute),
		Method: "GET",
		Event:  accesslog.EventDownload,
		Path:   "/file/a.txt",
		Status: 200,
		Result: accesslog.ResultSuccess,
	}); err != nil {
		t.Fatalf("write download log: %v", err)
	}

	logs, err := controller.GetLogs(LogFilterDTO{Event: accesslog.EventLogin})
	if err != nil {
		t.Fatalf("GetLogs returned error: %v", err)
	}
	if len(logs) != 1 || logs[0].Path != "/login" {
		t.Fatalf("GetLogs login filter = %+v, want login event", logs)
	}

	exportPath := filepath.Join(t.TempDir(), "nested", "logs.csv")
	if err := controller.ExportLogs(exportPath, LogFilterDTO{}); err != nil {
		t.Fatalf("ExportLogs returned error: %v", err)
	}
	rows, err := readCSVRows(exportPath)
	if err != nil {
		t.Fatalf("read exported CSV: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("exported row count = %d, want header plus two events", len(rows))
	}
	if strings.Join(rows[0], ",") != "time,client_ip,method,event,path,status,result" {
		t.Fatalf("CSV header = %+v", rows[0])
	}

	if err := controller.ClearLogs(); err != nil {
		t.Fatalf("ClearLogs returned error: %v", err)
	}
	afterClear, err := controller.GetLogs(LogFilterDTO{})
	if err != nil {
		t.Fatalf("GetLogs after ClearLogs returned error: %v", err)
	}
	if len(afterClear) != 0 {
		t.Fatalf("logs after ClearLogs = %+v, want empty", afterClear)
	}
}

func TestControllerExportLogsRejectsEmptyPathAndDirectory(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()

	if err := controller.ExportLogs("", LogFilterDTO{}); err == nil {
		t.Fatal("ExportLogs empty path returned nil error")
	}
	if err := controller.ExportLogs(t.TempDir(), LogFilterDTO{}); err == nil {
		t.Fatal("ExportLogs directory path returned nil error")
	}
}

func TestControllerExportLogsOverwritesExistingFile(t *testing.T) {
	resetControllerTest(t)
	controller := NewController()
	exportPath := filepath.Join(t.TempDir(), "logs.csv")
	if err := os.WriteFile(exportPath, []byte("old content"), 0o600); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if err := controller.ExportLogs(exportPath, LogFilterDTO{}); err != nil {
		t.Fatalf("ExportLogs returned error: %v", err)
	}
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("ReadFile exported path: %v", err)
	}
	if strings.Contains(string(content), "old content") {
		t.Fatalf("ExportLogs did not overwrite existing file: %q", string(content))
	}
	if !strings.HasPrefix(string(content), "time,client_ip,method,event,path,status,result\n") {
		t.Fatalf("ExportLogs content = %q, want CSV header", string(content))
	}
}

func resetControllerTest(t *testing.T) {
	t.Helper()

	originalConfig := *conf.Config
	viper.Reset()
	t.Cleanup(func() {
		*conf.Config = originalConfig
		runtimeconf.SetCurrent(nil)
		viper.Reset()
	})

	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("APPDATA", configRoot)
	t.Setenv("HOME", configRoot)
	t.Setenv("USERPROFILE", configRoot)

	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir returned error: %v", err)
	}
	if runtime.GOOS != "darwin" {
		rel, err := filepath.Rel(configRoot, userCfgDir)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("os.UserConfigDir returned %q outside isolated root %q", userCfgDir, configRoot)
		}
	}
	cfgDir := filepath.Join(userCfgDir, "tiny")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("MkdirAll config dir: %v", err)
	}
	cfgFile := filepath.Join(cfgDir, "config.yml")
	if err := os.WriteFile(cfgFile, nil, 0o600); err != nil {
		t.Fatalf("WriteFile config.yml: %v", err)
	}
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yml")
	viper.Set("server.road", "")
	viper.Set("server.port", 0)
	viper.Set("server.max_level", 0)
	viper.Set("server.allow_upload", false)
	viper.Set("account.secure", false)

	conf.Config.RootPath = t.TempDir()
	conf.Config.Port = 0
	conf.Config.MaxLevel = 0
	conf.Config.IsAllowUpload = false
	conf.Config.IsSecure = false
	conf.Config.IP = "127.0.0.1"
	conf.Config.Username = ""
	conf.Config.Password = ""
	conf.Config.SessionVal = "test-session"
	runtimeconf.SetCurrent(nil)
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

func readCSVRows(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return csv.NewReader(file).ReadAll()
}

func waitTCPReachable(host string, port int, want bool) bool {
	deadline := time.Now().Add(500 * time.Millisecond)
	address := net.JoinHostPort(host, strconv.Itoa(port))
	for {
		conn, err := net.DialTimeout("tcp", address, 25*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			if want {
				return true
			}
		} else if !want {
			return false
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

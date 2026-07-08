package gui

import (
	"net"
	"testing"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/control"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
)

type fakeDialogAdapter struct {
	confirmCalls int
	onConfirm    func()
}

func (f *fakeDialogAdapter) ChooseDirectory(string) (string, error) {
	return "", nil
}

func (f *fakeDialogAdapter) ChooseExportPath() (string, error) {
	return "", nil
}

func (f *fakeDialogAdapter) OpenConfigDir() error {
	return nil
}

func (f *fakeDialogAdapter) ConfirmQuitWhileRunning(onConfirm func()) error {
	f.confirmCalls++
	f.onConfirm = onConfirm
	return nil
}

func TestServiceConfirmQuitReturnsTrueWithoutPromptWhenStopped(t *testing.T) {
	resetGUIControlTest(t)
	controller := control.NewController()
	dialogs := &fakeDialogAdapter{}
	service := NewService(controller, dialogs)
	confirmed := false

	if !service.requestQuit(func() { confirmed = true }) {
		t.Fatal("requestQuit() = false, want true when stopped")
	}
	if dialogs.confirmCalls != 0 {
		t.Fatalf("confirm calls = %d, want 0 when stopped", dialogs.confirmCalls)
	}
	if confirmed {
		t.Fatal("confirm callback called for stopped service")
	}
}

func TestServiceConfirmQuitKeepsRunningWhenCanceled(t *testing.T) {
	controller, service, dialogs := runningGUIService(t)
	confirmed := false

	if service.requestQuit(func() { confirmed = true }) {
		t.Fatal("requestQuit() = true, want false while waiting for user confirmation")
	}
	if dialogs.confirmCalls != 1 {
		t.Fatalf("confirm calls = %d, want 1", dialogs.confirmCalls)
	}
	if dialogs.onConfirm == nil {
		t.Fatal("confirm callback not registered")
	}
	if confirmed {
		t.Fatal("confirm callback called before user confirms")
	}
	status, err := controller.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if !status.Running {
		t.Fatal("service stopped after canceled quit, want still running")
	}
}

func TestServiceConfirmQuitStopsSharingWhenConfirmedAsynchronously(t *testing.T) {
	controller, service, dialogs := runningGUIService(t)
	confirmed := false

	if service.requestQuit(func() { confirmed = true }) {
		t.Fatal("requestQuit() = true, want false while confirmation dialog is pending")
	}
	if dialogs.confirmCalls != 1 {
		t.Fatalf("confirm calls = %d, want 1", dialogs.confirmCalls)
	}
	if dialogs.onConfirm == nil {
		t.Fatal("confirm callback not registered")
	}
	status, err := controller.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if !status.Running {
		t.Fatal("service stopped before async confirmation")
	}

	dialogs.onConfirm()
	if !confirmed {
		t.Fatal("confirm callback not called after async confirmation")
	}
	status, err = controller.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if status.Running {
		t.Fatal("service still running after confirmed quit")
	}
}

func runningGUIService(t *testing.T) (*control.Controller, *Service, *fakeDialogAdapter) {
	t.Helper()
	resetGUIControlTest(t)
	conf.Config.RootPath = t.TempDir()
	conf.Config.IP = "127.0.0.1"
	conf.Config.Port = freeGUIPort(t)

	controller := control.NewController()
	if _, err := controller.StartSharing(); err != nil {
		t.Fatalf("StartSharing returned error: %v", err)
	}
	t.Cleanup(func() {
		_, _ = controller.StopSharing()
	})
	dialogs := &fakeDialogAdapter{}
	return controller, NewService(controller, dialogs), dialogs
}

func resetGUIControlTest(t *testing.T) {
	t.Helper()
	original := *conf.Config
	originalRuntime := runtimeconf.Current()
	runtimeconf.SetCurrent(nil)
	conf.Config.IsSecure = false
	conf.Config.Username = ""
	conf.Config.Password = ""
	t.Cleanup(func() {
		runtimeconf.SetCurrent(originalRuntime)
		*conf.Config = original
	})
}

func freeGUIPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on free port: %v", err)
	}
	defer func() { _ = listener.Close() }()
	return listener.Addr().(*net.TCPAddr).Port
}

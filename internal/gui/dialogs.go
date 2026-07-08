package gui

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type DialogAdapter interface {
	ChooseDirectory(current string) (string, error)
	ChooseExportPath() (string, error)
	OpenConfigDir() error
	ConfirmQuitWhileRunning(onConfirm func()) error
}

type WailsDialogAdapter struct {
	app       *application.App
	window    application.Window
	configDir string
}

func NewWailsDialogAdapter(app *application.App, window application.Window, configDir string) *WailsDialogAdapter {
	return &WailsDialogAdapter{app: app, window: window, configDir: configDir}
}

func (d *WailsDialogAdapter) ChooseDirectory(current string) (string, error) {
	dialog := d.app.Dialog.OpenFile().
		SetTitle("选择共享目录").
		SetButtonText("选择").
		CanChooseDirectories(true).
		CanChooseFiles(false)
	if strings.TrimSpace(current) != "" {
		dialog.SetDirectory(current)
	}
	if d.window != nil {
		dialog.AttachToWindow(d.window)
	}
	return dialog.PromptForSingleSelection()
}

func (d *WailsDialogAdapter) ChooseExportPath() (string, error) {
	defaultName := "onetiny-access-" + time.Now().Format("20060102-150405") + ".csv"
	dialog := d.app.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title: "导出访问日志",
	}).
		SetButtonText("导出").
		SetFilename(defaultName).
		AddFilter("CSV 文件", "*.csv").
		CanCreateDirectories(true)
	if d.window != nil {
		dialog.AttachToWindow(d.window)
	}
	return dialog.PromptForSingleSelection()
}

func (d *WailsDialogAdapter) OpenConfigDir() error {
	return openPath(filepath.Clean(d.configDir))
}

func (d *WailsDialogAdapter) ConfirmQuitWhileRunning(onConfirm func()) error {
	dialog := d.app.Dialog.Question().
		SetTitle("退出 OneTiny").
		SetMessage("共享服务仍在运行，退出会停止共享。是否退出？")
	quit := dialog.AddButton("退出").SetAsDefault().OnClick(func() {
		if onConfirm != nil {
			go onConfirm()
		}
	})
	cancel := dialog.AddButton("取消").SetAsCancel()
	dialog.SetDefaultButton(quit)
	dialog.SetCancelButton(cancel)
	if d.window != nil {
		dialog.AttachToWindow(d.window)
	}
	dialog.Show()
	return nil
}

func openPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

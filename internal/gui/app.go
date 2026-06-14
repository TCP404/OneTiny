package gui

import (
	"embed"
	"sync/atomic"

	"github.com/TCP404/OneTiny-cli/internal/control"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
)

func Run(assets embed.FS) error {
	controller := control.NewController()
	service := NewService(controller, nil)
	var quitting atomic.Bool

	var app *application.App
	app = application.New(application.Options{
		Name:        "OneTiny",
		Description: "OneTiny 桌面控制面板",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		Linux: application.LinuxOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		ShouldQuit: func() bool {
			if !quitting.Load() && !service.requestQuit(func() {
				quitting.Store(true)
				app.Quit()
			}) {
				return false
			}
			quitting.Store(true)
			service.shutdown()
			return true
		},
		OnShutdown: service.shutdown,
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:      "OneTiny",
		Width:      960,
		Height:     680,
		MinWidth:   820,
		MinHeight:  560,
		StartState: application.WindowStateNormal,
	})
	service.setDialogAdapter(NewWailsDialogAdapter(app, window))

	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if quitting.Load() {
			return
		}
		event.Cancel()
		window.Hide()
	})

	openPanel := func() {
		window.Show().Focus()
	}
	quit := func() {
		if !service.requestQuit(func() {
			quitting.Store(true)
			app.Quit()
		}) {
			return
		}
		quitting.Store(true)
		app.Quit()
	}

	tray := app.SystemTray.New()
	tray.SetLabel("OneTiny")
	tray.SetTooltip("OneTiny")
	tray.SetTemplateIcon(icons.SystrayMacTemplate)
	tray.SetIcon(icons.SystrayLight)
	tray.SetDarkModeIcon(icons.SystrayDark)
	tray.AttachWindow(window).WindowOffset(8)
	tray.OnClick(openPanel)
	tray.OnRightClick(tray.ShowMenu)

	menu := app.NewMenu()
	menu.Add("打开面板").OnClick(func(_ *application.Context) {
		openPanel()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(_ *application.Context) {
		quit()
	})
	tray.SetMenu(menu)

	return app.Run()
}

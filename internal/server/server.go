package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
	"github.com/tcp404/OneTiny/internal/server/middleware"
	"github.com/tcp404/OneTiny/internal/server/routes"
)

type coreManager interface {
	Start() error
	Stop() error
	Done() <-chan error
}

// RunCore 函数负责启动 gin 实例，开始提供 HTTP 服务
func RunCore() {
	cfg := runtimeconf.NewRuntimeConfig(snapshotFromGlobalConfig())
	manager := NewServiceManager(cfg)
	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(q)

	if runCoreWithManager(manager, q, printInfo) {
		os.Exit(0)
	}
}

func runCoreWithManager(manager coreManager, signalChan <-chan os.Signal, printStartup func()) bool {
	if err := manager.Start(); err != nil {
		log.Println(color.RedString(err.Error()))
		return false
	}
	if printStartup != nil {
		printStartup()
	}

	select {
	case <-signalChan:
		if err := manager.Stop(); err != nil {
			log.Println(color.RedString(err.Error()))
		}
		fmt.Println(color.GreenString("\nbye~"))
		return true
	case err := <-manager.Done():
		if err != nil {
			log.Println(color.RedString(err.Error()))
		}
		return false
	}
}

func setupEngine(r *gin.Engine) *gin.Engine {
	middleware.Setup(r)
	routes.Setup(r)
	return r
}

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

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo() {
	log.SetOutput(color.Output)
	// Print IP information
	if conf.Config.IP != "" {
		log.Printf("Run on   [ %s ]", color.BlueString("http://%s:%d", conf.Config.IP, conf.Config.Port))
	} else {
		log.Printf("%s", color.YellowString("Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]"))
	}

	// Print RootPath information
	log.Printf("Run with [ %s ]", color.BlueString("%s", conf.Config.RootPath))

	// Print Max allow access level
	log.Printf("Allow access level: [ %s ]", color.BlueString("%d", conf.Config.MaxLevel))

	// Print Allow upload Status
	status := color.RedString("%t", conf.Config.IsAllowUpload)
	if conf.Config.IsAllowUpload {
		status = color.GreenString("%t", conf.Config.IsAllowUpload)
	}
	log.Printf("Allow upload: [ %s ]", status)

	// Print Secure status
	status = color.RedString("%t", conf.Config.IsSecure)
	if conf.Config.IsSecure {
		status = color.GreenString("%t", conf.Config.IsSecure)
	}
	log.Printf("Need Login: [ %s ]\n\n", status)
}

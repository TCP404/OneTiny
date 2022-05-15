package core

import (
	"errors"
	"log"
	"net"
	"oneTiny/common/config"
	"oneTiny/core/middleware"
	"oneTiny/core/routes"
	"strconv"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

// RunCore 函数负责启动 gin 实例，开始提供 HTTP 服务
func RunCore() error {
	r := StartUpGin()

	printInfo()

	err := r.Run(":" + strconv.Itoa(config.Port))
	if _, ok := err.(*net.OpError); ok {
		return errors.New(color.RedString("指定的 %d 端口已被占用，请换一个端口号。", config.Port))
	}
	return nil
}

func StartUpGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	middleware.Setup(r)
	routes.Setup(r)
	return r
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo() {
	log.SetOutput(color.Output)
	// Print IP infomation
	if config.IP != "" {
		log.Printf("Run on   [ %s ]", color.BlueString("http://%s:%d", config.IP, config.Port))
	} else {
		log.Printf("%s", color.YellowString("Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]"))
	}

	// Print RootPath infomation
	log.Printf("Run with [ %s ]", color.BlueString("%s", config.RootPath))

	// Print Max allow access level
	log.Printf("Allow access level: [ %s ]", color.BlueString("%d", config.MaxLevel))

	// Print Allow upload Status
	status := color.RedString("%t", config.IsAllowUpload)
	if config.IsAllowUpload {
		status = color.GreenString("%t", config.IsAllowUpload)
	}
	log.Printf("Allow upload: [ %s ]", status)

	// Print Secure status
	status = color.RedString("%t", config.IsSecure)
	if config.IsSecure {
		status = color.GreenString("%t", config.IsSecure)
	}
	log.Printf("Need Login: [ %s ]\n\n", status)
}

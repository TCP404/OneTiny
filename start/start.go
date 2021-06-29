package start

import (
	"log"
	"net"
	"oneTiny/config"
	"oneTiny/controller"
	"oneTiny/middleware"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

// Start 函数负责启动 gin 实例，开始提供 HTTP 服务
func Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.LoggerWithWriter(config.Output),gin.Recovery())
	r.Use(middleware.InterceptICO)
	r.Use(middleware.CheckLevel)

	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)
	r.POST("/upload", controller.Upload)

	printInfo()

	err := r.Run(":" + config.Port)
	if _, ok := err.(*net.OpError); ok {
		log.Fatal(color.RedString("指定的 %s 端口已被占用，请换一个端口号。", config.Port))
	}
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo() {
	log.SetOutput(color.Output)
	// Print IP infomation
	if config.IP != "" {
		log.Printf("Run on   [ %s ]", color.BlueString("http://%s:%s", config.IP, config.Port))
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
	log.Printf("Allow upload : [ %s ]", status)
}

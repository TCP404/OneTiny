package start

import (
	"log"
	"oneTiny/config"
	"oneTiny/controller"
	"oneTiny/util"

	"github.com/gin-gonic/gin"
)

// Start 函数负责启动 gin 实例，开始提供 HTTP 服务
func Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)
	r.POST("/upload", controller.Upload)

	ip, err := util.GetIP()
	if err != nil {
		ip = ""
	}
	printInfo(ip, config.RootPath, config.IsAllowUpload)

	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal(err)
	}
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
// 
// 参数:
// 		ip: 本机的局域网 IP 地址，
// 		path: 共享目录的绝对路径
// 		upload: 是否允许上传的状态
func printInfo(ip, path string, upload bool) {
	// Print IP infomation
	if ip != "" {
		log.Printf("Run on   [ %s ]", util.Renderf([]uint8{util.F_BLUE}, "http://%s:%s", ip, config.Port))
	} else {
		log.Printf("%s", util.Renderf([]uint8{util.F_YELLOW}, "Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]", nil))
	}

	// Print RootPath infomation
	log.Printf("Run with [ %s ]", util.Renderf([]uint8{util.F_BLUE}, "%s", path))

	// Print Allow upload Status
	color := util.F_RED
	if upload {
		color = util.F_GREEN
	}
	log.Printf("Allow upload : [ %s ]", util.Renderf([]uint8{color}, "%t", upload))
}

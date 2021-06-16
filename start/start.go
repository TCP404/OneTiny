package start

import (
	"log"
	"oneTiny/config"
	"oneTiny/controller"
	"oneTiny/util"

	"github.com/gin-gonic/gin"
)

func Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)
	r.POST("/upload", controller.Upload)

	if ip, err := util.GetIP(); err == nil {
		log.Printf("Run on   [ %shttp://%s:%s%s ]", util.GetF(util.F_BLUE), ip, config.Port, util.GetEnd())
	} else {
		log.Printf("%s暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。%s", util.GetF(util.F_RED), util.GetEnd())
	}
	log.Printf("Run with [ %s%s%s ]", util.GetF(util.F_BLUE), config.RootPath, util.GetEnd())
	log.Printf("Allow upload status: [ %s%t%s ]", util.GetF(util.F_RED), config.IsAllowUpload, util.GetEnd())
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal(err)
	}
}

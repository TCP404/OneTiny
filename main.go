package main

import (
	"flag"
	"log"
	"oneTiny/controller"
	"oneTiny/util"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func init() {
	wd, _ := os.Getwd()
	flag.StringVar(&controller.RootPath, "r", wd, "指定对外开放的目录路径")
	flag.StringVar(&controller.Port, "p", controller.PORT, "指定开放的端口")
	flag.BoolVar(&controller.IsAllowUpload, "a", false, "指定是否允许访问者上传")
	flag.Parse()
	controller.RootPath = strings.TrimSuffix(controller.RootPath, controller.SEPARATORS)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)
	r.POST("/upload", controller.Upload)

	if ip, err := util.GetIP(); err == nil {
		log.Printf("Run on   [ http://%s:%s ]", ip, controller.Port)
	} else {
		log.Println("暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。")
	}
	log.Printf("Run with [ %s ]", controller.RootPath)

	if err := r.Run(":" + controller.Port); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"flag"
	"log"
	"os"
	"strings"
	base "tinyServer/controller"
	"tinyServer/util"

	"github.com/gin-gonic/gin"
)

func init() {
	wd, _ := os.Getwd()
	flag.StringVar(&base.RootPath, "r", wd, "指定对外开放的目录路径")
	flag.StringVar(&base.Port, "p", base.PORT, "指定开放的端口")
	flag.Parse()
	base.RootPath = strings.TrimSuffix(base.RootPath, base.SEPARATORS)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.NoRoute(base.NotFound)
	r.GET("/*filename", base.Handler)

	if ip, err := util.GetIP(); err == nil {
		log.Printf("Run on   [ http://%s:%s ]", ip, base.Port)
	} else {
		log.Println("暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。")
	}
	log.Printf("Run with [ %s ]", base.RootPath)
	r.Run(":" + base.Port)
}

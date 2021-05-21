package main

import (
	"log"
	"tinyServer/controller"

	"tinyServer/util"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)
	if ip, ok := util.GetIP(); ok {
		log.Printf("[ Run on http://%s:%s ]", ip, controller.Port)
	}
	r.Run(":" + controller.Port)

}

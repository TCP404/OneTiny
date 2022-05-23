package routes

import (
	controller2 "github.com/TCP404/OneTiny-cli/internal/controller"
	middleware2 "github.com/TCP404/OneTiny-cli/internal/middleware"

	"github.com/gin-gonic/gin"
)

func loadCoreRoute(r *gin.Engine) *gin.RouterGroup {
	fileG := r.Group("/file", middleware2.CheckLogin, middleware2.CheckLevel)
	{
		fileG.GET("/*filename", controller2.Downloader)
		fileG.POST("/upload", controller2.Uploader)
	}
	return fileG
}

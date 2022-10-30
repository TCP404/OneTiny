package routes

import (
	"github.com/TCP404/OneTiny-cli/internal/constant"
	"github.com/TCP404/OneTiny-cli/internal/handle/core"
	"github.com/TCP404/OneTiny-cli/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

func loadCoreRoute(r *gin.Engine) *gin.RouterGroup {
	fileG := r.Group(constant.FileGroupPrefix, middleware.CheckLogin, middleware.CheckLevel)
	{
		fileG.GET("/*filename", core.Downloader)
		fileG.POST("/upload", core.Uploader)
	}
	return fileG
}

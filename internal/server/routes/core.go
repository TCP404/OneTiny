package routes

import (
	"github.com/tcp404/OneTiny/internal/constant"
	"github.com/tcp404/OneTiny/internal/handle/core"
	"github.com/tcp404/OneTiny/internal/server/middleware"

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

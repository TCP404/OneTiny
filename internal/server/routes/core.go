package routes

import (
	"github.com/tcp404/OneTiny/internal/server/handler/file"
	"github.com/tcp404/OneTiny/internal/server/middleware"
	"github.com/tcp404/OneTiny/internal/server/routepath"

	"github.com/gin-gonic/gin"
)

func loadCoreRoute(r *gin.Engine) *gin.RouterGroup {
	fileG := r.Group(routepath.FileGroupPrefix, middleware.CheckLogin, middleware.CheckLevel)
	{
		fileG.GET("/*filename", file.Downloader)
		fileG.POST("/upload", file.Uploader)
	}
	return fileG
}

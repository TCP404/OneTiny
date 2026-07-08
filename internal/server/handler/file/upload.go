package file

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/kit/safepath"
	"github.com/tcp404/OneTiny/internal/server/handler"
	"github.com/tcp404/OneTiny/internal/server/routepath"
	"github.com/tcp404/OneTiny/internal/share"
)

func Uploader(c *gin.Context) {
	cfg := currentSnapshot(c)
	if !cfg.IsAllowUpload {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultReject, http.StatusInternalServerError)
		handler.ErrorHandle(c, "当前未开启上传")
		return
	}

	f, err := c.FormFile("upload_file")
	if err != nil {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handler.ErrorHandle(c, "文件上传失败！")
		return
	}

	currPath := c.PostForm("path")
	filename, ok := share.SafeUploadFilename(f.Filename)
	if !ok {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handler.ErrorHandle(c, "文件保存失败！")
		return
	}
	target, ok := safepath.ResolveCreateWithinRoot(cfg.RootPath, currPath, filename)
	if !ok {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handler.ErrorHandle(c, "文件保存失败！")
		return
	}
	err = c.SaveUploadedFile(f, target)
	if err != nil {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handler.ErrorHandle(c, "文件保存失败！")
		return
	}
	logRequestEvent(c, accesslog.EventUpload, accesslog.ResultSuccess, http.StatusMovedPermanently)
	c.Redirect(http.StatusMovedPermanently, path.Join(routepath.FileGroupPrefix, currPath))
}

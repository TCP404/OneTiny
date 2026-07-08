package core

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/constant"
	"github.com/tcp404/OneTiny/internal/handle"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
)

func Uploader(c *gin.Context) {
	cfg := currentSnapshot(c)
	if !cfg.IsAllowUpload {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultReject, http.StatusInternalServerError)
		handle.ErrorHandle(c, "当前未开启上传")
		return
	}

	f, err := c.FormFile("upload_file")
	if err != nil {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handle.ErrorHandle(c, "文件上传失败！")
		return
	}

	currPath := c.PostForm("path")
	filename, ok := safeUploadFilename(f.Filename)
	if !ok {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handle.ErrorHandle(c, "文件保存失败！")
		return
	}
	target, ok := runtimeconf.ResolveCreateWithinRoot(cfg.RootPath, currPath, filename)
	if !ok {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handle.ErrorHandle(c, "文件保存失败！")
		return
	}
	err = c.SaveUploadedFile(f, target)
	if err != nil {
		logRequestEvent(c, accesslog.EventUpload, accesslog.ResultFailure, http.StatusInternalServerError)
		handle.ErrorHandle(c, "文件保存失败！")
		return
	}
	logRequestEvent(c, accesslog.EventUpload, accesslog.ResultSuccess, http.StatusMovedPermanently)
	c.Redirect(http.StatusMovedPermanently, path.Join(constant.FileGroupPrefix, currPath))
}

func safeUploadFilename(filename string) (string, bool) {
	base := path.Base(strings.ReplaceAll(filename, "\\", "/"))
	base = filepath.Base(base)
	if base == "" || base == "." || base == ".." {
		return "", false
	}
	return base, true
}

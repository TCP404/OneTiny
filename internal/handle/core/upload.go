package core

import (
	"net/http"
	"path"
	"path/filepath"

	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/constant"
	"github.com/TCP404/OneTiny-cli/internal/handle"
	"github.com/gin-gonic/gin"
)

func Uploader(c *gin.Context) {
	f, err := c.FormFile("upload_file")
	if err != nil {
		handle.ErrorHandle(c, "文件上传失败！")
		return
	}

	currPath := c.PostForm("path")
	err = c.SaveUploadedFile(f, filepath.Join(conf.Config.RootPath, currPath, f.Filename))
	if err != nil {
		handle.ErrorHandle(c, "文件保存失败！")
		return
	}
	c.Redirect(http.StatusMovedPermanently, path.Join(constant.FileGroupPrefix, currPath))
}

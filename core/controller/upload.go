package controller

import (
	"net/http"
	"oneTiny/common/config"
	"path"

	"github.com/gin-gonic/gin"
)

func Uploader(c *gin.Context) {
	f, err := c.FormFile("upload_file")
	if err != nil {
		errorHandle(c, "文件上传失败！")
		return
	}
	currPath := c.GetString("filename")
	err = c.SaveUploadedFile(f, path.Join(config.RootPath, currPath, f.Filename))
	if err != nil {
		errorHandle(c, "文件保存失败！")
		return
	}
	c.Redirect(http.StatusMovedPermanently, currPath)
}

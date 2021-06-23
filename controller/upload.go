package controller

import (
	"net/http"
	"oneTiny/config"
	"path"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	f, err := c.FormFile("upload_file")
	if err != nil {
		errorHandle(c, "文件上传失败！")
		return
	}
	err = c.SaveUploadedFile(f, path.Join(config.RootPath, config.CurrPath, f.Filename))
	if err != nil {
		errorHandle(c, "文件保存失败！")
		return
	}
	c.Redirect(http.StatusMovedPermanently, config.CurrPath)
}

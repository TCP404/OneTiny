package controller

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	f, err := c.FormFile("upload_file")
	if err != nil {
		errorHandle(c, "文件上传失败！")
		return
	}
	err = c.SaveUploadedFile(f, path.Join(RootPath, currPath, f.Filename))
	if err != nil {
		errorHandle(c, "文件保存失败！")
		return
	}
	c.Redirect(http.StatusMovedPermanently, currPath)
}

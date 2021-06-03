package controller

import (
	"net/http"
	"os"
	"path"
	"strings"
	"tinyServer/util"

	"github.com/gin-gonic/gin"
)

var (
	RootPath string
	Port     string
)

const (
	PORT       = "9090"
	ROOT       = "/"
	SEPARATORS = "/"
)

func NotFound(c *gin.Context) {
	c.String(http.StatusNotFound, "404 Page Not Found", nil)
}

func Handler(c *gin.Context) {
	filePath := c.Param("filename")
	if IsDir(filePath) {
		ShowFloder(c, filePath) // 如果是目录，就展示
	} else {
		Download(c, filePath) // 如果是文件，就下载
	}
}

// 判断是否是目录
func IsDir(filePath string) bool {
	if filePath == ROOT {
		return true
	}
	finfo, _ := os.Stat(path.Join(RootPath, filePath))
	return finfo.IsDir()
}

// 读取目录下所有文件
func ReadDir(absPath string) []string {
	files, _ := os.ReadDir(absPath)
	relPaths := make([]string, len(files))
	for i, f := range files {
		relPaths[i] = path.Join(strings.TrimPrefix(absPath, RootPath), f.Name())
	}
	return relPaths
}

// 展示目录下所有文件
func ShowFloder(c *gin.Context, floder string) {
	relPaths := ReadDir(path.Join(RootPath, floder))
	html := util.GenerateHTML(relPaths, floder)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(html))
}

// 下载文件
func Download(c *gin.Context, filePath string) {
	c.Writer.WriteHeader(http.StatusOK)
	// c.Header("Content-Type", "application/octet-stream")
	// c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filePath)
	c.File(path.Join(RootPath, filePath))
}

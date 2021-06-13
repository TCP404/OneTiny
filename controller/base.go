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
	currPath string
)

const (
	PORT       = "9090"
	ROOT       = "/"
	SEPARATORS = "/"
)

func Handler(c *gin.Context) {
	filePath := c.Param("filename")
	currPath = filePath
	if strings.HasSuffix(filePath, ".ico") {
		return
	}
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

// 展示目录下所有文件
func ShowFloder(c *gin.Context, floder string) {
	relPaths := ReadDir(c, path.Join(RootPath, floder))
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

// 读取目录下所有文件
func ReadDir(c *gin.Context, absPath string) []string {
	files, err := os.ReadDir(absPath)
	if err != nil {
		errorHandle(c, "目录读取失败！")
		return nil
	}
	relPaths := make([]string, len(files))
	for i, f := range files {
		relPaths[i] = path.Join(strings.TrimPrefix(absPath, RootPath), f.Name())
	}
	return relPaths
}

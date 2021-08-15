package middleware

import (
	"net/http"
	"oneTiny/config"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckLevel 负责检查当前访问层级是否超出设定最大层级
// 例如：
// 		共享目录为 /a/b/ , 最大层级为 2
//		✓: /a/b/			根目录
//		✓: /a/b/file	    根目录下文件
// 		✓: /a/b/c/			根目录下第一层目录
// 		✓: /a/b/c/file		根目录下第一层目录下的文件
//		✓: /a/b/c/d/		根目录下第二层目录
// 		✓: /a/b/c/d/file	根目录下第二层目录下的文件
//		x: /a/b/c/d/e/		根目录下第三层目录
// 		x: /a/b/c/d/e/file	根目录下第三层目录下的文件
func CheckLevel(c *gin.Context) {
	filePath := strings.TrimPrefix(c.Param("filename"), "/file")

	c.Set("filename", filePath)

	isD := isDir(filePath)
	c.Set("isDirectory", isD)
	isFile := !isD
	if isOverLevel(filePath, isFile) {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
	}
}

// 判断是否是目录
func isDir(filePath string) bool {
	if filePath == config.ROOT {
		return true
	}
	finfo, _ := os.Stat(path.Join(config.RootPath, filePath))
	return finfo.IsDir()
}

// 检查当前访问的路径是否超过限定层级
func isOverLevel(relPath string, isFile bool) bool {
	rel, _ := filepath.Rel(config.RootPath, filepath.Join(config.RootPath, relPath))
	i := strings.Split(rel, config.SEPARATORS)
	level := len(i)
	if i[0] == "." {
		level = 0
	}
	if isFile {
		level--
	}
	return level > int(config.MaxLevel)
}

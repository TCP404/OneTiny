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
// 		共享目录为 /a/b , 最大层级为 2
//		✓: /a/b
//		✓: /a/b/file
// 		✓: /a/b/c
// 		✓: /a/b/c/file
//		✓: /a/b/c/d
// 		✓: /a/b/c/d/file
//		×: /a/b/c/d/e
// 		×: /a/b/c/d/e/file
func CheckLevel(c *gin.Context) {
	filePath := c.Param("filename")
	config.CurrPath = filePath

	isD := isDir(filePath)
	if isD {
		c.Set("isDirectory", true)
	} else {
		c.Set("isDirectory", false)
	}

	if isOverLevel(filePath, !isD) {
		c.Abort()
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
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

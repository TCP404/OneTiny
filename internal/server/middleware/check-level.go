package middleware

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcp404/OneTiny/internal/constant"
	"github.com/tcp404/OneTiny/internal/kit/safepath"
	"github.com/tcp404/OneTiny/internal/state"

	"github.com/gin-gonic/gin"
)

// CheckLevel 负责检查当前访问层级是否超出设定最大层级
// 例如：
//
//	共享目录为 /a/b/ , 最大层级为 2
//	✓: /a/b/			根目录
//	✓: /a/b/file	    根目录下文件
//	✓: /a/b/c/			根目录下第一层目录
//	✓: /a/b/c/file		根目录下第一层目录下的文件
//	✓: /a/b/c/d/		根目录下第二层目录
//	✓: /a/b/c/d/file	根目录下第二层目录下的文件
//	x: /a/b/c/d/e/		根目录下第三层目录
//	x: /a/b/c/d/e/file	根目录下第三层目录下的文件
func CheckLevel(c *gin.Context) {
	cfg := currentSnapshot()
	c.Set(state.ContextKey, cfg)

	rawFilePath := parseWildcardFilename(c.Param("filename"))
	target, ok := safepath.ResolveWithinRoot(cfg.RootPath, rawFilePath)
	if !ok {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
		return
	}
	filePath := cleanRelPath(rawFilePath)

	if pathExists(target) {
		if _, ok := safepath.ResolveExistingWithinRoot(cfg.RootPath, filePath); !ok {
			c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
			c.Abort()
			return
		}
	}

	c.Set("filename", filePath)

	isD := isDir(cfg.RootPath, filePath)
	c.Set("isDirectory", isD)
	isFile := !isD
	if isOverLevel(cfg.RootPath, cfg.MaxLevel, filePath, isFile, c.Query("action") == "dl") {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
		return
	}
}

func parseWildcardFilename(filename string) string {
	if filename == "" || filename == constant.ROOT {
		return constant.ROOT
	}
	return strings.TrimPrefix(filename, "/")
}

func cleanRelPath(filePath string) string {
	if filePath == constant.ROOT {
		return constant.ROOT
	}
	cleaned := filepath.Clean(filePath)
	if cleaned == "." {
		return constant.ROOT
	}
	return cleaned
}

// 判断是否是目录
func isDir(rootPath, filePath string) bool {
	if filePath == constant.ROOT {
		return true
	}
	target, ok := safepath.ResolveExistingWithinRoot(rootPath, filePath)
	if !ok {
		return false
	}
	fInfo, err := os.Stat(target)
	if err != nil || fInfo == nil {
		return false
	}
	return fInfo.IsDir()
}

// 检查当前访问的路径是否超过限定层级
func isOverLevel(rootPath string, maxLevel uint8, relPath string, isFile bool, isDl bool) bool {
	cleanRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return true
	}
	target, ok := safepath.ResolveWithinRoot(rootPath, relPath)
	if !ok {
		return true
	}
	if pathExists(target) {
		target, ok = safepath.ResolveExistingWithinRoot(rootPath, relPath)
		if !ok {
			return true
		}
	}
	rel, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return true
	}
	i := strings.Split(rel, string(filepath.Separator))
	level := len(i)
	if i[0] == "." {
		level = 0
	}
	if isFile || isDl {
		level--
	}
	return level > int(maxLevel)
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

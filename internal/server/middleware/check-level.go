package middleware

import (
	"net/http"

	"github.com/tcp404/OneTiny/internal/kit/safepath"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/share"

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
	cfg := currentSnapshot(c)
	c.Set(runtime.ContextKey, cfg)

	rawFilePath := share.ParseWildcardFilename(c.Param("filename"))
	target, ok := safepath.ResolveWithinRoot(cfg.RootPath, rawFilePath)
	if !ok {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
		return
	}
	filePath := share.CleanRelPath(rawFilePath)

	if share.PathExists(target) {
		if _, ok := safepath.ResolveExistingWithinRoot(cfg.RootPath, filePath); !ok {
			c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
			c.Abort()
			return
		}
	}

	c.Set("filename", filePath)

	isD := share.IsDir(cfg.RootPath, filePath)
	c.Set("isDirectory", isD)
	isFile := !isD
	if share.IsOverLevel(cfg.RootPath, cfg.MaxLevel, filePath, isFile, c.Query("action") == "dl") {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
		return
	}
}

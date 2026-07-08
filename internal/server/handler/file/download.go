package file

import (
	"bufio"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/eutil"

	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/kit/progress"
	"github.com/tcp404/OneTiny/internal/kit/safepath"
	"github.com/tcp404/OneTiny/internal/server/handler"
	"github.com/tcp404/OneTiny/internal/server/routepath"
	"github.com/tcp404/OneTiny/internal/share"
)

type agent struct {
	abs         string
	rel         string
	action      string
	isDir       bool
	rootPath    string
	allowUpload bool
	output      io.Writer
}

func Downloader(c *gin.Context) {
	cfg := currentSnapshot(c)
	road := c.GetString("filename")
	abs, ok := safepath.ResolveWithinRoot(cfg.RootPath, road)
	if !ok {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
		return
	}
	if _, err := os.Lstat(abs); err == nil {
		abs, ok = safepath.ResolveExistingWithinRoot(cfg.RootPath, road)
		if !ok {
			c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
			c.Abort()
			return
		}
	} else if !os.IsNotExist(err) {
		c.String(http.StatusNotFound, "访问超出允许范围，请返回！")
		c.Abort()
		return
	}
	a := &agent{
		abs:         abs,
		rel:         road,
		action:      c.Query("action"),
		isDir:       c.GetBool("isDirectory"),
		rootPath:    cfg.RootPath,
		allowUpload: cfg.IsAllowUpload,
		output:      cfg.Output,
	}

	if a.rel == routepath.Root {
		a.readDir(c)
		return
	}
	if a.isDir {
		a.dir(c)
		return
	}
	a.file(c)
}

func (a *agent) file(c *gin.Context) {
	src, err := os.Open(a.abs)
	if err != nil {
		handler.ErrorHandle(c, err.Error())
		return
	}
	defer func(src *os.File) { _ = src.Close() }(src)

	log.Println("preparing file...")
	contentLen := share.ContentLen(a.abs)

	c.Status(http.StatusOK)
	if a.action == "dl" {
		c.Header("Content-Disposition", "attachment; filename="+filepath.Base(a.rel))
		c.Header("Content-Length", strconv.Itoa(int(contentLen)))
	}

	buf := make([]byte, share.BufferLimit)
	bar := progress.GetBar(a.rel, contentLen, a.outputOrDefault())

	// 小于 buf 时直接读取到 buf 中然后返回
	if contentLen < int64(len(buf)) {
		if _, err := io.CopyBuffer(io.MultiWriter(c.Writer, bar), src, buf); err != nil {
			logDownloadFailure(c)
			writeDownloadError(c)
			return
		}
		logRequestEvent(c, accesslog.EventDownload, accesslog.ResultSuccess, http.StatusOK)
		return
	}
	// 超过 buf 的大小时分片传输
	if err := a.flush(c, src, buf); err != nil {
		logDownloadFailure(c)
		writeDownloadError(c)
		return
	}
	logRequestEvent(c, accesslog.EventDownload, accesslog.ResultSuccess, http.StatusOK)
}

func (a *agent) dir(c *gin.Context) {
	if a.action != "dl" {
		a.readDir(c)
		return
	}

	log.Println("preparing archive directory...")
	contentLen := share.ContentLen(a.abs)
	c.Status(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(a.rel)+".zip")
	c.Header("Content-Length", strconv.Itoa(int(contentLen)))
	c.Header("Content-Type", mime.TypeByExtension("zip"))

	// 创建准备写入的压缩文件
	srcZip, err := os.CreateTemp(os.TempDir(), "temp.*.zip")
	if err != nil {
		handler.ErrorHandle(c, "压缩目录失败")
		return
	}
	defer func(srcZip *os.File) {
		_ = srcZip.Close()
		_ = os.Remove(srcZip.Name())
	}(srcZip)

	err = eutil.Zip(srcZip, a.abs)
	if err != nil {
		handler.ErrorHandle(c, "压缩目录失败, "+err.Error())
		return
	}

	buf := make([]byte, share.BufferLimit)
	if contentLen < int64(len(buf)) {
		c.File(srcZip.Name())
		logRequestEvent(c, accesslog.EventDownload, accesslog.ResultSuccess, http.StatusOK)
		return
	}
	if _, err := srcZip.Seek(0, io.SeekStart); err != nil {
		logDownloadFailure(c)
		writeDownloadError(c)
		return
	}
	if err := a.flush(c, srcZip, buf); err != nil {
		logDownloadFailure(c)
		writeDownloadError(c)
		return
	}
	logRequestEvent(c, accesslog.EventDownload, accesslog.ResultSuccess, http.StatusOK)
}

func (a *agent) readDir(c *gin.Context) {
	files, err := share.List(a.rootPath, a.abs)
	if err != nil {
		handler.ErrorHandle(c, "目录读取失败！")
		return
	}
	c.HTML(http.StatusOK, "list.tpl", gin.H{
		"pathTitle":   a.rel,
		"breadcrumbs": share.BuildBreadcrumbs(a.rel),
		"upload":      a.allowUpload,
		"files":       files,
	})
}

func (a *agent) flush(c *gin.Context, src io.Reader, buf []byte) error {
	data := bufio.NewReader(src)
	bar := progress.GetBar(a.rel, share.ContentLen(a.abs), a.outputOrDefault())

	for {
		n, err := data.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if _, writeErr := bar.Write(chunk); writeErr != nil {
				return writeErr
			}
			if _, writeErr := c.Writer.Write(chunk); writeErr != nil {
				return writeErr
			}
			c.Writer.(http.Flusher).Flush()
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *agent) outputOrDefault() io.Writer {
	if a.output != nil {
		return a.output
	}
	return os.Stderr
}

func logRequestEvent(c *gin.Context, event, result string, status int) {
	logger := accessLogger(c)
	if logger == nil {
		return
	}
	_ = logger.Write(accesslog.Event{
		ClientIP: clientIP(c),
		Method:   method(c),
		Event:    event,
		Path:     requestPath(c),
		Status:   status,
		Result:   result,
	})
}

func accessLogger(c *gin.Context) *accesslog.Logger {
	if c == nil {
		return nil
	}
	value, ok := c.Get(accesslog.ContextKey)
	if !ok {
		return nil
	}
	logger, ok := value.(*accesslog.Logger)
	if !ok {
		return nil
	}
	return logger
}

func logDownloadFailure(c *gin.Context) {
	logRequestEvent(c, accesslog.EventDownload, accesslog.ResultFailure, http.StatusInternalServerError)
}

func writeDownloadError(c *gin.Context) {
	if c == nil {
		return
	}
	if c.Writer != nil && c.Writer.Written() {
		return
	}
	c.Status(http.StatusInternalServerError)
}

func clientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.ClientIP()
}

func method(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return c.Request.Method
}

func requestPath(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Path
}

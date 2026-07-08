package core

import (
	"bufio"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/eutil"

	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/constant"
	"github.com/tcp404/OneTiny/internal/handle"
	"github.com/tcp404/OneTiny/internal/kit/progress"
	"github.com/tcp404/OneTiny/internal/kit/safepath"
)

type fileStructure struct {
	Size          string
	IsDir         bool
	DeviceAbsPath string
	URLRelPath    string
	Name          string
}

type pathCrumb struct {
	Name    string
	URL     string
	Current bool
}

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

	if a.rel == constant.ROOT {
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
		handle.ErrorHandle(c, err.Error())
		return
	}
	defer func(src *os.File) { _ = src.Close() }(src)

	log.Println("preparing file...")
	contentLen := getContentLen(a.abs)

	c.Status(http.StatusOK)
	if a.action == "dl" {
		c.Header("Content-Disposition", "attachment; filename="+filepath.Base(a.rel))
		c.Header("Content-Length", strconv.Itoa(int(contentLen)))
	}

	buf := make([]byte, constant.BufferLimit)
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
	contentLen := getContentLen(a.abs)
	c.Status(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(a.rel)+".zip")
	c.Header("Content-Length", strconv.Itoa(int(contentLen)))
	c.Header("Content-Type", mime.TypeByExtension("zip"))

	// 创建准备写入的压缩文件
	srcZip, err := os.CreateTemp(os.TempDir(), "temp.*.zip")
	if err != nil {
		handle.ErrorHandle(c, "压缩目录失败")
		return
	}
	defer func(srcZip *os.File) {
		_ = srcZip.Close()
		_ = os.Remove(srcZip.Name())
	}(srcZip)

	err = eutil.Zip(srcZip, a.abs)
	if err != nil {
		handle.ErrorHandle(c, "压缩目录失败, "+err.Error())
		return
	}

	buf := make([]byte, constant.BufferLimit)
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
	files := getFileInfos(c, a.rootPath, a.abs)
	c.HTML(http.StatusOK, "list.tpl", gin.H{
		"pathTitle":   a.rel,
		"breadcrumbs": buildBreadcrumbs(a.rel),
		"upload":      a.allowUpload,
		"files":       files,
	})
}

func getFileInfos(c *gin.Context, rootPath, absPath string) []fileStructure {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		handle.ErrorHandle(c, "目录读取失败！")
		return nil
	}

	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		relPath = strings.TrimPrefix(absPath, rootPath)
	}
	relPath = filepath.ToSlash(relPath)
	fileInfos := make([]fileStructure, len(dirEntries))

	for i, f := range dirEntries {
		info, _ := f.Info()
		size := info.Size()
		deviceAbsPath := filepath.Join(absPath, f.Name())
		urlRelPath := path.Join(constant.FileGroupPrefix, relPath, f.Name())
		isDir := f.Type().IsDir()

		if isDir { // 将目录的 size 设置为 0，文件则照常
			size = 0
			deviceAbsPath += string(filepath.Separator) // 如果是目录，不在路径末尾加上分隔符的话，返回上一级会有问题
			urlRelPath += string(filepath.Separator)
		}
		fileInfos[i] = fileStructure{
			DeviceAbsPath: deviceAbsPath,
			URLRelPath:    urlRelPath,
			Name:          f.Name(),
			Size:          eutil.SizeFmt(size),
			IsDir:         isDir,
		}
	}
	sort.SliceStable(fileInfos, func(i, j int) bool {
		if fileInfos[i].IsDir != fileInfos[j].IsDir {
			return fileInfos[i].IsDir
		}
		return strings.ToLower(fileInfos[i].Name) < strings.ToLower(fileInfos[j].Name)
	})
	return fileInfos
}

func buildBreadcrumbs(rel string) []pathCrumb {
	cleanRel := strings.Trim(filepath.ToSlash(rel), "/")
	if cleanRel == "" || rel == constant.ROOT {
		return []pathCrumb{{
			Name:    "根目录",
			URL:     constant.FileGroupPrefix + "/?action=view",
			Current: true,
		}}
	}

	parts := strings.Split(cleanRel, "/")
	crumbs := make([]pathCrumb, 0, len(parts)+1)
	crumbs = append(crumbs, pathCrumb{
		Name: "根目录",
		URL:  constant.FileGroupPrefix + "/?action=view",
	})

	for i, part := range parts {
		joined := path.Join(parts[:i+1]...)
		crumbs = append(crumbs, pathCrumb{
			Name:    part,
			URL:     path.Join(constant.FileGroupPrefix, joined) + "/?action=view",
			Current: i == len(parts)-1,
		})
	}
	return crumbs
}

func (a *agent) flush(c *gin.Context, src io.Reader, buf []byte) error {
	data := bufio.NewReader(src)
	bar := progress.GetBar(a.rel, getContentLen(a.abs), a.outputOrDefault())

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

func getContentLen(absPath string) int64 {
	var contentLen int64 = -1
	info, err := os.Stat(absPath)
	if err == nil {
		contentLen = info.Size()
	}
	return contentLen
}

func logRequestEvent(c *gin.Context, event, result string, status int) {
	accesslog.Log(accesslog.Event{
		ClientIP: clientIP(c),
		Method:   method(c),
		Event:    event,
		Path:     requestPath(c),
		Status:   status,
		Result:   result,
	})
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

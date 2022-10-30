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
	"strconv"
	"strings"

	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/constant"
	"github.com/TCP404/OneTiny-cli/internal/handle"
	"github.com/TCP404/OneTiny-cli/pkg"
	"github.com/TCP404/eutil"
	"github.com/gin-gonic/gin"
)

type fileStructure struct {
	Size          string
	IsDir         bool
	DeviceAbsPath string
	URLRelPath    string
	Name          string
}
type agent struct {
	abs    string
	rel    string
	action string
	isDir  bool
}

func Downloader(c *gin.Context) {
	road := c.GetString("filename")
	a := &agent{
		abs:    filepath.Join(conf.Config.RootPath, road),
		rel:    road,
		action: c.Query("action"),
		isDir:  c.GetBool("isDirectory"),
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
	bar := pkg.GetBar(a.rel, contentLen, conf.Config.Output)

	// 小于 buf 时直接读取到 buf 中然后返回
	if contentLen < int64(len(buf)) {
		_, _ = io.CopyBuffer(io.MultiWriter(c.Writer, bar), src, buf)
		return
	}
	// 超过 buf 的大小时分片传输
	a.flush(c, src, buf)
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
	}
	defer func(srcZip *os.File) {
		_ = srcZip.Close()
		_ = os.Remove(srcZip.Name())
	}(srcZip)

	err = eutil.Zip(srcZip, filepath.Join(conf.Config.RootPath, a.rel))
	if err != nil {
		handle.ErrorHandle(c, "压缩目录失败, "+err.Error())
	}

	buf := make([]byte, constant.BufferLimit)
	if contentLen < int64(len(buf)) {
		c.File(srcZip.Name())
		return
	}
	a.flush(c, srcZip, buf)
}

func (a *agent) readDir(c *gin.Context) {
	files := getFileInfos(c, filepath.Join(conf.Config.RootPath, a.rel))
	c.HTML(http.StatusOK, "list.tpl", gin.H{
		"pathTitle": a.rel,
		"upload":    conf.Config.IsAllowUpload,
		"files":     files,
	})
}

func getFileInfos(c *gin.Context, absPath string) []fileStructure {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		handle.ErrorHandle(c, "目录读取失败！")
		return nil
	}

	relPath := strings.TrimPrefix(absPath, conf.Config.RootPath)
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
	return fileInfos
}

func (a *agent) flush(c *gin.Context, src io.Reader, buf []byte) {
	data := bufio.NewReader(src)
	bar := pkg.GetBar(a.rel, getContentLen(a.abs), conf.Config.Output)

	for {
		_, err := data.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			handle.ErrorHandle(c, "server error!")
			return
		}
		_, _ = bar.Write(buf)
		_, _ = c.Writer.Write(buf)
		c.Writer.(http.Flusher).Flush()
	}
}

func getContentLen(absPath string) int64 {
	var contentLen int64 = -1
	info, err := os.Stat(absPath)
	if err == nil {
		contentLen = info.Size()
	}
	return contentLen
}

package controller

import (
	"bufio"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TCP404/OneTiny-cli/common/define"
	"github.com/TCP404/OneTiny-cli/config"
	"github.com/TCP404/OneTiny-cli/internal/model"
	"github.com/TCP404/OneTiny-cli/internal/util"

	"github.com/TCP404/eutil"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Downloader(c *gin.Context) {
	var (
		filePath = c.GetString("filename")
		action   = c.Query("action")
		isDir    = c.GetBool("isDirectory")
	)

	if filePath == define.ROOT {
		readDir(c, filePath)
		return
	}

	switch action {
	case "view":
		if isDir {
			readDir(c, filePath)
		} else {
			readFile(c, filepath.Join(config.RootPath, filePath))
		}
		return
	case "dl":
		if isDir {
			downloadDir(c, filePath)
		} else {
			downloadFile(c, filePath)
		}
		return
	}
}

func readDir(c *gin.Context, rel string) {
	files := viewDir(c, filepath.Join(config.RootPath, rel))
	html := util.GenerateIndexHTML(files, rel, config.IsAllowUpload)
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte(html))
}

func viewDir(c *gin.Context, absPath string) []model.FileStructure {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		errorHandle(c, "目录读取失败！")
		return nil
	}

	relPath := strings.TrimPrefix(absPath, config.RootPath)
	files := make([]model.FileStructure, len(dirEntries))

	for i, f := range dirEntries {
		info, _ := f.Info()
		size := info.Size()

		deviceAbsPath := filepath.Join(absPath, f.Name())
		urlRelPath := path.Join("/file", relPath, f.Name())
		fType := f.Type()
		if fType.IsDir() { // 将目录的 size 设置为 0，文件则照常
			size = 0
			deviceAbsPath += string(filepath.Separator) // 如果是目录，不在路径末尾加上分隔符的话，返回上一级会有问题
			urlRelPath += string(filepath.Separator)
		}
		files[i] = model.FileStructure{
			DeviceAbsPath: deviceAbsPath,
			URLRelPath:    urlRelPath,
			Name:          f.Name(),
			Size:          eutil.SizeFmt(size),
			Mode:          fType,
		}
	}
	return files
}

func readFile(c *gin.Context, absPath string) {
	src, err := os.Open(absPath)
	if err != nil {
		errorHandle(c, "获取文件失败！")
		return
	}
	defer func(src *os.File) { _ = src.Close() }(src)
	viewFile(c, src, make([]byte, 5*define.MB), getContentLen(absPath))
}

func viewFile(c *gin.Context, src *os.File, buf []byte, contentLen int64) {
	// TODO 处理文件类型，检查，并做响应处理
	// 小于 buf 时直接读取到 buf 中然后返回
	if contentLen < int64(len(buf)) {
		b, err := io.ReadAll(src)
		if err != nil {
			errorHandle(c, "读取失败！")
			return
		}
		c.String(http.StatusOK, string(b))
		return
	}
	// 超过 buf 的大小时分片传输
	flush(c, src, buf)
}

func downloadFile(c *gin.Context, rel string) {
	absPath := filepath.Join(config.RootPath, rel)
	src, err := os.Open(absPath)
	if err != nil {
		errorHandle(c, "获取文件失败！")
		return
	}
	defer func(src *os.File) { _ = src.Close() }(src)
	dlFile(c, src, make([]byte, 5*define.MB), rel, getContentLen(absPath))
}

func dlFile(c *gin.Context, src *os.File, buf []byte, rel string, contentLen int64) {
	c.Status(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(rel))
	c.Header("Content-Length", strconv.Itoa(int(contentLen)))

	log.Println("preparing file...")

	bar := util.GetBar(rel, contentLen)
	if contentLen < int64(len(buf)) {
		_, _ = io.CopyBuffer(io.MultiWriter(c.Writer, bar), src, buf)
		return
	}

	flush(c, src, buf)
}

func downloadDir(c *gin.Context, rel string) {
	// 创建准备写入的压缩文件
	srcZip, err := os.CreateTemp(os.TempDir(), "temp.*.zip")
	if err != nil {
		errorHandle(c, "压缩目录失败")
	}
	defer func(srcZip *os.File) {
		_ = srcZip.Close()
		_ = os.Remove(srcZip.Name())
	}(srcZip)

	dlDir(c, srcZip, make([]byte, 5*define.MB), rel, getContentLen(filepath.Join(config.RootPath, rel)))
}

func dlDir(c *gin.Context, srcZip *os.File, buf []byte, rel string, contentLen int64) {
	c.Status(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(rel)+".zip")
	c.Header("Content-Length", strconv.Itoa(int(contentLen)))
	c.Header("Content-Type", mime.TypeByExtension("zip"))

	log.Println("preparing archive directory...")

	err := eutil.Zip(srcZip, filepath.Join(config.RootPath, rel))
	if err != nil {
		errorHandle(c, "压缩目录失败, "+err.Error())
	}

	if contentLen < int64(len(buf)) {
		c.File(srcZip.Name())
		return
	}
	flush(c, srcZip, buf)
}

func getContentLen(absPath string) int64 {
	var contentLen int64 = -1
	info, err := os.Stat(absPath)
	if err == nil {
		contentLen = info.Size()
	}
	return contentLen
}

func flush(c *gin.Context, src io.Reader, buf []byte) {
	data := bufio.NewReader(src)
	for {
		_, err := data.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			errorHandle(c, "server error!")
			return
		}
		_, _ = c.Writer.Write(buf)
		c.Writer.(http.Flusher).Flush()
	}
}

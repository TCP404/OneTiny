package controller

import (
	"io"
	"io/fs"
	"net/http"
	"oneTiny/common"
	"oneTiny/common/config"
	"oneTiny/common/define"
	"oneTiny/core/model"
	"oneTiny/core/util"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Downloader(c *gin.Context) {
	filePath := c.GetString("filename")
	if c.GetBool("isDirectory") {
		showDir(c, filePath) // 如果是目录，就展示
	} else {
		download(c, filePath) // 如果是文件，就下载
	}
}

// showDir 会在文件类型为目录时调用，展示目录下所有文件
//
// 参数：
//		c   *gin.Context: gin 上下文对象
//		rel string: 用户点击的路径
func showDir(c *gin.Context, rel string) {
	files := readDir(c, filepath.Join(config.RootPath, rel))
	html := util.GenerateIndexHTML(files, rel, config.IsAllowUpload)
	c.Status(http.StatusOK)
	c.Writer.Write([]byte(html))
}

// download 会在文件类型*不*为目录时调用，将文件内容传输给客户端，即下载文件功能
//
// 参数：
//		c   *gin.Context: gin 上下文对象
//		rel string: 用户点击的路径
func download(c *gin.Context, rel string) {
	c.Status(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(rel))

	fileAbsPath := filepath.Join(config.RootPath, rel)
	src, err := os.Open(fileAbsPath)
	if err != nil {
		errorHandle(c, "获取文件失败！")
	}
	defer src.Close()

	var contentLen int64 = -1
	info, err := os.Stat(fileAbsPath)
	if err == nil {
		contentLen = info.Size()
	}
	c.Header("Content-Length", strconv.Itoa(int(contentLen)))

	buf := make([]byte, 32*1024) // 32kb
	io.CopyBuffer(io.MultiWriter(c.Writer, util.GetBar(rel, contentLen)), src, buf)

	// c.File(filepath.Join(config.RootPath, rel))
}

// readDir 读取目录下所有文件，将每个文件的相关信息存储在 model 中并返回
//
// 参数:
//		c *gin.Context: gin 上下文对象
// 		absPath string: 目录在系统中的绝对路径
// 返回值:
//		[]model.FileStruction: 目录下所有文件的相关信息的集合
func readDir(c *gin.Context, absPath string) []model.FileStruction {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		errorHandle(c, "目录读取失败！")
		return nil
	}

	relPath := strings.TrimPrefix(absPath, config.RootPath)
	files := make([]model.FileStruction, len(dirEntries))

	for i, f := range dirEntries {
		info, _ := f.Info()
		size := info.Size()

		deviceAbsPath := filepath.Join(absPath, f.Name())
		urlRelPath := path.Join("/file", relPath, f.Name())
		fType := f.Type()
		if fType == fs.ModeDir { // 将目录的 size 设置为 0，文件则照常
			size = 0
			deviceAbsPath += define.SEPARATORS // 如果是目录，不在路径末尾加上分隔符的话，返回上一级会有问题
			urlRelPath += define.SEPARATORS
		}
		files[i] = model.FileStruction{
			DeviceAbsPath: deviceAbsPath,
			URLRelPath:    urlRelPath,
			Name:          f.Name(),
			Size:          common.SizeFmt(size),
			Mode:          fType,
		}
	}
	return files
}

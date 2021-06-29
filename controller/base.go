package controller

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"oneTiny/config"
	"oneTiny/model"
	"oneTiny/util"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	pb "github.com/schollz/progressbar/v3"
)

func Handler(c *gin.Context) {
	filePath := c.Param("filename")
	if c.GetBool("isDirectory") {
		showFloder(c, filePath) // 如果是目录，就展示
	} else {
		download(c, filePath) // 如果是文件，就下载
	}
}

// showFloder 会在文件类型为目录时调用，展示目录下所有文件
//
// 参数：
//		c   *gin.Context: gin 上下文对象
//		rel string: 用户点击的路径
func showFloder(c *gin.Context, rel string) {
	files := readDir(c, filepath.Join(config.RootPath, rel))
	html := util.GenerateHTML(files, rel, config.IsAllowUpload)
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

	var contentLen int64
	info, err := os.Stat(fileAbsPath)
	if err != nil {
		contentLen = -1
	} else {
		contentLen = info.Size()
	}

	// 使用下载进度条，当访问者点击下载时，共享者会有进度条提示
	ops := []pb.Option{
		pb.OptionSetDescription(color.GreenString("Downloading ") + color.BlueString(rel)),
		pb.OptionSetWriter(config.Output),
		pb.OptionShowBytes(true),
		pb.OptionSetWidth(10),
		pb.OptionThrottle(65 * time.Millisecond),
		pb.OptionShowCount(),
		pb.OptionOnCompletion(func() {
			fmt.Fprint(config.Output, "\n")
		}),
		pb.OptionSpinnerType(14),
		pb.OptionFullWidth(),
	}

	bar := pb.NewOptions64(contentLen, ops...)
	io.Copy(io.MultiWriter(c.Writer, bar), src)

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
		fType := f.Type()
		size := info.Size()
		abs := filepath.Join(absPath, f.Name())
		rel := filepath.Join(relPath, f.Name())
		if fType == fs.ModeDir { // 将目录的 size 设置为 0，文件则照常
			size = 0
			abs += config.SEPARATORS // 如果是目录，不在路径末尾加上分隔符的话，返回上一级会有问题
			rel += config.SEPARATORS
		}
		files[i] = model.FileStruction{
			Abs:  abs,
			Rel:  rel,
			Name: f.Name(),
			Size: size,
			Mode: fType,
		}
	}
	return files
}

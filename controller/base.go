package controller

import (
	"io/fs"
	"net/http"
	"oneTiny/model"
	"oneTiny/util"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	RootPath string // 共享目录的根路径
	Port     string // 指定的服务端口
	currPath string // 上传目录时的当前路径
)

const (
	PORT       = "9090"
	ROOT       = "/"
	SEPARATORS = "/"
)

func Handler(c *gin.Context) {
	filePath := c.Param("filename")
	currPath = filePath
	if strings.HasSuffix(filePath, ".ico") { // 拦截浏览器默认请求 favicon.ico 的行为
		return
	}
	if IsDir(filePath) {
		ShowFloder(c, filePath) // 如果是目录，就展示
	} else {
		Download(c, filePath) // 如果是文件，就下载
	}
}

// 判断是否是目录
func IsDir(filePath string) bool {
	if filePath == ROOT {
		return true
	}
	finfo, _ := os.Stat(path.Join(RootPath, filePath))
	return finfo.IsDir()
}

// 展示目录下所有文件
func ShowFloder(c *gin.Context, rel string) {
	files := ReadDir(c, path.Join(RootPath, rel))
	html := util.GenerateHTML(files, rel)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(html))
}

// 下载文件
func Download(c *gin.Context, rel string) {
	// c.Header("Content-Type", "application/octet-stream")
	// c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+rel)
	c.Writer.WriteHeader(http.StatusOK)
	c.File(path.Join(RootPath, rel))
}

// 读取目录下所有文件
func ReadDir(c *gin.Context, absPath string) []model.FileStruction {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		errorHandle(c, "目录读取失败！")
		return nil
	}
	relPath := strings.TrimPrefix(absPath, RootPath)
	files := make([]model.FileStruction, len(dirEntries))
	for i, f := range dirEntries {
		info, _ := f.Info()
		size := info.Size()
		if f.Type() == fs.ModeDir {	// 将目录的 size 设置为 0，文件则照常
			size = 0
		}
		files[i] = model.FileStruction{
			Abs:  path.Join(absPath, f.Name()),
			Rel:  path.Join(relPath, f.Name()),
			Name: f.Name(),
			Size: size,
			Mode: f.Type(),
		}
	}
	return files
}

package controller

import (
	"io/fs"
	"net/http"
	"oneTiny/config"
	"oneTiny/model"
	"oneTiny/util"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	filePath := c.Param("filename")
	if strings.HasSuffix(filePath, ".ico") { // 拦截浏览器默认请求 favicon.ico 的行为
		return
	}
	config.CurrPath = filePath
	if isDir(filePath) {
		showFloder(c, filePath) // 如果是目录，就展示
	} else {
		download(c, filePath) // 如果是文件，就下载
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

// showFloder 会在文件类型为目录时调用，展示目录下所有文件
//
// 参数：
//		c   *gin.Context: gin 上下文对象
//		rel string: 用户点击的路径
func showFloder(c *gin.Context, rel string) {
	files := readDir(c, path.Join(config.RootPath, rel))
	html := util.GenerateHTML(files, rel, config.IsAllowUpload)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(html))
}

// download 会在文件类型*不*为目录时调用，将文件内容传输给客户端，即下载文件功能
//
// 参数：
//		c   *gin.Context: gin 上下文对象
//		rel string: 用户点击的路径
func download(c *gin.Context, rel string) {
	// c.Header("Content-Type", "application/octet-stream")
	// c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(rel))
	c.Writer.WriteHeader(http.StatusOK)
	c.File(path.Join(config.RootPath, rel))
}

// readDir 读取目录下所有文件，将每个文件的相关信息存储在 model 中并返回
// 
// 参数:
//		c 		*gin.Context: gin 上下文对象
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
		if fType == fs.ModeDir { // 将目录的 size 设置为 0，文件则照常
			size = 0
		}
		files[i] = model.FileStruction{
			Abs:  path.Join(absPath, f.Name()),
			Rel:  path.Join(relPath, f.Name()),
			Name: f.Name(),
			Size: size,
			Mode: fType,
		}
	}
	return files
}

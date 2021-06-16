package config

import (
	"flag"

	"os"
	"strings"
)

var (
	RootPath      string // 共享目录的根路径
	Port          string // 指定的服务端口
	Goos          string
	IsAllowUpload bool
)

const (
	PORT       = "9090"
	ROOT       = "/"
	SEPARATORS = "/"
)

func init() {
	wd, _ := os.Getwd()
	flag.StringVar(&RootPath, "r", wd, "指定对外开放的目录路径")
	flag.StringVar(&Port, "p", PORT, "指定开放的端口")
	flag.BoolVar(&IsAllowUpload, "a", false, "指定是否允许访问者上传")
	flag.Parse()
	RootPath = strings.TrimSuffix(RootPath, SEPARATORS)
}

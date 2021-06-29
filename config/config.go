package config

import (
	"io"
	"oneTiny/util"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/integrii/flaggy"
)

const (
	VERSION    string = "v0.2.2"
	PORT       string = "9090" // 默认端口
	ROOT       string = "/"
	SEPARATORS string = "/"
)

var (
	wd, _ = os.Getwd()
	ip, _ = util.GetIP()

	RootPath      string = wd           // 共享目录的根路径
	Port          string = PORT         // 指定的服务端口
	Goos          string = runtime.GOOS // 程序所在的操作系统
	IP            string = ip           // 本机局域网IP
	MaxLevel      uint8  = 0            // 允许访问的最大层级
	IsAllowUpload bool   = false        // 是否允许上传
)

var (
	CurrPath string // 上传目录时的当前路径
	Output   io.Writer
)

func init() {
	flaggy.SetName("OneTiny")
	flaggy.SetVersion(VERSION)
	flaggy.SetDescription("一个用于局域网内共享文件的FTP程序")
	flaggy.SetHelpFlagDescription("打印帮助信息")
	flaggy.SetVersionFlagDescription("打印版本信息，当前版本: " + VERSION)

	flaggy.String(&RootPath, "r", "road", "指定对外开放的目录路径")
	flaggy.String(&Port, "p", "port", "指定开放的端口")
	flaggy.Bool(&IsAllowUpload, "a", "allow", "指定是否允许访问者上传")
	flaggy.UInt8(&MaxLevel, "x", "max", "指定允许访问的深度（默认仅限访问共享目录）")

	// updateCmd := flaggy.NewSubcommand("update")
	// flaggy.AttachSubcommand(updateCmd, 1)
	flaggy.Parse()
	RootPath = strings.TrimSuffix(RootPath, SEPARATORS)
	if Goos == "windows" {
		Output = color.Output
	} else {
		Output = os.Stderr
	}
}

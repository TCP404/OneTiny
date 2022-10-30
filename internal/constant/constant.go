package constant

import "github.com/TCP404/eutil"

var VERSION string

const (
	ROOT            string = "/"
	FileGroupPrefix string = "/file"
)

const (
	MaxLevel      uint8  = 0           // 允许访问的最大层级
	Port          int    = 8192        // 指定的服务端口
	IsAllowUpload bool   = false       // 是否允许上传
	IsSecure      bool   = false       // 是否开启访问登录
	RootPath      string = "/"         // 共享目录的根路径，默认值：当前目录
	Username      string = "admin"     // 访问登录的帐号
	Password      string = "admin"     // 访问登录的密码
	IP            string = "127.0.0.1" // 本机局域网IP
)

const (
	VersionListURL   = "https://api.github.com/repos/TCP404/OneTiny/tags"
	VersionLatestURL = "https://api.github.com/repos/TCP404/OneTiny/releases/latest"
	VersionByTagURL  = "https://api.github.com/repos/TCP404/OneTiny/releases/tags/"
)

var (
	ReleaseName = map[string]string{
		"linux":   "OneTiny",
		"windows": "OneTiny.exe",
		"darwin":  "OneTiny_mac",
	}
)

var BufferLimit = 512 * eutil.KB

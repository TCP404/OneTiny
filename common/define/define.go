package define

const (
	VERSION    string = "v0.3.0"
	ROOT       string = "/"
	SEPARATORS string = "/"
)

const (
	MaxLevel      uint8  = 0           // 允许访问的最大层级，默认值  0
	Port          int    = 9090        // 指定的服务端口，默认值 9090
	IsAllowUpload bool   = false       // 是否允许上传，默认值：否
	IsSecure      bool   = false       // 是否开启访问登录，默认值：否
	RootPath      string = "/"         // 共享目录的根路径，默认值：当前目录
	Username      string = "admin"     // 访问登录的帐号
	Password      string = "admin"     // 访问登录的密码
	IP            string = "127.0.0.1" // 本机局域网IP
)

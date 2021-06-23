package start

import (
	"log"
	"oneTiny/config"
	"strconv"

	"os"
)

func Verify() {
	iPort, _ := strconv.Atoi(config.Port)
	verifyPort(iPort)
	verifyPath(config.RootPath)
}

// verifyPort 负责校验用户指定的端口号是否在合法范围内。
// 对于 unix 系列的操作系统，端口允许范围在 [1024,65535]；
// 对于微软操作系统，端口允许范围在 [5001, 65535]；
// 对于其他操作系统暂时不做验证；
//
// 参数:
// 		iPort int: 用户指定的端口
func verifyPort(iPort int) {
	switch config.Goos {
	case "linux", "darwin":
		if !(iPort >= 1024 && iPort <= 65535) {
			log.Printf("不可以设置系统预留端口 %d，您可以设置的范围在 [ 1024 ~ 65535 ] 之间。", iPort)
			os.Exit(1)
		}
	case "windows":
		if !(iPort >= 5001 && iPort <= 65535) {
			log.Printf("不可以设置系统预留端口 %d，您可以设置的范围在 [ 5001 ~ 65535 ] 之间。", iPort)
			os.Exit(1)
		}
	}
}

// verifyPath 负责校验用户指定的共享目录的绝对路径
//
// 参数:
// 		rootPath string: 用户指定的共享目录绝对路径
func verifyPath(rootPath string) {
	if _, err := os.Stat(rootPath); err != nil {
		if !os.IsExist(err) {
			log.Printf("无法设置您指定的共享路径，请检查给出的路径是否有问题：%s", rootPath)
			os.Exit(1)
		}
	}
}

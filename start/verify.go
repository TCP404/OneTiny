package start

import (
	"log"
	"oneTiny/config"
	"runtime"
	"strconv"

	"os"
)

func Verify() {
	config.Goos = runtime.GOOS
	iPort, _ := strconv.Atoi(config.Port)
	verifyPort(iPort)
	verifyPath(config.RootPath)
}

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

func verifyPath(rootPath string) {
	if _, err := os.Stat(rootPath); err != nil {
		if !os.IsExist(err) {
			log.Printf("无法设置您指定的共享路径，请检查给出的路径是否有问题：%s", rootPath)
			os.Exit(1)
		}
	}
}

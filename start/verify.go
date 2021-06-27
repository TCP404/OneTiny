package start

import (
	"errors"
	"log"
	"oneTiny/config"
	"strconv"

	"os"

	"github.com/fatih/color"
)

func Verify() {
	iPort, _ := strconv.Atoi(config.Port)
	if err := verifyPort(iPort); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := verifyPath(config.RootPath); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

// verifyPort 负责校验用户指定的端口号是否在合法范围内。
// 对于 unix 系列的操作系统，端口允许范围在 [1024,65535]；
// 对于微软操作系统，端口允许范围在 [5001, 65535]；
// 对于其他操作系统暂时不做验证；
//
// 参数:
// 		iPort int: 用户指定的端口
// 返回值:
// 		error: 错误信息
func verifyPort(iPort int) error {
	switch config.Goos {
	case "linux", "darwin":
		if !(iPort >= 1024 && iPort <= 65535) {
			return errors.New(color.RedString("不可以设置系统预留端口 %d，您可以设置的范围在 [ 1024 ~ 65535 ] 之间。", iPort))
		}
	case "windows":
		if !(iPort >= 5001 && iPort <= 65535) {
			return errors.New(color.RedString("不可以设置系统预留端口 %d，您可以设置的范围在 [ 5001 ~ 65535 ] 之间。", iPort))
		}
	}
	return nil
}

// verifyPath 负责校验用户指定的共享目录的绝对路径
//
// 参数:
// 		rootPath string: 用户指定的共享目录绝对路径
// 返回值:
// 		error: 错误信息
func verifyPath(rootPath string) error {
	if _, err := os.Stat(rootPath); err != nil {
		if !os.IsExist(err) {
			return errors.New(color.RedString("无法设置您指定的共享路径，请检查给出的路径是否有问题：%s", rootPath))
		}
	}
	return nil
}

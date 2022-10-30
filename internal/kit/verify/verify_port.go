package verify

import (
	"errors"
	"runtime"

	"github.com/TCP404/OneTiny-cli/pkg/container"
	"github.com/fatih/color"
)

type portVerifier struct {
	port int
}

var _ container.Handler = (*portVerifier)(nil)

func NewPortVerifier(port int) container.Handler { return &portVerifier{port: port} }

// portHandler.Handle 负责校验用户指定的端口号是否在合法范围内。
// 对于 unix 系列的操作系统，端口允许范围在 [1024,65535]；
// 对于微软操作系统，端口允许范围在 [5001, 65535]；
// 对于其他操作系统暂时不做验证；
func (p *portVerifier) Handle() error {
	switch runtime.GOOS {
	case "linux", "darwin":
		if !(p.port >= 1024 && p.port <= 65535) {
			return errors.New(color.RedString("不可以设置系统预留端口 %d, 您可以设置的范围在 [ 1024 ~ 65535 ] 之间。", p.port))
		}
	case "windows":
		if !(p.port >= 5001 && p.port <= 65535) {
			return errors.New(color.RedString("不可以设置系统预留端口 %d, 您可以设置的范围在 [ 5001 ~ 65535 ] 之间。", p.port))
		}
	}
	return nil
}

package verify

import (
	"errors"
	"os"

	"github.com/TCP404/OneTiny-cli/pkg/container"
	"github.com/fatih/color"
)

type pathVerifier struct {
	rootPath string
}

var _ container.Handler = (*pathVerifier)(nil)

func NewPathVerifier(path string) container.Handler { return &pathVerifier{rootPath: path} }

// pathHandler.Handle 负责校验用户指定的共享目录的绝对路径
func (p *pathVerifier) Handle() error {
	if _, err := os.Stat(p.rootPath); err != nil {
		if !os.IsExist(err) {
			return errors.New(color.RedString("无法设置您指定的共享路径, 请检查给出的路径是否有问题：%s", p.rootPath))
		}
	}
	return nil
}

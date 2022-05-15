package util

import (
	"fmt"
	"oneTiny/common/config"
	"time"

	pb "github.com/schollz/progressbar/v3"
)

func GetBar(filename string, contentLen int64) *pb.ProgressBar {
	// 使用下载进度条，当访问者点击下载时，共享者会有进度条提示
	ops := []pb.Option{
		pb.OptionSetDescription("[green]Downloading[reset] [blue]" + filename + "[reset]"),
		pb.OptionSetWidth(10),
		pb.OptionThrottle(65 * time.Millisecond),
		pb.OptionShowCount(),
		pb.OptionOnCompletion(func() { fmt.Fprint(config.Output, "\n") }),
		pb.OptionSpinnerType(14),
		pb.OptionFullWidth(),
		pb.OptionSetWriter(config.Output),
		pb.OptionEnableColorCodes(true),
		pb.OptionShowBytes(true),
		pb.OptionSetWidth(50),
	}
	return pb.NewOptions64(contentLen, ops...)
}

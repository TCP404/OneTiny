package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/runtime"
)

func Logger(rt *runtime.Runtime) gin.HandlerFunc {
	var output io.Writer = os.Stderr
	if rt != nil {
		if snapshotOutput := rt.Snapshot().Output; snapshotOutput != nil {
			output = snapshotOutput
		}
	}
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Output: output,
		Formatter: func(param gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}
			if param.Latency > time.Minute {
				// Truncate in a golang < 1.8 safe way
				param.Latency = param.Latency - param.Latency%time.Second
			}
			return fmt.Sprintf("%v %s %3d %s %13v | %15s |%s %-7s %s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage,
			)
		}})
}

package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/TCP404/OneTiny-cli/internal/conf"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Output: conf.Config.Output,
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

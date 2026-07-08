package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
)

func AccessLogger(logger *accesslog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if logger != nil {
			c.Set(accesslog.ContextKey, logger)
		}
		c.Next()
	}
}

func AccessLog(logger *accesslog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				c.Status(http.StatusInternalServerError)
				logRequestEvent(logger, c, accesslog.EventError, accesslog.ResultFailure, http.StatusInternalServerError)
				panic(recovered)
			}
		}()

		c.Next()

		status := c.Writer.Status()
		event, result := classifyRequestStatus(status)
		logRequestEvent(logger, c, event, result, status)
	}
}

func classifyRequestStatus(status int) (string, string) {
	switch {
	case status >= http.StatusInternalServerError:
		return accesslog.EventError, accesslog.ResultFailure
	case status >= http.StatusBadRequest:
		return accesslog.EventReject, accesslog.ResultReject
	default:
		return accesslog.EventAccess, accesslog.ResultSuccess
	}
}

func logRequestEvent(logger *accesslog.Logger, c *gin.Context, event, result string, status int) {
	if logger == nil {
		return
	}
	_ = logger.Write(accesslog.Event{
		ClientIP: clientIP(c),
		Method:   method(c),
		Event:    event,
		Path:     requestPath(c),
		Status:   status,
		Result:   result,
	})
}

func clientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.ClientIP()
}

func method(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return c.Request.Method
}

func requestPath(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Path
}

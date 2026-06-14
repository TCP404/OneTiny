package middleware

import (
	"net/http"

	"github.com/TCP404/OneTiny-cli/internal/accesslog"
	"github.com/gin-gonic/gin"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				c.Status(http.StatusInternalServerError)
				logRequestEvent(c, accesslog.EventError, accesslog.ResultFailure, http.StatusInternalServerError)
				panic(recovered)
			}
		}()

		c.Next()

		status := c.Writer.Status()
		event, result := classifyRequestStatus(status)
		logRequestEvent(c, event, result, status)
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

func logRequestEvent(c *gin.Context, event, result string, status int) {
	accesslog.Log(accesslog.Event{
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

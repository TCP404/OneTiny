package secure

import (
	"net/http"

	"github.com/TCP404/OneTiny-cli/internal/accesslog"
	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/runtimeconf"
	"github.com/TCP404/OneTiny-cli/internal/security"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LoginGet 接收 GET /login 请求,返回登录页面
func LoginGet(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tpl", nil)
}

// LoginPost 接收 POST /login 请求,验证帐号密码,验证通过则生成 session
func LoginPost(c *gin.Context) {
	// 检查帐号密码
	// 通过则生成session，跳转首页
	// 不通过则返回登录页
	cfg := loginSnapshot()
	if c.PostForm("username") == cfg.Username &&
		security.VerifyPassword(cfg.PasswordHash, c.PostForm("password")) == nil {
		session := sessions.Default(c)
		session.Set("login", cfg.SessionVal)
		session.Save()
		logLoginEvent(c, accesslog.ResultSuccess)
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "登录成功"})
		return
	}
	logLoginEvent(c, accesslog.ResultFailure)
	logRejectEvent(c)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "登录失败"})
}

type loginCredentialSnapshot struct {
	Username     string
	PasswordHash string
	SessionVal   string
}

func loginSnapshot() loginCredentialSnapshot {
	result := loginCredentialSnapshot{
		Username:     conf.Config.Username,
		PasswordHash: conf.Config.Password,
		SessionVal:   conf.Config.SessionVal,
	}

	cfg := runtimeconf.Current()
	if cfg == nil {
		return result
	}

	snapshot := cfg.Snapshot()
	if snapshot.Username != "" {
		result.Username = snapshot.Username
	}
	if snapshot.PasswordHash != "" {
		result.PasswordHash = snapshot.PasswordHash
	}
	if snapshot.SessionVal != "" {
		result.SessionVal = snapshot.SessionVal
	}
	return result
}

func logLoginEvent(c *gin.Context, result string) {
	accesslog.Log(accesslog.Event{
		ClientIP: clientIP(c),
		Method:   method(c),
		Event:    accesslog.EventLogin,
		Path:     requestPath(c),
		Status:   http.StatusOK,
		Result:   result,
	})
}

func logRejectEvent(c *gin.Context) {
	accesslog.Log(accesslog.Event{
		ClientIP: clientIP(c),
		Method:   method(c),
		Event:    accesslog.EventReject,
		Path:     requestPath(c),
		Status:   http.StatusUnauthorized,
		Result:   accesslog.ResultReject,
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

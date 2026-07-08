package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/security"
)

func TestLoginPostWithValidCredentialsSetsSession(t *testing.T) {
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(runtime.ContextKey, runtime.Snapshot{
			Username:     "admin",
			PasswordHash: hash,
			SessionVal:   "session",
		})
		c.Next()
	})
	r.Use(sessions.Sessions("SESSIONID", cookie.NewStore([]byte("secret"))))
	r.POST("/login", LoginPost)

	form := url.Values{"username": {"admin"}, "password": {"secret"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() == "" || !strings.Contains(rec.Body.String(), "登录成功") {
		t.Fatalf("body = %q, want success", rec.Body.String())
	}
}

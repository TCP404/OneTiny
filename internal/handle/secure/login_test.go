package secure

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/security"
	"github.com/tcp404/OneTiny/internal/state"
)

func resetLoginTestConfig(t *testing.T) *accesslog.Logger {
	t.Helper()
	original := *conf.UnsafeCurrentForTest()
	originalRuntime := state.Current()
	state.SetCurrent(nil)
	logger := accesslog.New(filepath.Join(t.TempDir(), "access.log"))
	restoreLogger := accesslog.SetLoggerForTest(logger)
	t.Cleanup(func() {
		restoreLogger()
		state.SetCurrent(originalRuntime)
		*conf.UnsafeCurrentForTest() = original
	})
	return logger
}

func setLoginTestGinMode(t *testing.T) {
	t.Helper()
	original := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(original)
	})
}

func TestLoginPostVerifiesPlainUsernameAndBcryptPassword(t *testing.T) {
	resetLoginTestConfig(t)
	setLoginTestGinMode(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.UnsafeCurrentForTest().Username = "admin"
	conf.UnsafeCurrentForTest().PasswordHash = hash

	body, rec := postLogin(t, "admin", "secret")

	if got := body["code"]; got != float64(1) {
		t.Fatalf("code = %v, want 1; body=%v", got, body)
	}
	cookies := rec.Header().Values("Set-Cookie")
	if len(cookies) == 0 {
		t.Fatal("successful login did not set a session cookie")
	}
	if !strings.Contains(cookies[0], "onetiny-test") {
		t.Fatalf("Set-Cookie = %q, want session cookie name onetiny-test", cookies[0])
	}
}

func TestLoginPostRejectsWrongPassword(t *testing.T) {
	resetLoginTestConfig(t)
	setLoginTestGinMode(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.UnsafeCurrentForTest().Username = "admin"
	conf.UnsafeCurrentForTest().PasswordHash = hash

	body, _ := postLogin(t, "admin", "wrong")

	if got := body["code"]; got != float64(0) {
		t.Fatalf("code = %v, want 0; body=%v", got, body)
	}
}

func TestLoginPostWritesSuccessAndFailureEvents(t *testing.T) {
	logger := resetLoginTestConfig(t)
	setLoginTestGinMode(t)
	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.UnsafeCurrentForTest().Username = "admin"
	conf.UnsafeCurrentForTest().PasswordHash = hash

	postLogin(t, "admin", "secret")
	postLogin(t, "admin", "wrong")

	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventLogin})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Read returned %d login events, want 2: %+v", len(events), events)
	}
	if events[0].Result != accesslog.ResultSuccess || events[0].Status != http.StatusOK {
		t.Fatalf("first login event = %+v, want success", events[0])
	}
	if events[1].Result != accesslog.ResultFailure || events[1].Status != http.StatusOK {
		t.Fatalf("second login event = %+v, want failure", events[1])
	}

	rejects, err := logger.Read(accesslog.Filter{Event: accesslog.EventReject})
	if err != nil {
		t.Fatalf("Read reject returned error: %v", err)
	}
	if len(rejects) != 1 {
		t.Fatalf("Read returned %d reject events, want 1: %+v", len(rejects), rejects)
	}
	if rejects[0].Path != "/login" || rejects[0].Status != http.StatusUnauthorized || rejects[0].Result != accesslog.ResultReject {
		t.Fatalf("reject event = %+v, want failed login reject", rejects[0])
	}
}

func TestLoginPostUsesRuntimeCredentials(t *testing.T) {
	logger := resetLoginTestConfig(t)
	setLoginTestGinMode(t)
	hash, err := security.HashPassword("runtime-pass")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.UnsafeCurrentForTest().Username = "global-user"
	conf.UnsafeCurrentForTest().PasswordHash = "$2a$10$YJCMw3VjB9FlGm8zJbv8we8z0N1P6l4L7jXWaCOc3SNH0WcEjPzNe"
	state.SetCurrent(state.NewRuntimeConfig(state.ConfigSnapshot{
		Username:     "runtime-user",
		PasswordHash: hash,
		SessionVal:   "runtime-session",
	}))

	result, rec := postLogin(t, "runtime-user", "runtime-pass")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if result["code"] != float64(1) {
		t.Fatalf("login result = %+v, want success", result)
	}
	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventLogin})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 || events[0].Result != accesslog.ResultSuccess {
		t.Fatalf("login events = %+v, want one success", events)
	}
}

func TestLoginPostFallsBackToGlobalCredentialsWhenRuntimeCredentialsEmpty(t *testing.T) {
	resetLoginTestConfig(t)
	setLoginTestGinMode(t)
	hash, err := security.HashPassword("global-pass")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	conf.UnsafeCurrentForTest().Username = "global-user"
	conf.UnsafeCurrentForTest().PasswordHash = hash
	state.SetCurrent(state.NewRuntimeConfig(state.ConfigSnapshot{
		RootPath: t.TempDir(),
		Port:     9090,
		MaxLevel: 2,
	}))

	result, rec := postLogin(t, "global-user", "global-pass")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if result["code"] != float64(1) {
		t.Fatalf("login result = %+v, want success", result)
	}
	cookies := rec.Header().Values("Set-Cookie")
	if len(cookies) == 0 {
		t.Fatal("successful login did not set a session cookie")
	}
}

func postLogin(t *testing.T, username, password string) (map[string]any, *httptest.ResponseRecorder) {
	t.Helper()
	r := gin.New()
	r.Use(sessions.Sessions("onetiny-test", cookie.NewStore([]byte("secret"))))
	r.POST("/login", LoginPost)

	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response %q: %v", rec.Body.String(), err)
	}
	return body, rec
}

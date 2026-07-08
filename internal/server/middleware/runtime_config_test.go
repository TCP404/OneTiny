package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
)

func resetMiddlewareRuntimeTest(t *testing.T) {
	t.Helper()

	originalConfig := *conf.Config
	originalMode := gin.Mode()

	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		*conf.Config = originalConfig
		runtimeconf.SetCurrent(nil)
		gin.SetMode(originalMode)
	})
}

func TestCheckLoginUsesRuntimeSecureFlag(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	conf.Config.IsSecure = true
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		IsSecure: false,
	}))

	rec := serveCheckLoginRequest()

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestCheckLoginRedirectsWhenRuntimeSecureAndSessionMissing(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	conf.Config.IsSecure = false
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		IsSecure: true,
	}))

	rec := serveCheckLoginRequest()

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTemporaryRedirect)
	}
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Fatalf("Location = %q, want /login", got)
	}
}

func TestCheckLoginUsesRuntimeSessionVal(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	conf.Config.IsSecure = false
	conf.Config.SessionVal = "global-session"
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		IsSecure:   true,
		SessionVal: "runtime-session",
	}))

	rec := serveCheckLoginRequestWithSession(t, "runtime-session")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestCheckLoginRejectsGlobalSessionWhenRuntimeSessionValSet(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	conf.Config.IsSecure = false
	conf.Config.SessionVal = "global-session"
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		IsSecure:   true,
		SessionVal: "runtime-session",
	}))

	rec := serveCheckLoginRequestWithSession(t, "global-session")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTemporaryRedirect)
	}
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Fatalf("Location = %q, want /login", got)
	}
}

func TestCheckLevelUsesRuntimeRootForDirectoryDetection(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	globalRoot := t.TempDir()
	runtimeRoot := t.TempDir()
	conf.Config.RootPath = globalRoot
	conf.Config.MaxLevel = 5

	if err := os.WriteFile(filepath.Join(globalRoot, "docs"), []byte("file"), 0o600); err != nil {
		t.Fatalf("write global docs file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(runtimeRoot, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir runtime docs: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: runtimeRoot,
		MaxLevel: 5,
	}))

	rec := serveCheckLevelRequest("/file/docs/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != "dir=true" {
		t.Fatalf("body = %q, want dir=true", got)
	}
}

func TestCheckLevelMissingPathDoesNotPanic(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.MaxLevel = 0
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 0,
	}))

	rec := serveCheckLevelRequest("/file/missing.txt")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != "dir=false" {
		t.Fatalf("body = %q, want dir=false", got)
	}
}

func TestCheckLevelUsesRuntimeMaxLevel(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.MaxLevel = 5
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 0,
	}))

	rec := serveCheckLevelRequest("/file/nested/")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "访问超出允许范围") {
		t.Fatalf("body = %q, want over-level message", rec.Body.String())
	}
}

func TestCheckLevelRejectsEscapedPath(t *testing.T) {
	resetMiddlewareRuntimeTest(t)

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(parent, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 5,
	}))

	rec := serveCheckLevelParam("/../secret.txt")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestParseWildcardFilenameUsesURLPathSlash(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "nested file directory", filename: "/file/hello.txt", want: "file/hello.txt"},
		{name: "root slash", filename: "/", want: "/"},
		{name: "empty root", filename: "", want: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseWildcardFilename(tt.filename); got != tt.want {
				t.Fatalf("parseWildcardFilename(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func serveCheckLoginRequest() *httptest.ResponseRecorder {
	r := gin.New()
	r.Use(sessions.Sessions("onetiny-test", cookie.NewStore([]byte("secret"))))
	r.GET("/private", CheckLogin, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/private", nil))
	return rec
}

func serveCheckLoginRequestWithSession(t *testing.T, login string) *httptest.ResponseRecorder {
	t.Helper()

	r := gin.New()
	r.Use(sessions.Sessions("onetiny-test", cookie.NewStore([]byte("secret"))))
	r.GET("/set-session", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("login", login)
		if err := session.Save(); err != nil {
			t.Fatalf("session.Save() error = %v", err)
		}
		c.Status(http.StatusNoContent)
	})
	r.GET("/private", CheckLogin, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	setRec := httptest.NewRecorder()
	r.ServeHTTP(setRec, httptest.NewRequest(http.MethodGet, "/set-session", nil))
	if setRec.Code != http.StatusNoContent {
		t.Fatalf("set-session status = %d, want %d", setRec.Code, http.StatusNoContent)
	}

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	for _, cookie := range setRec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func serveCheckLevelRequest(target string) *httptest.ResponseRecorder {
	r := gin.New()
	r.GET("/file/*filename", CheckLevel, func(c *gin.Context) {
		c.String(http.StatusOK, "dir=%t", c.GetBool("isDirectory"))
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func serveCheckLevelParam(filename string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/file/"+filename, nil)
	c.Params = gin.Params{{Key: "filename", Value: filename}}

	CheckLevel(c)
	if !c.IsAborted() && rec.Code == http.StatusOK {
		c.String(http.StatusOK, "dir=%t", c.GetBool("isDirectory"))
	}
	return rec
}

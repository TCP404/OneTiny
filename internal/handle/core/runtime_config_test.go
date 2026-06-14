package core

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TCP404/OneTiny-cli/internal/accesslog"
	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/runtimeconf"
	"github.com/TCP404/OneTiny-cli/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func resetCoreRuntimeTest(t *testing.T) *accesslog.Logger {
	t.Helper()

	originalConfig := *conf.Config
	originalMode := gin.Mode()
	logger := accesslog.New(filepath.Join(t.TempDir(), "access.log"))
	restoreLogger := accesslog.SetLoggerForTest(logger)

	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		restoreLogger()
		*conf.Config = originalConfig
		runtimeconf.SetCurrent(nil)
		gin.SetMode(originalMode)
	})
	return logger
}

func TestUploaderRejectsWhenRuntimeDisallowsUpload(t *testing.T) {
	resetCoreRuntimeTest(t)

	root := t.TempDir()
	conf.Config.RootPath = root
	conf.Config.IsAllowUpload = true
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: false,
	}))

	rec := serveUploadRequest(t, "blocked.txt", "blocked", "/")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "当前未开启上传") {
		t.Fatalf("body = %q, want upload disabled message", rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "blocked.txt")); !os.IsNotExist(err) {
		t.Fatalf("blocked upload touched file, stat err=%v", err)
	}
}

func TestUploaderSavesToRuntimeRootPath(t *testing.T) {
	resetCoreRuntimeTest(t)

	globalRoot := t.TempDir()
	runtimeRoot := t.TempDir()
	conf.Config.RootPath = globalRoot
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      runtimeRoot,
		IsAllowUpload: true,
	}))

	rec := serveUploadRequest(t, "runtime.txt", "runtime root", "/")

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusMovedPermanently, rec.Body.String())
	}
	if got, err := os.ReadFile(filepath.Join(runtimeRoot, "runtime.txt")); err != nil || string(got) != "runtime root" {
		t.Fatalf("runtime root file = %q, err=%v; want content", string(got), err)
	}
	if _, err := os.Stat(filepath.Join(globalRoot, "runtime.txt")); !os.IsNotExist(err) {
		t.Fatalf("global root was used, stat err=%v", err)
	}
}

func TestUploaderRejectsEscapedUploadPath(t *testing.T) {
	resetCoreRuntimeTest(t)

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: true,
	}))

	rec := serveUploadRequest(t, "escape.txt", "escape", "../outside")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(outside, "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("escaped upload touched outside file, stat err=%v", err)
	}
}

func TestUploaderSanitizesUploadedFilename(t *testing.T) {
	resetCoreRuntimeTest(t)

	root := t.TempDir()
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: true,
	}))

	rec := serveUploadRequest(t, "..\\evil.txt", "safe content", "/")

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusMovedPermanently, rec.Body.String())
	}
	if got, err := os.ReadFile(filepath.Join(root, "evil.txt")); err != nil || string(got) != "safe content" {
		t.Fatalf("sanitized file = %q, err=%v; want content", string(got), err)
	}
	if _, err := os.Stat(filepath.Join(root, "..\\evil.txt")); !os.IsNotExist(err) {
		t.Fatalf("unsanitized filename was used, stat err=%v", err)
	}
}

func TestDownloaderReadsFromRuntimeRootPath(t *testing.T) {
	resetCoreRuntimeTest(t)

	globalRoot := t.TempDir()
	runtimeRoot := t.TempDir()
	conf.Config.RootPath = globalRoot
	if err := os.WriteFile(filepath.Join(globalRoot, "hello.txt"), []byte("global"), 0o600); err != nil {
		t.Fatalf("write global file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runtimeRoot, "hello.txt"), []byte("runtime"), 0o600); err != nil {
		t.Fatalf("write runtime file: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: runtimeRoot,
	}))

	r := gin.New()
	r.GET("/file/hello.txt", func(c *gin.Context) {
		c.Set("filename", "hello.txt")
		c.Set("isDirectory", false)
		Downloader(c)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/file/hello.txt", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != "runtime" {
		t.Fatalf("body = %q, want runtime", got)
	}
}

func TestDownloaderWritesDownloadSuccessAfterOpeningFile(t *testing.T) {
	logger := resetCoreRuntimeTest(t)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("runtime"), 0o600); err != nil {
		t.Fatalf("write runtime file: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
	}))

	rec := serveDownloadRequest("/file/hello.txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}

	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventDownload})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Read returned %d download events, want 1: %+v", len(events), events)
	}
	if events[0].Path != "/file/hello.txt" || events[0].Status != http.StatusOK || events[0].Result != accesslog.ResultSuccess {
		t.Fatalf("download event = %+v, want success for /file/hello.txt", events[0])
	}
}

func TestDownloaderWritesFailureWhenResponseWriteFails(t *testing.T) {
	logger := resetCoreRuntimeTest(t)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("runtime"), 0o600); err != nil {
		t.Fatalf("write runtime file: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
	}))

	writeErr := errors.New("response write failed")
	r := gin.New()
	r.GET("/file/*filename", middleware.CheckLevel, func(c *gin.Context) {
		c.Writer = failingResponseWriter{ResponseWriter: c.Writer, err: writeErr}
		Downloader(c)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/file/hello.txt", nil))

	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventDownload})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Read returned %d download events, want 1: %+v", len(events), events)
	}
	if events[0].Result != accesslog.ResultFailure || events[0].Status != http.StatusInternalServerError {
		t.Fatalf("download event = %+v, want transfer failure", events[0])
	}
}

func TestDownloaderWritesDownloadSuccessAfterZippingDirectory(t *testing.T) {
	logger := resetCoreRuntimeTest(t)

	root := t.TempDir()
	dir := filepath.Join(root, "docs")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("zip me"), 0o600); err != nil {
		t.Fatalf("write docs file: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 1,
	}))

	rec := serveDownloadRequest("/file/docs?action=dl")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}

	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventDownload})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Read returned %d download events, want 1: %+v", len(events), events)
	}
	if events[0].Path != "/file/docs" || events[0].Status != http.StatusOK || events[0].Result != accesslog.ResultSuccess {
		t.Fatalf("download event = %+v, want zip success for /file/docs", events[0])
	}
}

func TestDownloaderUsesSnapshotFromCheckLevelContext(t *testing.T) {
	resetCoreRuntimeTest(t)

	rootA := t.TempDir()
	rootB := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootA, "hello.txt"), []byte("root-a"), 0o600); err != nil {
		t.Fatalf("write rootA file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootB, "hello.txt"), []byte("root-b"), 0o600); err != nil {
		t.Fatalf("write rootB file: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: rootA,
		MaxLevel: 0,
	}))

	r := gin.New()
	r.GET("/file/*filename", middleware.CheckLevel, func(c *gin.Context) {
		runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
			RootPath: rootB,
			MaxLevel: 0,
		}))
		Downloader(c)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/file/hello.txt", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != "root-a" {
		t.Fatalf("body = %q, want root-a", got)
	}
}

func TestCheckLevelAndDownloaderPreserveFilePathSegment(t *testing.T) {
	resetCoreRuntimeTest(t)

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "file"), 0o755); err != nil {
		t.Fatalf("mkdir file dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("root file"), 0o600); err != nil {
		t.Fatalf("write root hello: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "file", "hello.txt"), []byte("nested file"), 0o600); err != nil {
		t.Fatalf("write nested hello: %v", err)
	}
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 1,
	}))

	r := gin.New()
	r.GET("/file/*filename", middleware.CheckLevel, Downloader)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/file/file/hello.txt", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); got != "nested file" {
		t.Fatalf("body = %q, want nested file", got)
	}
}

func TestCheckLevelAndDownloaderRejectSymlinkEscape(t *testing.T) {
	resetCoreRuntimeTest(t)

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "secret.txt")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(outside, []byte("outside secret"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	symlinkOrSkip(t, outside, filepath.Join(root, "link.txt"))
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 0,
	}))

	rec := serveDownloadRequest("/file/link.txt")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "outside secret") {
		t.Fatalf("response leaked symlink target content: %q", rec.Body.String())
	}
}

func TestCheckLevelAndDownloaderMissingFileDoesNotPanic(t *testing.T) {
	resetCoreRuntimeTest(t)

	root := t.TempDir()
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: root,
		MaxLevel: 0,
	}))

	rec := serveDownloadRequest("/file/missing.txt")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func TestUploaderRejectsParentSymlinkEscape(t *testing.T) {
	resetCoreRuntimeTest(t)

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	symlinkOrSkip(t, outside, filepath.Join(root, "uploads"))
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: true,
	}))

	rec := serveUploadRequest(t, "escape.txt", "escape", "uploads")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(outside, "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("escaped upload touched outside file, stat err=%v", err)
	}
}

func TestUploaderRejectsExistingTargetSymlink(t *testing.T) {
	resetCoreRuntimeTest(t)

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside.txt")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(outside, []byte("original"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	symlinkOrSkip(t, outside, filepath.Join(root, "target.txt"))
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: true,
	}))

	rec := serveUploadRequest(t, "target.txt", "overwrite", "/")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if got, err := os.ReadFile(outside); err != nil || string(got) != "original" {
		t.Fatalf("outside file = %q, err=%v; want original", string(got), err)
	}
}

func TestUploaderWritesRejectAndSuccessEvents(t *testing.T) {
	logger := resetCoreRuntimeTest(t)

	root := t.TempDir()
	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: false,
	}))
	blocked := serveUploadRequest(t, "blocked.txt", "blocked", "/")
	if blocked.Code != http.StatusInternalServerError {
		t.Fatalf("blocked status = %d, want %d; body=%q", blocked.Code, http.StatusInternalServerError, blocked.Body.String())
	}

	runtimeconf.SetCurrent(runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      root,
		IsAllowUpload: true,
	}))
	accepted := serveUploadRequest(t, "accepted.txt", "accepted", "/")
	if accepted.Code != http.StatusMovedPermanently {
		t.Fatalf("accepted status = %d, want %d; body=%q", accepted.Code, http.StatusMovedPermanently, accepted.Body.String())
	}

	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventUpload})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Read returned %d upload events, want 2: %+v", len(events), events)
	}
	if events[0].Result != accesslog.ResultReject || events[0].Status != http.StatusInternalServerError {
		t.Fatalf("first upload event = %+v, want reject", events[0])
	}
	if events[1].Result != accesslog.ResultSuccess || events[1].Status != http.StatusMovedPermanently {
		t.Fatalf("second upload event = %+v, want success", events[1])
	}
}

func serveUploadRequest(t *testing.T, filename, content, currPath string) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("path", currPath); err != nil {
		t.Fatalf("write path field: %v", err)
	}
	part, err := writer.CreateFormFile("upload_file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.WriteString(part, content); err != nil {
		t.Fatalf("write upload content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	r := gin.New()
	r.POST("/upload", Uploader)

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func serveDownloadRequest(target string) *httptest.ResponseRecorder {
	r := gin.New()
	r.GET("/file/*filename", middleware.CheckLevel, Downloader)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func symlinkOrSkip(t *testing.T, oldname, newname string) {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		t.Skipf("os.Symlink unsupported or not permitted: %v", err)
	}
}

type failingResponseWriter struct {
	gin.ResponseWriter
	err error
}

func (w failingResponseWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func (w failingResponseWriter) WriteString(string) (int, error) {
	return 0, w.err
}

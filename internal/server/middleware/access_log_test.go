package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/TCP404/OneTiny-cli/internal/accesslog"
	"github.com/gin-gonic/gin"
)

func TestAccessLogClassifiesSuccessRejectAndErrorEvents(t *testing.T) {
	originalMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(originalMode)
	})

	logger := accesslog.New(filepath.Join(t.TempDir(), "access.log"))
	restore := accesslog.SetLoggerForTest(logger)
	t.Cleanup(restore)

	r := gin.New()
	r.Use(AccessLog())
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error")
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ok", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/error", nil))

	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventAccess})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Read returned %d access events, want 1: %+v", len(events), events)
	}
	if events[0].Path != "/ok" || events[0].Status != http.StatusOK || events[0].Result != accesslog.ResultSuccess {
		t.Fatalf("access event = %+v, want /ok success", events[0])
	}

	rejects, err := logger.Read(accesslog.Filter{Event: accesslog.EventReject})
	if err != nil {
		t.Fatalf("Read reject returned error: %v", err)
	}
	if len(rejects) != 1 {
		t.Fatalf("Read returned %d reject events, want 1: %+v", len(rejects), rejects)
	}
	if rejects[0].Path != "/missing" || rejects[0].Status != http.StatusNotFound || rejects[0].Result != accesslog.ResultReject {
		t.Fatalf("reject event = %+v, want /missing reject", rejects[0])
	}

	errors, err := logger.Read(accesslog.Filter{Event: accesslog.EventError})
	if err != nil {
		t.Fatalf("Read error returned error: %v", err)
	}
	if len(errors) != 1 {
		t.Fatalf("Read returned %d error events, want 1: %+v", len(errors), errors)
	}
	if errors[0].Path != "/error" || errors[0].Status != http.StatusInternalServerError || errors[0].Result != accesslog.ResultFailure {
		t.Fatalf("error event = %+v, want /error failure", errors[0])
	}
}

func TestAccessLogWritesFailureEventWhenRecoveryHandlesPanic(t *testing.T) {
	originalMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(originalMode)
	})

	logger := accesslog.New(filepath.Join(t.TempDir(), "access.log"))
	restore := accesslog.SetLoggerForTest(logger)
	t.Cleanup(restore)

	r := gin.New()
	r.Use(gin.Recovery(), AccessLog())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventError})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Read returned %d error events, want 1: %+v", len(events), events)
	}
	if events[0].Path != "/panic" || events[0].Status != http.StatusInternalServerError || events[0].Result != accesslog.ResultFailure {
		t.Fatalf("panic error event = %+v, want 500 failure for /panic", events[0])
	}
}

func TestSetupPanicRecordsSingleAccessFailure(t *testing.T) {
	originalMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(originalMode)
	})

	logger := accesslog.New(filepath.Join(t.TempDir(), "access.log"))
	restore := accesslog.SetLoggerForTest(logger)
	t.Cleanup(restore)

	r := gin.New()
	Setup(r)
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	events, err := logger.Read(accesslog.Filter{Event: accesslog.EventError})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Read returned %d error events, want 1: %+v", len(events), events)
	}
	if events[0].Path != "/panic" || events[0].Status != http.StatusInternalServerError || events[0].Result != accesslog.ResultFailure {
		t.Fatalf("setup panic error event = %+v, want one 500 failure", events[0])
	}
}

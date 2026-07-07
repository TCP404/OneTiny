package routes

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogoRouteServesWhiteBackgroundPNG(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	loadLogoRoute(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/logo.png", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("Content-Type = %q, want image/png", got)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("logo.png is not a valid PNG: %v", err)
	}
	if cfg.Width != 512 || cfg.Height != 512 {
		t.Fatalf("logo size = %dx%d, want 512x512", cfg.Width, cfg.Height)
	}
}

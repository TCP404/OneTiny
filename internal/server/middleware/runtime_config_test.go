package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/runtime"
)

func TestRuntimeSnapshotSetsRequestContext(t *testing.T) {
	rt := runtime.New(runtime.Snapshot{RootPath: "/tmp/root", Port: 8192, SessionVal: "session"})
	r := gin.New()
	r.Use(RuntimeSnapshot(rt))
	r.GET("/", func(c *gin.Context) {
		got := currentSnapshot(c)
		if got.RootPath != "/tmp/root" {
			t.Fatalf("RootPath = %q, want /tmp/root", got.RootPath)
		}
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

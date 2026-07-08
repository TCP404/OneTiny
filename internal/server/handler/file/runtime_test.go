package file

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/runtime"
)

func TestCurrentSnapshotReadsRequestContext(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set(runtime.ContextKey, runtime.Snapshot{RootPath: "/tmp/root", Port: 8192})

	got := currentSnapshot(c)
	if got.RootPath != "/tmp/root" {
		t.Fatalf("RootPath = %q, want /tmp/root", got.RootPath)
	}
}

func TestCurrentSnapshotWithoutContextReturnsZeroValue(t *testing.T) {
	got := currentSnapshot(nil)
	if got != (runtime.Snapshot{}) {
		t.Fatalf("snapshot = %+v, want zero value", got)
	}
}

func TestUploaderRejectsWhenUploadDisabled(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/upload", nil)
	c.Set(runtime.ContextKey, runtime.Snapshot{RootPath: t.TempDir(), IsAllowUpload: false})

	Uploader(c)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

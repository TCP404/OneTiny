package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/scratch"
	"github.com/tcp404/OneTiny/internal/server/middleware"
)

func TestScratchRouteRequiresLoginWhenSecure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, err := scratch.NewStore(scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	rt := runtime.New(runtime.Snapshot{
		IsSecure:            true,
		SessionVal:          "session",
		ScratchMaxItems:     10,
		ScratchMaxItemBytes: 1024,
	})
	router := gin.New()
	middleware.Setup(router, rt, nil)
	if err := Setup(router, store); err != nil {
		t.Fatalf("Setup returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/scratch/items", strings.NewReader(url.Values{
		"kind": {"text"},
		"text": {"unauthorized"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusTemporaryRedirect, rec.Body.String())
	}
	if location := rec.Header().Get("Location"); location != "/login" {
		t.Fatalf("Location = %q, want /login", location)
	}
	if got := len(store.List()); got != 0 {
		t.Fatalf("scratch store item count = %d, want 0", got)
	}
}

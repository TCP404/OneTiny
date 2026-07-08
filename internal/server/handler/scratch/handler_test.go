package scratchhandler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/scratch"
)

func TestCreateTextItemFromFormReturnsJSON(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	router := newTestRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/scratch/items", strings.NewReader(url.Values{
		"kind": {"text"},
		"text": {"hello"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json response: %v", err)
	}
	if payload.ID == "" {
		t.Fatalf("response id is empty")
	}
	items := store.List()
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Kind != scratch.KindText {
		t.Fatalf("item kind = %q, want %q", items[0].Kind, scratch.KindText)
	}
	if string(items[0].Data) != "hello" {
		t.Fatalf("item data = %q, want %q", string(items[0].Data), "hello")
	}
}

func TestCreateImageItemFromMultipartDetectsPNG(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	router := newTestRouter(store)
	body, contentType := multipartBody(t, "image", "image.png", pngBytes())

	req := httptest.NewRequest(http.MethodPost, "/scratch/items", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	items := store.List()
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Kind != scratch.KindImage {
		t.Fatalf("item kind = %q, want %q", items[0].Kind, scratch.KindImage)
	}
	if items[0].MimeType != "image/png" {
		t.Fatalf("mime type = %q, want image/png", items[0].MimeType)
	}
}

func TestCreateImageItemDetectsSupportedMIMETypes(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		data     []byte
		wantMIME string
	}{
		{name: "jpeg", fileName: "image.jpg", data: []byte{0xff, 0xd8, 0xff, 0xdb, 0x00}, wantMIME: "image/jpeg"},
		{name: "gif", fileName: "image.gif", data: []byte("GIF89a\x01\x00\x01\x00"), wantMIME: "image/gif"},
		{name: "webp", fileName: "image.webp", data: []byte("RIFF\x00\x00\x00\x00WEBPVP8 "), wantMIME: "image/webp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
			router := newTestRouter(store)
			body, contentType := multipartBody(t, "image", tt.fileName, tt.data)

			req := httptest.NewRequest(http.MethodPost, "/scratch/items", body)
			req.Header.Set("Content-Type", contentType)
			req.Header.Set("Accept", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
			}
			items := store.List()
			if len(items) != 1 {
				t.Fatalf("len(items) = %d, want 1", len(items))
			}
			if items[0].MimeType != tt.wantMIME {
				t.Fatalf("mime type = %q, want %q", items[0].MimeType, tt.wantMIME)
			}
		})
	}
}

func TestCreateImageRejectsOversizedPayloadBeforeTypeDetection(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 8})
	router := newTestRouter(store)
	body, contentType := multipartBody(t, "image", "large.bin", []byte("not-an-image"))

	req := httptest.NewRequest(http.MethodPost, "/scratch/items", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if len(store.List()) != 0 {
		t.Fatalf("store changed after oversized image add: %+v", store.List())
	}
}

func TestGetItemReturnsStoredContent(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	item, err := store.Add(scratch.KindText, scratch.TextMIME, []byte("hello"))
	if err != nil {
		t.Fatalf("store.Add returned error: %v", err)
	}
	router := newTestRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/scratch/items/"+item.ID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != scratch.TextMIME {
		t.Fatalf("Content-Type = %q, want %q", got, scratch.TextMIME)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body = %q, want hello", rec.Body.String())
	}
}

func TestGetItemDownloadSetsContentDisposition(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	item, err := store.Add(scratch.KindText, scratch.TextMIME, []byte("hello"))
	if err != nil {
		t.Fatalf("store.Add returned error: %v", err)
	}
	router := newTestRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/scratch/items/"+item.ID+"?download=1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := "attachment; filename=" + item.ID
	if got := rec.Header().Get("Content-Disposition"); got != want {
		t.Fatalf("Content-Disposition = %q, want %q", got, want)
	}
}

func TestCreateTextRejectsEmptyContent(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	router := newTestRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/scratch/items", strings.NewReader(url.Values{
		"kind": {"text"},
		"text": {""},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateRejectsOversizedContent(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 3})
	router := newTestRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/scratch/items", strings.NewReader(url.Values{
		"kind": {"text"},
		"text": {"toolong"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
}

func TestGetItemNotFound(t *testing.T) {
	store := newTestStore(t, scratch.Limits{MaxItems: 10, MaxItemBytes: 1024})
	router := newTestRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/scratch/items/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestIndexWithNilStoreDoesNotPanic(t *testing.T) {
	router := newTestRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/scratch/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func newTestStore(t *testing.T, limits scratch.Limits) *scratch.Store {
	t.Helper()
	store, err := scratch.NewStore(limits)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	return store
}

func newTestRouter(store *scratch.Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/scratch")
	Register(group, store)
	return router
}

func multipartBody(t *testing.T, fieldName, fileName string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("kind", "image"); err != nil {
		t.Fatalf("write kind field: %v", err)
	}
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func pngBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde,
	}
}

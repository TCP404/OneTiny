package scratchhandler

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/scratch"
)

const scratchTemplate = "scratch.tpl"

const multipartOverheadBytes = 1 << 20
const textPreviewBytes = 4096

type handler struct {
	store *scratch.Store
}

func Register(group *gin.RouterGroup, store *scratch.Store) {
	h := handler{store: store}
	group.GET("/", h.index)
	group.POST("/items", h.create)
	group.GET("/items/:id", h.get)
}

func (h handler) index(c *gin.Context) {
	if h.store == nil {
		c.String(http.StatusInternalServerError, "临时列表不可用")
		return
	}
	h.renderIndex(c, http.StatusOK, "")
}

func (h handler) create(c *gin.Context) {
	if h.store == nil {
		h.handleCreateError(c, http.StatusInternalServerError, "临时列表不可用")
		return
	}

	if status, err := h.prepareCreateRequest(c); err != nil {
		h.handleCreateError(c, status, err.Error())
		return
	}

	kind, mimeType, data, status, err := h.readCreatePayload(c)
	if err != nil {
		h.handleCreateError(c, status, err.Error())
		return
	}

	item, err := h.store.Add(kind, mimeType, data)
	if err != nil {
		h.handleCreateError(c, createStatus(err), err.Error())
		return
	}

	if wantsJSON(c) {
		c.JSON(http.StatusOK, gin.H{"id": item.ID})
		return
	}
	c.Redirect(http.StatusSeeOther, "/scratch/")
}

func (h handler) get(c *gin.Context) {
	if h.store == nil {
		c.String(http.StatusInternalServerError, "临时列表不可用")
		return
	}

	item, ok := h.store.Get(c.Param("id"))
	if !ok {
		c.String(http.StatusNotFound, scratch.ErrItemNotFound.Error())
		return
	}

	if c.Query("download") == "1" {
		c.Header("Content-Disposition", "attachment; filename="+downloadFilename(item))
	}
	c.Data(http.StatusOK, item.MimeType, item.Data)
}

func (h handler) readCreatePayload(c *gin.Context) (scratch.Kind, string, []byte, int, error) {
	if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
		return h.readJSONPayload(c)
	}

	switch c.PostForm("kind") {
	case string(scratch.KindText):
		return h.readTextPayload(c)
	case string(scratch.KindImage):
		return h.readImagePayload(c)
	default:
		return "", "", nil, http.StatusBadRequest, scratch.ErrUnsupportedType
	}
}

func (h handler) readTextPayload(c *gin.Context) (scratch.Kind, string, []byte, int, error) {
	text := c.PostForm("text")
	return scratch.KindText, scratch.TextMIME, []byte(text), http.StatusBadRequest, nil
}

func (h handler) readJSONPayload(c *gin.Context) (scratch.Kind, string, []byte, int, error) {
	var payload struct {
		Kind string `json:"kind"`
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		if isMaxBytesError(err) {
			return "", "", nil, http.StatusRequestEntityTooLarge, scratch.ErrItemTooLarge
		}
		return "", "", nil, http.StatusBadRequest, err
	}
	switch scratch.Kind(payload.Kind) {
	case scratch.KindText:
		return scratch.KindText, scratch.TextMIME, []byte(payload.Text), http.StatusBadRequest, nil
	default:
		return "", "", nil, http.StatusBadRequest, scratch.ErrUnsupportedType
	}
}

func (h handler) prepareCreateRequest(c *gin.Context) (int, error) {
	if h.store == nil {
		return http.StatusInternalServerError, errors.New("临时列表不可用")
	}

	limit := h.store.Limits().MaxItemBytes
	if limit > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, createBodyLimit(limit))
	}

	contentType := c.GetHeader("Content-Type")
	switch {
	case strings.Contains(contentType, "multipart/form-data"):
		if err := c.Request.ParseMultipartForm(createBodyLimit(limit)); err != nil {
			if isMaxBytesError(err) {
				return http.StatusRequestEntityTooLarge, scratch.ErrItemTooLarge
			}
			return http.StatusBadRequest, scratch.ErrUnsupportedType
		}
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		if err := c.Request.ParseForm(); err != nil {
			if isMaxBytesError(err) {
				return http.StatusRequestEntityTooLarge, scratch.ErrItemTooLarge
			}
			return http.StatusBadRequest, err
		}
	}

	return 0, nil
}

func (h handler) readImagePayload(c *gin.Context) (scratch.Kind, string, []byte, int, error) {
	file, err := c.FormFile("image")
	if err != nil {
		if isMaxBytesError(err) {
			return "", "", nil, http.StatusRequestEntityTooLarge, scratch.ErrItemTooLarge
		}
		return "", "", nil, http.StatusBadRequest, scratch.ErrEmptyContent
	}

	opened, err := file.Open()
	if err != nil {
		return "", "", nil, http.StatusInternalServerError, err
	}
	defer opened.Close()

	limit := h.store.Limits().MaxItemBytes
	data, err := io.ReadAll(io.LimitReader(opened, int64(limit)+1))
	if err != nil {
		return "", "", nil, http.StatusInternalServerError, err
	}
	if limit > 0 && len(data) > limit {
		return "", "", nil, http.StatusRequestEntityTooLarge, scratch.ErrItemTooLarge
	}
	if len(data) == 0 {
		return "", "", nil, http.StatusBadRequest, scratch.ErrEmptyContent
	}

	mimeType := normalizeMIMEType(http.DetectContentType(data))
	if !strings.HasPrefix(mimeType, "image/") {
		return "", "", nil, http.StatusBadRequest, scratch.ErrUnsupportedType
	}
	return scratch.KindImage, mimeType, data, http.StatusBadRequest, nil
}

func createBodyLimit(limit int) int64 {
	return int64(limit) + multipartOverheadBytes
}

func isMaxBytesError(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

func normalizeMIMEType(value string) string {
	mediaType, _, err := mime.ParseMediaType(value)
	if err == nil {
		value = mediaType
	}
	return strings.ToLower(value)
}

func downloadFilename(item scratch.Item) string {
	return item.ID + extensionForItem(item)
}

func extensionForItem(item scratch.Item) string {
	switch item.Kind {
	case scratch.KindText:
		return ".txt"
	case scratch.KindImage:
		switch item.MimeType {
		case "image/png":
			return ".png"
		case "image/jpeg":
			return ".jpg"
		case "image/gif":
			return ".gif"
		case "image/webp":
			return ".webp"
		}
	}
	return ""
}

func (h handler) handleCreateError(c *gin.Context, status int, message string) {
	if wantsJSON(c) {
		c.JSON(status, gin.H{"error": message})
		return
	}
	if h.store == nil {
		c.String(status, message)
		return
	}
	h.renderIndex(c, status, message)
}

func (h handler) renderIndex(c *gin.Context, status int, message string) {
	data := gin.H{
		"items":  []scratch.Summary{},
		"limits": scratch.Limits{},
	}
	if h.store != nil {
		data["items"] = h.store.ListSummaries(textPreviewBytes)
		data["limits"] = h.store.Limits()
	} else if message == "" {
		message = "临时列表不可用"
		status = http.StatusInternalServerError
	}
	if message != "" {
		data["error"] = message
	}
	c.HTML(status, scratchTemplate, data)
}

func wantsJSON(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Accept"), "application/json") ||
		strings.Contains(c.GetHeader("Content-Type"), "application/json")
}

func createStatus(err error) int {
	switch {
	case errors.Is(err, scratch.ErrItemTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, scratch.ErrEmptyContent), errors.Is(err, scratch.ErrUnsupportedType):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

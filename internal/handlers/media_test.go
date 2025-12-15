package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServeFileHandler_ServesFile(t *testing.T) {
	// Create temp dir and file
	tmpDir, err := os.MkdirTemp("", "uploads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := "1_1765789611227708560.mp4"
	content := "hello world"
	filePath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	mh := NewMediaHandler(nil)
	mh.uploadDir = tmpDir

	// Set up router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/media/files/:name", mh.ServeFileHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/media/files/"+filename, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if strings.TrimSpace(string(body)) != content {
		t.Fatalf("unexpected file content: %q", string(body))
	}
}

func TestServeFileHandler_InvalidFilename(t *testing.T) {
	mh := NewMediaHandler(nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/media/files/:name", mh.ServeFileHandler)

	// Path traversal attempt
	req := httptest.NewRequest(http.MethodGet, "/api/media/files/../secret", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid filename, got %d", w.Code)
	}
}

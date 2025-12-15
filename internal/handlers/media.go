package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ristep/smanzy_backend/internal/models"
	"gorm.io/gorm"
)

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	db        *gorm.DB
	uploadDir string
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(db *gorm.DB) *MediaHandler {
	// Ensure upload directory exists
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		_ = os.MkdirAll(uploadDir, 0755)
	}

	return &MediaHandler{
		db:        db,
		uploadDir: uploadDir,
	}
}

// UploadHandler handles file uploads
func (mh *MediaHandler) UploadHandler(c *gin.Context) {
	// Get current user
	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	user := authUser.(*models.User)

	// Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No file uploaded"})
		return
	}

	// Generate unique stored name
	ext := filepath.Ext(file.Filename)
	uniqueName := fmt.Sprintf("%d_%d%s", user.ID, time.Now().UnixNano(), ext)
	dst := filepath.Join(mh.uploadDir, uniqueName)

	// Save file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save file"})
		return
	}

	// Create media record
	media := models.Media{
		Filename:   file.Filename,
		StoredName: uniqueName,
		URL:        "/api/media/files/" + uniqueName, // This will need a static file server handler or direct streaming
		Type:       "file",                           // Simplified, could verify mime type
		MimeType:   file.Header.Get("Content-Type"),
		Size:       file.Size,
		UserID:     user.ID,
	}

	if err := mh.db.Create(&media).Error; err != nil {
		// Clean up file if DB save fails
		_ = os.Remove(dst)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save media record"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Data: media})
}

// GetMediaHandler downloads/streams the file
func (mh *MediaHandler) GetMediaHandler(c *gin.Context) {
	mediaID := c.Param("id")

	var media models.Media
	if err := mh.db.First(&media, mediaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Any authenticated user can download (per requirement "others only read access")
	// If we wanted strictly public, we'd skip AuthMiddleware for this route,
	// but requirement implies "others" (other users) have read access.

	filePath := filepath.Join(mh.uploadDir, media.StoredName)
	c.File(filePath)
}

// ListPublicMediasHandler returns a paginated list of medias for public consumption
// Query params: limit (default 100), offset (default 0)
func (mh *MediaHandler) ListPublicMediasHandler(c *gin.Context) {
	limit := 100
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	var medias []models.Media
	if err := mh.db.Select("id, filename, url, type, mime_type, size, created_at").Order("created_at desc").Limit(limit).Offset(offset).Find(&medias).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: medias})
}

// UpdateMediaRequest represents payload for updating media
type UpdateMediaRequest struct {
	Filename string `json:"filename"`
}

// UpdateMediaHandler updates media metadata
func (mh *MediaHandler) UpdateMediaHandler(c *gin.Context) {
	mediaID := c.Param("id")
	var req UpdateMediaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Get current user
	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	user := authUser.(*models.User)

	var media models.Media
	if err := mh.db.Preload("UploadedBy").First(&media, mediaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Access Control: Owner or Admin
	if media.UserID != user.ID && !user.HasRole("admin") {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	// Update fields
	if req.Filename != "" {
		media.Filename = req.Filename
	}

	if err := mh.db.Save(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update media"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: media})
}

// DeleteMediaHandler deletes media file and record
func (mh *MediaHandler) DeleteMediaHandler(c *gin.Context) {
	mediaID := c.Param("id")

	// Get current user
	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	user := authUser.(*models.User)

	var media models.Media
	if err := mh.db.First(&media, mediaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Access Control: Owner or Admin
	if media.UserID != user.ID && !user.HasRole("admin") {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	// Delete file from disk
	filePath := filepath.Join(mh.uploadDir, media.StoredName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		// Log error but continue to delete DB record?
		// Or fail? Usually better to try to clean up DB even if file is gone,
		// but if file deletion fails due to permission, we might want to know.
		// For now, let's proceed to delete auth record so we don't have dangling refs.
		fmt.Printf("Warning: Failed to delete file %s: %v\n", filePath, err)
	}

	// Delete from DB
	if err := mh.db.Delete(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete media record"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: map[string]string{"message": "Media deleted successfully"}})
}

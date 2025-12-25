package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ristep/smanzy_backend/internal/models"
	"github.com/ristep/smanzy_backend/internal/services"
	"gorm.io/gorm"
)

// AlbumHandler handles album-related HTTP requests
type AlbumHandler struct {
	albumService *services.AlbumService
}

// NewAlbumHandler creates a new album handler
func NewAlbumHandler(db *gorm.DB) *AlbumHandler {
	return &AlbumHandler{
		albumService: services.NewAlbumService(db),
	}
}

// CreateAlbumHandler handles creating a new album
func (ah *AlbumHandler) CreateAlbumHandler(c *gin.Context) {
	// Get current user
	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	user := authUser.(*models.User)

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	album, err := ah.albumService.CreateAlbum(user.ID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, album)
}

// GetAlbumHandler retrieves a specific album by ID
func (ah *AlbumHandler) GetAlbumHandler(c *gin.Context) {
	albumID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid album ID"})
		return
	}

	album, err := ah.albumService.GetAlbumByID(uint(albumID))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, album)
}

// GetUserAlbumsHandler retrieves all albums for the current user
func (ah *AlbumHandler) GetUserAlbumsHandler(c *gin.Context) {
	// Get current user
	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	user := authUser.(*models.User)

	albums, err := ah.albumService.GetUserAlbums(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if albums == nil {
		albums = []models.Album{}
	}

	c.JSON(http.StatusOK, albums)
}

// UpdateAlbumHandler updates an album's details
func (ah *AlbumHandler) UpdateAlbumHandler(c *gin.Context) {
	albumID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid album ID"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	album, err := ah.albumService.UpdateAlbum(uint(albumID), req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, album)
}

// AddMediaToAlbumHandler adds a media file to an album
func (ah *AlbumHandler) AddMediaToAlbumHandler(c *gin.Context) {
	albumID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid album ID"})
		return
	}

	var req struct {
		MediaID uint `json:"media_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := ah.albumService.AddMediaToAlbum(uint(albumID), req.MediaID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Media added to album successfully"})
}

// RemoveMediaFromAlbumHandler removes a media file from an album
func (ah *AlbumHandler) RemoveMediaFromAlbumHandler(c *gin.Context) {
	albumID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid album ID"})
		return
	}

	var req struct {
		MediaID uint `json:"media_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := ah.albumService.RemoveMediaFromAlbum(uint(albumID), req.MediaID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Media removed from album successfully"})
}

// DeleteAlbumHandler soft deletes an album
func (ah *AlbumHandler) DeleteAlbumHandler(c *gin.Context) {
	albumID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid album ID"})
		return
	}

	if err := ah.albumService.DeleteAlbum(uint(albumID)); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Album deleted successfully"})
}

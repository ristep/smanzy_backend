package services

import (
	"errors"

	"github.com/ristep/smanzy_backend/internal/models"
	"gorm.io/gorm"
)

// AlbumService handles business logic for album operations
type AlbumService struct {
	db *gorm.DB
}

// NewAlbumService creates a new album service
func NewAlbumService(db *gorm.DB) *AlbumService {
	return &AlbumService{db: db}
}

// CreateAlbum creates a new album for a user
func (as *AlbumService) CreateAlbum(userID uint, title, description string) (*models.Album, error) {
	if title == "" {
		return nil, errors.New("album title is required")
	}

	album := models.Album{
		Title:       title,
		Description: description,
		UserID:      userID,
	}

	if err := as.db.Create(&album).Error; err != nil {
		return nil, err
	}

	return &album, nil
}

// GetAlbumByID retrieves an album by its ID
func (as *AlbumService) GetAlbumByID(albumID uint) (*models.Album, error) {
	var album models.Album
	if err := as.db.Preload("MediaFiles").First(&album, albumID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("album not found")
		}
		return nil, err
	}
	return &album, nil
}

// GetUserAlbums retrieves all albums for a user
func (as *AlbumService) GetUserAlbums(userID uint) ([]models.Album, error) {
	var albums []models.Album
	if err := as.db.Where("user_id = ?", userID).
		Preload("MediaFiles").
		Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}

// UpdateAlbum updates an album's title and description
func (as *AlbumService) UpdateAlbum(albumID uint, title, description string) (*models.Album, error) {
	album, err := as.GetAlbumByID(albumID)
	if err != nil {
		return nil, err
	}

	if title != "" {
		album.Title = title
	}

	album.Description = description

	if err := as.db.Save(album).Error; err != nil {
		return nil, err
	}

	return album, nil
}

// AddMediaToAlbum adds a media file to an album
func (as *AlbumService) AddMediaToAlbum(albumID, mediaID uint) error {
	// Verify album exists
	if err := as.db.Model(&models.Album{}).Where("id = ?", albumID).First(&models.Album{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("album not found")
		}
		return err
	}

	// Verify media exists
	if err := as.db.Model(&models.Media{}).Where("id = ?", mediaID).First(&models.Media{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("media not found")
		}
		return err
	}

	// Load the album first
	album := &models.Album{}
	if err := as.db.First(album, albumID).Error; err != nil {
		return err
	}

	// Create media reference with just the ID
	media := &models.Media{ID: mediaID}

	// Add media to album using association
	if err := as.db.Model(album).Association("MediaFiles").Append(media); err != nil {
		return err
	}

	return nil
}

// RemoveMediaFromAlbum removes a media file from an album
func (as *AlbumService) RemoveMediaFromAlbum(albumID, mediaID uint) error {
	// Verify album exists
	if err := as.db.Model(&models.Album{}).Where("id = ?", albumID).First(&models.Album{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("album not found")
		}
		return err
	}

	// Load the album first
	album := &models.Album{}
	if err := as.db.First(album, albumID).Error; err != nil {
		return err
	}

	// Remove media from album using association
	if err := as.db.Model(album).Association("MediaFiles").Delete(&models.Media{ID: mediaID}); err != nil {
		return err
	}

	return nil
}

// DeleteAlbum performs a soft delete on an album
func (as *AlbumService) DeleteAlbum(albumID uint) error {
	album, err := as.GetAlbumByID(albumID)
	if err != nil {
		return err
	}

	if err := as.db.Delete(album).Error; err != nil {
		return err
	}

	return nil
}

// PermanentlyDeleteAlbum permanently deletes an album from the database
func (as *AlbumService) PermanentlyDeleteAlbum(albumID uint) error {
	album, err := as.GetAlbumByID(albumID)
	if err != nil {
		return err
	}

	// First, clear all associated media
	if err := as.db.Model(album).Association("MediaFiles").Clear(); err != nil {
		return err
	}

	// Then permanently delete the album
	if err := as.db.Unscoped().Delete(album).Error; err != nil {
		return err
	}

	return nil
}

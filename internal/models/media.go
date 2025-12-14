package models

import "gorm.io/gorm"

// Media represents a media file in the system
type Media struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Filename   string         `gorm:"not null" json:"filename"`
	StoredName string         `gorm:"not null" json:"stored_name"`
	URL        string         `gorm:"not null" json:"url"`
	Type       string         `json:"type"`      // e.g., "image", "video"
	MimeType   string         `json:"mime_type"` // e.g., "image/jpeg"
	Size       int64          `json:"size"`      // in bytes
	UserID     uint           `json:"user_id"`   // Uploaded by
	UploadedBy User           `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt  int64          `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt  int64          `gorm:"autoUpdateTime:milli" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Media
func (Media) TableName() string {
	return "media"
}

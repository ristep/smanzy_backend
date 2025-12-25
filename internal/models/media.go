package models

import "gorm.io/gorm"

// Media represents a media file uploaded to the system
// It tracks file metadata and links to the physical file storage.
type Media struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Filename   string `gorm:"not null" json:"filename"`    // Original name of the file
	StoredName string `gorm:"not null" json:"stored_name"` // Unique name on disk (to prevent overwrites)
	URL        string `gorm:"not null" json:"url"`         // Public URL to access the file

	Type     string `json:"type"`      // General category (e.g., "image", "video")
	MimeType string `json:"mime_type"` // Specific MIME type (e.g., "image/jpeg", "application/pdf")
	Size     int64  `json:"size"`      // File size in bytes

	// Foreign Keys
	// UserID links this media file to a specific User
	UserID uint `json:"user_id"`

	// UploadedBy is the actual User struct that corresponds to UserID.
	// gorm:"foreignKey:UserID" tells GORM how to link the two.
	// json:"-" prevents endless recursion or exposing too much user info when listing media.
	UploadedBy User `gorm:"foreignKey:UserID" json:"-"`

	// Albums represents the many-to-many relationship with Album
	// A media file can belong to multiple albums, and an album can contain multiple media files
	// "many2many:album_media" tells GORM to use the join table named "album_media"
	// json:"-" prevents endless recursion when listing media
	Albums []Album `gorm:"many2many:album_media;" json:"-"`

	CreatedAt int64          `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Media
func (Media) TableName() string {
	return "media"
}

// end of Media struct

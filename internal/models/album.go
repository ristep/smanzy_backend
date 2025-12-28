package models

import "gorm.io/gorm"

// MediaAlbum represents a collection/album of media files
// It allows users to organize their media into named collections
type Album struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`

	// Foreign Keys
	// UserID links this album to the user who created it
	UserID uint `json:"user_id"`

	// Creator is the actual User struct that corresponds to UserID.
	// gorm:"foreignKey:UserID" tells GORM how to link the two.
	// json:"-" prevents endless recursion or exposing too much user info.
	Creator User `gorm:"foreignKey:UserID" json:"-"`

	// MediaFiles represents the many-to-many relationship with Media
	// An album can contain multiple media files, and a media file can belong to multiple albums
	// "many2many:album_media" tells GORM to create a join table named "album_media"
	MediaFiles []Media `gorm:"many2many:album_media;" json:"media_files"`

	CreatedAt int64          `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for MediaAlbum
// yep
func (Album) TableName() string {
	return "album"
}

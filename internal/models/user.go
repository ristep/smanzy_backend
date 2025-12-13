package models

import "gorm.io/gorm"

// User represents a user in the system
type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Email         string         `gorm:"unique;not null" json:"email"`
	Password      string         `gorm:"not null" json:"-"` // Never expose password in JSON
	Name          string         `gorm:"not null" json:"name"`
	Tel           string         `json:"tel"`
	Age           int            `json:"age"`
	Address       string         `json:"address"`
	City          string         `json:"city"`
	Country       string         `json:"country"`
	Gender        string         `json:"gender"`
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`
	Roles         []Role         `gorm:"many2many:user_roles;" json:"roles"`
	CreatedAt     int64          `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt     int64          `gorm:"autoUpdateTime:milli" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Role represents a role in the system
type Role struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"unique;not null" json:"name"`
	Users     []User `gorm:"many2many:user_roles;" json:"users,omitempty"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null"`
	Username string `gorm:"uniqueIndex"`
	Name     string

	Password  string
	AvatarURL string
	Bio       string
	IsAdmin   bool `gorm:"default:false"`

	Tools []Tool `gorm:"foreignKey:UserID"`
	Lists []List `gorm:"foreignKey:UserID"`
}

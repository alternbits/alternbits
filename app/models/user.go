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

	GitHubID    string  `gorm:"uniqueIndex"`
	GitHubLogin string  `gorm:"index"`
	GoogleID    *string `gorm:"uniqueIndex"`

	TOTPSecret    string
	TOTPEnabled   bool `gorm:"default:false"`
	RecoveryCodes string

	AIs   []AI   `gorm:"foreignKey:UserID"`
	Lists []List `gorm:"foreignKey:UserID"`
}

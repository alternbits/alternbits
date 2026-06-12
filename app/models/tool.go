package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Tool struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Subtitle    string
	Description string
	LogoURL     string

	Categories []Category `gorm:"many2many:tool_categories;"`
	Genera     []Genus    `gorm:"many2many:tool_genera;"`

	Data datatypes.JSON `gorm:"type:jsonb"`

	UserID *uint
	User   *User
}

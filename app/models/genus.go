package models

import "gorm.io/gorm"

type Genus struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Slug     string `gorm:"uniqueIndex;not null"`
	Subtitle string
}

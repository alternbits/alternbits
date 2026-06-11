package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Slug     string `gorm:"uniqueIndex;not null"`
	Subtitle string

	ParentID *uint
	Parent   *Category  `gorm:"foreignKey:ParentID"`
	Children []Category `gorm:"foreignKey:ParentID"`
}

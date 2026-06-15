package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AIStatus string

const (
	AIStatusPending  AIStatus = "pending"
	AIStatusApproved AIStatus = "approved"
	AIStatusRejected AIStatus = "rejected"
)

type AI struct {
	gorm.Model
	Name        string   `gorm:"not null"`
	Slug        string   `gorm:"uniqueIndex;not null"`
	Status      AIStatus `gorm:"type:text;not null;default:'approved'"`
	Subtitle    string
	Description string
	LogoURL     string
	Website     string
	Screenshots datatypes.JSON `gorm:"type:jsonb"`
	HaveFree    bool           `gorm:"default:false"`

	Categories []Category `gorm:"many2many:ai_categories;joinForeignKey:ai_id;joinReferences:category_id"`
	Genera     []Genus    `gorm:"many2many:ai_genera;joinForeignKey:ai_id;joinReferences:genus_id"`

	Data datatypes.JSON `gorm:"type:jsonb"`

	ArtifactID *uint
	Artifact   *Artifact

	UserID *uint
	User   *User
}

func (AI) TableName() string { return "ais" }

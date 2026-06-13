package models

import "gorm.io/gorm"

type ArtifactFieldType string

const (
	ArtifactFieldTypeString  ArtifactFieldType = "string"
	ArtifactFieldTypeText    ArtifactFieldType = "text"
	ArtifactFieldTypeNumber  ArtifactFieldType = "number"
	ArtifactFieldTypeInteger ArtifactFieldType = "integer"
	ArtifactFieldTypeBoolean ArtifactFieldType = "boolean"
	ArtifactFieldTypeURL     ArtifactFieldType = "url"
	ArtifactFieldTypeEmail   ArtifactFieldType = "email"
	ArtifactFieldTypeDate    ArtifactFieldType = "date"
)

type Artifact struct {
	gorm.Model
	Name string `gorm:"not null"`
	Slug string `gorm:"uniqueIndex;not null"`

	Fields []ArtifactField `gorm:"foreignKey:ArtifactID"`
}

type ArtifactField struct {
	gorm.Model
	ArtifactID   uint              `gorm:"not null;index"`
	Name         string            `gorm:"not null"`
	Slug         string            `gorm:"not null"`
	FieldType    ArtifactFieldType `gorm:"not null;default:'string'"`
	Required     bool              `gorm:"not null;default:false"`
	DefaultValue *string
}

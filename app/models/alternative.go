package models

import "gorm.io/gorm"

type AlternativeStatus string

const (
	AlternativeStatusPending  AlternativeStatus = "pending"
	AlternativeStatusApproved AlternativeStatus = "approved"
	AlternativeStatusRejected AlternativeStatus = "rejected"
)

// Alternative records that one AI is a suggested alternative to another.
// The relationship is directional: AlternativeAI is suggested as an
// alternative to AI. At query time treat both directions as equivalent.
type Alternative struct {
	gorm.Model

	AIID uint `gorm:"not null;uniqueIndex:idx_alternative_pair"`
	AI   *AI

	AlternativeAIID uint `gorm:"not null;uniqueIndex:idx_alternative_pair"`
	AlternativeAI   *AI  `gorm:"foreignKey:AlternativeAIID"`

	Status AlternativeStatus `gorm:"type:text;not null;default:'pending'"`
	Note   string            // optional note from the suggester

	SuggestedByUserID *uint // nil = created directly by admin
	SuggestedBy       *User `gorm:"foreignKey:SuggestedByUserID"`

	ApprovedByUserID *uint // set when status moves to approved or rejected
	ApprovedBy       *User `gorm:"foreignKey:ApprovedByUserID"`
}

package models

import "gorm.io/gorm"

type SavedAI struct {
	gorm.Model
	UserID uint `gorm:"not null;uniqueIndex:idx_saved_ai_pair"`
	AIID   uint `gorm:"not null;uniqueIndex:idx_saved_ai_pair;column:ai_id"`

	User *User
	AI   *AI
}

func (SavedAI) TableName() string { return "saved_ais" }

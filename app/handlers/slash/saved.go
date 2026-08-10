package slash

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// savedIDSet returns the set of AI IDs the given user has saved. Safe to call
// with a nil user (returns an empty, non-nil map).
func savedIDSet(db *gorm.DB, user *models.User) map[uint]bool {
	set := map[uint]bool{}
	if user == nil {
		return set
	}
	var rows []models.SavedAI
	db.Where(&models.SavedAI{UserID: user.ID}).Find(&rows)
	for _, r := range rows {
		set[r.AIID] = true
	}
	return set
}

// SavedToggle is the AJAX endpoint the save button on AI cards/detail pages
// posts to. Requires auth. Toggles a SavedAI row for the current user and
// the posted ai_id, returning the resulting state as JSON.
func SavedToggle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)

		aiID, err := strconv.ParseUint(c.PostForm("ai_id"), 10, 64)
		if err != nil || aiID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ai_id"})
			return
		}

		var existing models.SavedAI
		if err := db.Where(&models.SavedAI{UserID: user.ID, AIID: uint(aiID)}).First(&existing).Error; err == nil {
			db.Delete(&existing)
			c.JSON(http.StatusOK, gin.H{"saved": false})
			return
		}

		var ai models.AI
		if err := db.Select("id").First(&ai, aiID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ai not found"})
			return
		}

		if err := db.Create(&models.SavedAI{UserID: user.ID, AIID: uint(aiID)}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"saved": true})
	}
}

// SavedPage lists the current user's saved AIs at /saved. Requires auth.
func SavedPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		data := headerData(db, user)

		var rows []models.SavedAI
		db.Preload("AI").Preload("AI.Categories").Preload("AI.Genera").
			Where(&models.SavedAI{UserID: user.ID}).
			Order("created_at DESC").
			Find(&rows)

		ais := make([]models.AI, 0, len(rows))
		saved := make(map[uint]bool, len(rows))
		for _, r := range rows {
			if r.AI == nil {
				continue
			}
			ais = append(ais, *r.AI)
			saved[r.AI.ID] = true
		}

		data["AIs"] = ais
		data["SavedIDs"] = saved
		c.HTML(http.StatusOK, "saved.tmpl", data)
	}
}

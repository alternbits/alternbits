package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AlternativesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var ai models.AI
		if err := db.
			Preload("Categories").
			Preload("Genera").
			Where("slug = ?", slug).
			First(&ai).Error; err != nil {
			c.HTML(http.StatusNotFound, "alternatives.tmpl", gin.H{"Error": "AI not found."})
			return
		}

		// Approved alternatives in both directions.
		var rels []models.Alternative
		db.Preload("AI").Preload("AI.Categories").Preload("AI.Genera").
			Preload("AlternativeAI").Preload("AlternativeAI.Categories").Preload("AlternativeAI.Genera").
			Where("(ai_id = ? OR alternative_ai_id = ?) AND status = ?",
				ai.ID, ai.ID, models.AlternativeStatusApproved).
			Order("created_at DESC").
			Find(&rels)

		seen := map[uint]bool{ai.ID: true}
		alts := make([]models.AI, 0, len(rels))
		for _, r := range rels {
			var other *models.AI
			if r.AIID == ai.ID {
				other = r.AlternativeAI
			} else {
				other = r.AI
			}
			if other == nil || seen[other.ID] {
				continue
			}
			seen[other.ID] = true
			alts = append(alts, *other)
		}

		var categories []models.Category
		db.Where("parent_id IS NULL").Find(&categories)

		var aiCount int64
		db.Model(&models.AI{}).Count(&aiCount)

		var currentUser *models.User
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				currentUser = &u
			}
		}

		c.HTML(http.StatusOK, "alternatives.tmpl", gin.H{
			"AI":           ai,
			"Alternatives": alts,
			"Categories":   categories,
			"AICount":      aiCount,
			"CurrentUser":  currentUser,
			"SavedIDs":     savedIDSet(db, currentUser),
		})
	}
}

package slash

import (
	"encoding/json"
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type screenshotItem struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

func AIHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var ai models.AI
		if err := db.
			Preload("Categories").
			Preload("Genera").
			Where("slug = ?", slug).
			First(&ai).Error; err != nil {
			c.HTML(http.StatusNotFound, "ai.tmpl", gin.H{
				"Error":       "AI not found.",
				"CurrentUser": nil,
			})
			return
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

		var screenshots []screenshotItem
		if len(ai.Screenshots) > 0 {
			_ = json.Unmarshal(ai.Screenshots, &screenshots)
		}

		var altRels []models.Alternative
		db.Preload("AI").Preload("AlternativeAI").
			Where("(ai_id = ? OR alternative_ai_id = ?) AND status = ?",
				ai.ID, ai.ID, models.AlternativeStatusApproved).
			Order("created_at DESC").
			Find(&altRels)

		seen := map[uint]bool{ai.ID: true}
		alternatives := make([]models.AI, 0, len(altRels))
		for _, r := range altRels {
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
			alternatives = append(alternatives, *other)
		}

		alternativesPreview := alternatives
		if len(alternativesPreview) > 2 {
			alternativesPreview = alternativesPreview[:2]
		}

		c.HTML(http.StatusOK, "ai.tmpl", gin.H{
			"AI":                  ai,
			"Screenshots":         screenshots,
			"Categories":          categories,
			"AICount":             aiCount,
			"CurrentUser":         currentUser,
			"AlternativesPreview": alternativesPreview,
			"AlternativesCount":   len(alternatives),
			"SavedIDs":            savedIDSet(db, currentUser),
		})
	}
}

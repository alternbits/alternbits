package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

		c.HTML(http.StatusOK, "ai.tmpl", gin.H{
			"AI":          ai,
			"Categories":  categories,
			"AICount":     aiCount,
			"CurrentUser": currentUser,
		})
	}
}

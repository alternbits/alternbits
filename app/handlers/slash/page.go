package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PageHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var pg models.Page
		if err := db.Where("slug = ?", slug).First(&pg).Error; err != nil {
			c.HTML(http.StatusNotFound, "page.tmpl", gin.H{
				"Error":       "Page not found.",
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

		c.HTML(http.StatusOK, "page.tmpl", gin.H{
			"Page":        pg,
			"Categories":  categories,
			"AICount":     aiCount,
			"CurrentUser": currentUser,
		})
	}
}

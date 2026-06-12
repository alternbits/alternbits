package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ToolHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var tool models.Tool
		if err := db.
			Preload("Categories").
			Preload("Genera").
			Where("slug = ?", slug).
			First(&tool).Error; err != nil {
			c.HTML(http.StatusNotFound, "tool.tmpl", gin.H{
				"Error":       "Tool not found.",
				"CurrentUser": nil,
			})
			return
		}

		var categories []models.Category
		db.Where("parent_id IS NULL").Find(&categories)

		var toolCount int64
		db.Model(&models.Tool{}).Count(&toolCount)

		var currentUser *models.User
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				currentUser = &u
			}
		}

		c.HTML(http.StatusOK, "tool.tmpl", gin.H{
			"Tool":        tool,
			"Categories":  categories,
			"ToolCount":   toolCount,
			"CurrentUser": currentUser,
		})
	}
}

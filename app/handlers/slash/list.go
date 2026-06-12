package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lists []models.List
		db.Preload("Items").Order("created_at desc").Find(&lists)

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

		c.HTML(http.StatusOK, "root_lists.tmpl", gin.H{
			"Lists":       lists,
			"Categories":  categories,
			"ToolCount":   toolCount,
			"CurrentUser": currentUser,
		})
	}
}

func ListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var list models.List
		if err := db.
			Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("sort ASC") }).
			Preload("Items.Tool").
			Preload("Items.Tool.Categories").
			Where("slug = ?", slug).
			First(&list).Error; err != nil {
			c.HTML(http.StatusNotFound, "list.tmpl", gin.H{"Error": "List not found."})
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

		c.HTML(http.StatusOK, "list.tmpl", gin.H{
			"List":        list,
			"Categories":  categories,
			"ToolCount":   toolCount,
			"CurrentUser": currentUser,
		})
	}
}

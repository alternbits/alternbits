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
		var currentUser *models.User
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				currentUser = &u
			}
		}

		var lists []models.List
		q := db.Preload("Items").Order("created_at desc")
		if currentUser != nil {
			q = q.Where("is_private = false OR user_id = ?", currentUser.ID)
		} else {
			q = q.Where("is_private = false")
		}
		q.Find(&lists)

		var categories []models.Category
		db.Where("parent_id IS NULL").Find(&categories)

		var aiCount int64
		db.Model(&models.AI{}).Count(&aiCount)

		c.HTML(http.StatusOK, "lists.tmpl", gin.H{
			"Lists":       lists,
			"Categories":  categories,
			"AICount":     aiCount,
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
			Preload("Items.AI").
			Preload("Items.AI.Categories").
			Where("slug = ?", slug).
			First(&list).Error; err != nil {
			c.HTML(http.StatusNotFound, "list.tmpl", gin.H{"Error": "List not found."})
			return
		}

		var currentUser *models.User
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				currentUser = &u
			}
		}

		if list.IsPrivate {
			isOwner := currentUser != nil && list.UserID != nil && *list.UserID == currentUser.ID
			if !isOwner {
				c.HTML(http.StatusNotFound, "list.tmpl", gin.H{"Error": "List not found."})
				return
			}
		}

		var categories []models.Category
		db.Where("parent_id IS NULL").Find(&categories)

		var aiCount int64
		db.Model(&models.AI{}).Count(&aiCount)

		c.HTML(http.StatusOK, "list.tmpl", gin.H{
			"List":        list,
			"Categories":  categories,
			"AICount":     aiCount,
			"CurrentUser": currentUser,
		})
	}
}

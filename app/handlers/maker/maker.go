package maker

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionUserIDKey = "user_id"

func Handler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		c.HTML(http.StatusOK, "maker_index.tmpl", gin.H{
			"Categories":  categories,
			"AICount":     aiCount,
			"CurrentUser": currentUser,
		})
	}
}

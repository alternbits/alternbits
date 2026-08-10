package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pageData struct {
	AIs         []models.AI
	Categories  []models.Category
	Lists       []models.List
	AICount     int64
	CurrentUser *models.User
	SavedIDs    map[uint]bool
}

func Handler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data pageData

		db.Preload("Categories").Order("created_at desc").Limit(12).Find(&data.AIs)
		db.Where("parent_id IS NULL").Find(&data.Categories)
		db.Order("created_at desc").Limit(6).Find(&data.Lists)
		db.Model(&models.AI{}).Count(&data.AICount)

		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				data.CurrentUser = &u
			}
		}
		data.SavedIDs = savedIDSet(db, data.CurrentUser)

		c.HTML(http.StatusOK, "slash.tmpl", data)
	}
}

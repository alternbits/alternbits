package maker

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionUserIDKey = "user_id"

func headerData(db *gorm.DB, user *models.User) gin.H {
	var categories []models.Category
	db.Where("parent_id IS NULL").Find(&categories)
	var aiCount int64
	db.Model(&models.AI{}).Count(&aiCount)
	return gin.H{
		"CurrentUser": user,
		"Categories":  categories,
		"AICount":     aiCount,
	}
}

func sessionUser(c *gin.Context, db *gorm.DB) *models.User {
	session := sessions.Default(c)
	uid, ok := session.Get(sessionUserIDKey).(uint)
	if !ok || uid == 0 {
		return nil
	}
	var u models.User
	if db.First(&u, uid).Error != nil {
		return nil
	}
	return &u
}

func Handler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := sessionUser(c, db)
		data := headerData(db, user)
		c.HTML(http.StatusOK, "maker_index.tmpl", data)
	}
}

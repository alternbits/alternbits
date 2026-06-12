package root

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DashboardHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users, ais, categories, admins, uncategorized, lists int64
		db.Model(&models.User{}).Count(&users)
		db.Model(&models.AI{}).Count(&ais)
		db.Model(&models.Category{}).Count(&categories)
		db.Model(&models.List{}).Count(&lists)
		db.Model(&models.User{}).Where(&models.User{IsAdmin: true}).Count(&admins)
		db.Model(&models.AI{}).
			Where("id NOT IN (SELECT ai_id FROM ai_categories)").
			Count(&uncategorized)

		var recentAIs []models.AI
		db.Preload("Categories").
			Order("created_at DESC").
			Limit(5).
			Find(&recentAIs)

		var recentUsers []models.User
		db.Order("created_at DESC").
			Limit(5).
			Find(&recentUsers)

		c.HTML(http.StatusOK, "root_dashboard.tmpl", gin.H{
			"ActiveNav":     "dashboard",
			"Users":         users,
			"AIs":           ais,
			"Categories":    categories,
			"Lists":         lists,
			"Admins":        admins,
			"Uncategorized": uncategorized,
			"RecentAIs":     recentAIs,
			"RecentUsers":   recentUsers,
		})
	}
}

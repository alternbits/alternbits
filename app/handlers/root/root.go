package root

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DashboardHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users, tools, categories, admins, uncategorized int64
		db.Model(&models.User{}).Count(&users)
		db.Model(&models.Tool{}).Count(&tools)
		db.Model(&models.Category{}).Count(&categories)
		db.Model(&models.User{}).Where(&models.User{IsAdmin: true}).Count(&admins)
		db.Model(&models.Tool{}).
			Where("id NOT IN (SELECT tool_id FROM tool_categories)").
			Count(&uncategorized)

		var recentTools []models.Tool
		db.Preload("Categories").
			Order("created_at DESC").
			Limit(5).
			Find(&recentTools)

		var recentUsers []models.User
		db.Order("created_at DESC").
			Limit(5).
			Find(&recentUsers)

		c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
			"ActiveNav":     "dashboard",
			"Users":         users,
			"Tools":         tools,
			"Categories":    categories,
			"Admins":        admins,
			"Uncategorized": uncategorized,
			"RecentTools":   recentTools,
			"RecentUsers":   recentUsers,
		})
	}
}

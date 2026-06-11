package root

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DashboardHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users, tools, categories int64
		db.Model(&models.User{}).Count(&users)
		db.Model(&models.Tool{}).Count(&tools)
		db.Model(&models.Category{}).Count(&categories)

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Users":      users,
			"Tools":      tools,
			"Categories": categories,
		})
	}
}

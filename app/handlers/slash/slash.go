package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pageData struct {
	Tools      []models.Tool
	Categories []models.Category
	Lists      []models.List
	ToolCount  int64
}

func Handler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data pageData

		db.Preload("Categories").Order("created_at desc").Limit(12).Find(&data.Tools)
		db.Where("parent_id IS NULL").Find(&data.Categories)
		db.Order("created_at desc").Limit(6).Find(&data.Lists)
		db.Model(&models.Tool{}).Count(&data.ToolCount)

		c.HTML(http.StatusOK, "slash.tmpl", data)
	}
}

package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CategoriesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var topLevel []models.Category
		db.Preload("Children").Where("parent_id IS NULL").Order("name ASC").Find(&topLevel)

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

		c.HTML(http.StatusOK, "categories.tmpl", gin.H{
			"TopLevel":    topLevel,
			"Categories":  topLevel,
			"ToolCount":   toolCount,
			"CurrentUser": currentUser,
		})
	}
}

func CategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var category models.Category
		if err := db.
			Preload("Children").
			Preload("Parent").
			Where("slug = ?", slug).
			First(&category).Error; err != nil {
			c.HTML(http.StatusNotFound, "category.tmpl", gin.H{"Error": "Category not found."})
			return
		}

		// Collect this category's ID plus all child IDs to query tools.
		catIDs := []uint{category.ID}
		for _, child := range category.Children {
			catIDs = append(catIDs, child.ID)
		}

		var tools []models.Tool
		db.Preload("Categories").Preload("Genera").
			Joins("JOIN tool_categories tc ON tc.tool_id = tools.id").
			Where("tc.category_id IN ?", catIDs).
			Order("tools.name ASC").
			Find(&tools)

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

		c.HTML(http.StatusOK, "category.tmpl", gin.H{
			"Category":    category,
			"Tools":       tools,
			"Categories":  categories,
			"ToolCount":   toolCount,
			"CurrentUser": currentUser,
		})
	}
}

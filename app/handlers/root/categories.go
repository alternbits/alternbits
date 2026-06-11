package root

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const categoriesPerPage = 20

type categoriesPage struct {
	Categories []models.Category
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

func CategoriesListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		var total int64
		if err := db.Model(&models.Category{}).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "categories.html", gin.H{"Error": "Failed to count categories"})
			return
		}

		totalPages := max(int((total+categoriesPerPage-1)/categoriesPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var categories []models.Category
		if err := db.
			Preload("Parent").
			Order("created_at DESC").
			Offset((page - 1) * categoriesPerPage).
			Limit(categoriesPerPage).
			Find(&categories).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "categories.html", gin.H{"Error": "Failed to load categories"})
			return
		}

		c.HTML(http.StatusOK, "categories.html", gin.H{
			"ActiveNav": "categories",
			"Page": categoriesPage{
				Categories: categories,
				Page:       page,
				TotalPages: totalPages,
				Total:      total,
				PrevPage:   page - 1,
				NextPage:   page + 1,
				HasPrev:    page > 1,
				HasNext:    page < totalPages,
			},
		})
	}
}

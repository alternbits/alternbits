package root

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

type categoriesFilters struct {
	Search   string
	Level    string // "top" | "sub" | ""
	Sort     string
	QueryStr string
}

func CategoriesListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		search := strings.TrimSpace(c.Query("q"))
		level := c.Query("level")
		sort := c.Query("sort")

		base := db.Model(&models.Category{})
		if search != "" {
			like := "%" + search + "%"
			base = base.Where("name ILIKE ? OR slug ILIKE ? OR subtitle ILIKE ?", like, like, like)
		}
		switch level {
		case "top":
			base = base.Where("parent_id IS NULL")
		case "sub":
			base = base.Where("parent_id IS NOT NULL")
		}

		orderClause := "created_at DESC"
		switch sort {
		case "oldest":
			orderClause = "created_at ASC"
		case "name_asc":
			orderClause = "name ASC"
		case "name_desc":
			orderClause = "name DESC"
		}

		var total int64
		if err := base.Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_categories.tmpl", gin.H{"Error": "Failed to count categories"})
			return
		}

		totalPages := max(int((total+categoriesPerPage-1)/categoriesPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var categories []models.Category
		if err := base.
			Preload("Parent").
			Order(orderClause).
			Offset((page - 1) * categoriesPerPage).
			Limit(categoriesPerPage).
			Find(&categories).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_categories.tmpl", gin.H{"Error": "Failed to load categories"})
			return
		}

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if level != "" {
			params.Set("level", level)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		c.HTML(http.StatusOK, "root_categories.tmpl", gin.H{
			"ActiveNav": "categories",
			"Filters": categoriesFilters{
				Search:   search,
				Level:    level,
				Sort:     sort,
				QueryStr: queryStr,
			},
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

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

const pagesPerPage = 20

type pagesPage struct {
	Pages      []models.Page
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type pagesFilters struct {
	Search   string
	Sort     string
	QueryStr string
}

func PagesListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		search := strings.TrimSpace(c.Query("q"))
		sort := c.Query("sort")

		base := db.Model(&models.Page{})
		if search != "" {
			like := "%" + search + "%"
			base = base.Where("title ILIKE ? OR slug ILIKE ?", like, like)
		}

		orderClause := "created_at DESC"
		switch sort {
		case "oldest":
			orderClause = "created_at ASC"
		case "name_asc":
			orderClause = "title ASC"
		case "name_desc":
			orderClause = "title DESC"
		}

		var total int64
		if err := base.Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_pages.tmpl", gin.H{"Error": "Failed to count pages"})
			return
		}

		totalPages := max(int((total+pagesPerPage-1)/pagesPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var pages []models.Page
		if err := base.
			Order(orderClause).
			Offset((page - 1) * pagesPerPage).
			Limit(pagesPerPage).
			Find(&pages).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_pages.tmpl", gin.H{"Error": "Failed to load pages"})
			return
		}

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		c.HTML(http.StatusOK, "root_pages.tmpl", gin.H{
			"ActiveNav": "pages",
			"Filters": pagesFilters{
				Search:   search,
				Sort:     sort,
				QueryStr: queryStr,
			},
			"Page": pagesPage{
				Pages:      pages,
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

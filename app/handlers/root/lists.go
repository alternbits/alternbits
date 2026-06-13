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

const listsPerPage = 20

type listRow struct {
	models.List
	ItemCount int
}

type listsPage struct {
	Lists      []listRow
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type listsFilters struct {
	Search     string
	Source     string // "auto" | "manual" | ""
	Visibility string // "public" | "private" | ""
	Sort       string
	QueryStr   string
}

func ListsListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		search := strings.TrimSpace(c.Query("q"))
		source := c.Query("source")
		visibility := c.Query("visibility")
		sort := c.Query("sort")

		base := db.Table("lists").Where("lists.deleted_at IS NULL")
		if search != "" {
			base = base.Where("lists.name ILIKE ? OR lists.slug ILIKE ?", "%"+search+"%", "%"+search+"%")
		}
		switch source {
		case "auto":
			base = base.Where("lists.is_auto_generated = ?", true)
		case "manual":
			base = base.Where("lists.is_auto_generated = ?", false)
		}
		switch visibility {
		case "public":
			base = base.Where("lists.is_private = ?", false)
		case "private":
			base = base.Where("lists.is_private = ?", true)
		}

		orderClause := "lists.created_at DESC"
		switch sort {
		case "oldest":
			orderClause = "lists.created_at ASC"
		case "name_asc":
			orderClause = "lists.name ASC"
		case "name_desc":
			orderClause = "lists.name DESC"
		case "most_ais":
			orderClause = "item_count DESC"
		case "fewest_ais":
			orderClause = "item_count ASC"
		}

		var total int64
		if err := base.Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_lists.tmpl", gin.H{"Error": "Failed to count lists"})
			return
		}

		totalPages := max(int((total+listsPerPage-1)/listsPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var lists []listRow
		if err := base.
			Select("lists.*, (SELECT COUNT(*) FROM list_ais WHERE list_ais.list_id = lists.id) AS item_count").
			Preload("User").
			Order(orderClause).
			Offset((page - 1) * listsPerPage).
			Limit(listsPerPage).
			Find(&lists).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_lists.tmpl", gin.H{"Error": "Failed to load lists"})
			return
		}

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if source != "" {
			params.Set("source", source)
		}
		if visibility != "" {
			params.Set("visibility", visibility)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		c.HTML(http.StatusOK, "root_lists.tmpl", gin.H{
			"ActiveNav": "lists",
			"Filters": listsFilters{
				Search:     search,
				Source:     source,
				Visibility: visibility,
				Sort:       sort,
				QueryStr:   queryStr,
			},
			"Page": listsPage{
				Lists:      lists,
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

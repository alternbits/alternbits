package root

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const aisPerPage = 20

type aisPage struct {
	AIs        []models.AI
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type aisFilters struct {
	Search   string
	Category string
	Genus    string
	Sort     string
	QueryStr string // filter params without page= for pagination links
}

func filteredAIs(db *gorm.DB, search, category, genus string) *gorm.DB {
	q := db.Model(&models.AI{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("ais.name ILIKE ? OR ais.slug ILIKE ? OR ais.subtitle ILIKE ?", like, like, like)
	}
	if category != "" {
		q = q.
			Joins("JOIN ai_categories ON ai_categories.ai_id = ais.id").
			Joins("JOIN categories ON categories.id = ai_categories.category_id").
			Where("categories.slug = ?", category)
	}
	if genus != "" {
		q = q.
			Joins("JOIN ai_genera ON ai_genera.ai_id = ais.id").
			Joins("JOIN genera ON genera.id = ai_genera.genus_id").
			Where("genera.slug = ?", genus)
	}
	return q
}

func AIsListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		search := c.Query("q")
		category := c.Query("category")
		genus := c.Query("genus")
		sort := c.Query("sort")

		orderClause := "ais.created_at DESC"
		switch sort {
		case "oldest":
			orderClause = "ais.created_at ASC"
		case "name_asc":
			orderClause = "ais.name ASC"
		case "name_desc":
			orderClause = "ais.name DESC"
		}

		var total int64
		if err := filteredAIs(db, search, category, genus).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_ais.tmpl", gin.H{"Error": "Failed to count AIs"})
			return
		}

		totalPages := max(int((total+int64(aisPerPage)-1)/int64(aisPerPage)), 1)
		if page > totalPages {
			page = totalPages
		}

		// Get matching IDs in sorted/paged order, then fetch with preloads to avoid JOIN duplicates
		var ids []uint
		if err := filteredAIs(db, search, category, genus).
			Select("ais.id").
			Order(orderClause).
			Offset((page-1)*aisPerPage).
			Limit(aisPerPage).
			Pluck("ais.id", &ids).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_ais.tmpl", gin.H{"Error": "Failed to load AIs"})
			return
		}

		var ais []models.AI
		if len(ids) > 0 {
			if err := db.
				Preload("Categories").
				Preload("Genera").
				Preload("User").
				Where("id IN ?", ids).
				Order(orderClause).
				Find(&ais).Error; err != nil {
				c.HTML(http.StatusInternalServerError, "root_ais.tmpl", gin.H{"Error": "Failed to load AIs"})
				return
			}
		}

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if category != "" {
			params.Set("category", category)
		}
		if genus != "" {
			params.Set("genus", genus)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		var categories []models.Category
		db.Order("name ASC").Find(&categories)

		var genera []models.Genus
		db.Order("name ASC").Find(&genera)

		c.HTML(http.StatusOK, "root_ais.tmpl", gin.H{
			"ActiveNav":  "ais",
			"Categories": categories,
			"Genera":     genera,
			"Filters": aisFilters{
				Search:   search,
				Category: category,
				Genus:    genus,
				Sort:     sort,
				QueryStr: queryStr,
			},
			"Page": aisPage{
				AIs:        ais,
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

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

const artifactsPerPage = 20

type artifactsPage struct {
	Artifacts  []models.Artifact
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type artifactsFilters struct {
	Search   string
	Sort     string
	QueryStr string
}

func ArtifactsListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		search := strings.TrimSpace(c.Query("q"))
		sort := c.Query("sort")

		base := db.Model(&models.Artifact{})
		if search != "" {
			like := "%" + search + "%"
			base = base.Where("name ILIKE ? OR slug ILIKE ?", like, like)
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
			c.HTML(http.StatusInternalServerError, "root_artifacts.tmpl", gin.H{"Error": "Failed to count artifacts"})
			return
		}

		totalPages := max(int((total+artifactsPerPage-1)/artifactsPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var artifacts []models.Artifact
		if err := base.
			Preload("Fields").
			Order(orderClause).
			Offset((page - 1) * artifactsPerPage).
			Limit(artifactsPerPage).
			Find(&artifacts).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_artifacts.tmpl", gin.H{"Error": "Failed to load artifacts"})
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

		c.HTML(http.StatusOK, "root_artifacts.tmpl", gin.H{
			"ActiveNav": "artifacts",
			"Filters": artifactsFilters{
				Search:   search,
				Sort:     sort,
				QueryStr: queryStr,
			},
			"Page": artifactsPage{
				Artifacts:  artifacts,
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

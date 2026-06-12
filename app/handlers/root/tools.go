package root

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const toolsPerPage = 20

type toolsPage struct {
	Tools      []models.Tool
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

func ToolsListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		var total int64
		if err := db.Model(&models.Tool{}).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "tools.tmpl", gin.H{"Error": "Failed to count tools"})
			return
		}

		totalPages := max(int((total+toolsPerPage-1)/toolsPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var tools []models.Tool
		if err := db.
			Preload("Categories").
			Preload("User").
			Order("created_at DESC").
			Offset((page - 1) * toolsPerPage).
			Limit(toolsPerPage).
			Find(&tools).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "tools.tmpl", gin.H{"Error": "Failed to load tools"})
			return
		}

		c.HTML(http.StatusOK, "tools.tmpl", gin.H{
			"ActiveNav": "tools",
			"Page": toolsPage{
				Tools:      tools,
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

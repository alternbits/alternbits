package root

import (
	"net/http"
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

func AIsListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		var total int64
		if err := db.Model(&models.AI{}).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_ais.tmpl", gin.H{"Error": "Failed to count AIs"})
			return
		}

		totalPages := max(int((total+aisPerPage-1)/aisPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var ais []models.AI
		if err := db.
			Preload("Categories").
			Preload("User").
			Order("created_at DESC").
			Offset((page - 1) * aisPerPage).
			Limit(aisPerPage).
			Find(&ais).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_ais.tmpl", gin.H{"Error": "Failed to load AIs"})
			return
		}

		c.HTML(http.StatusOK, "root_ais.tmpl", gin.H{
			"ActiveNav": "ais",
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

package root

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const generaPerPage = 20

type generaPage struct {
	Genera     []models.Genus
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

func GeneraListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		var total int64
		if err := db.Model(&models.Genus{}).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "genera.tmpl", gin.H{"Error": "Failed to count genera"})
			return
		}

		totalPages := max(int((total+generaPerPage-1)/generaPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var genera []models.Genus
		if err := db.
			Order("created_at DESC").
			Offset((page - 1) * generaPerPage).
			Limit(generaPerPage).
			Find(&genera).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "genera.tmpl", gin.H{"Error": "Failed to load genera"})
			return
		}

		c.HTML(http.StatusOK, "genera.tmpl", gin.H{
			"ActiveNav": "genera",
			"Page": generaPage{
				Genera:     genera,
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

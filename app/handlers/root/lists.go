package root

import (
	"net/http"
	"strconv"

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

func ListsListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		var total int64
		if err := db.Model(&models.List{}).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "lists.tmpl", gin.H{"Error": "Failed to count lists"})
			return
		}

		totalPages := max(int((total+listsPerPage-1)/listsPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var lists []listRow
		if err := db.
			Table("lists").
			Select("lists.*, (SELECT COUNT(*) FROM list_tools WHERE list_tools.list_id = lists.id) AS item_count").
			Preload("User").
			Where("lists.deleted_at IS NULL").
			Order("lists.created_at DESC").
			Offset((page - 1) * listsPerPage).
			Limit(listsPerPage).
			Find(&lists).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "lists.tmpl", gin.H{"Error": "Failed to load lists"})
			return
		}

		c.HTML(http.StatusOK, "lists.tmpl", gin.H{
			"ActiveNav": "lists",
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

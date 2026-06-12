package root

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const usersPerPage = 20

type usersPage struct {
	Users      []models.User
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

func UsersListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		var total int64
		if err := db.Model(&models.User{}).Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "users.tmpl", gin.H{"Error": "Failed to count users"})
			return
		}

		totalPages := max(int((total+usersPerPage-1)/usersPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var users []models.User
		if err := db.
			Order("created_at DESC").
			Offset((page - 1) * usersPerPage).
			Limit(usersPerPage).
			Find(&users).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "users.tmpl", gin.H{"Error": "Failed to load users"})
			return
		}

		c.HTML(http.StatusOK, "users.tmpl", gin.H{
			"ActiveNav": "users",
			"Page": usersPage{
				Users:      users,
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

package root

import (
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UsersSearchAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			c.JSON(http.StatusOK, []any{})
			return
		}

		like := "%" + q + "%"
		var users []models.User
		db.Select("id, name, username, email").
			Where("username ILIKE ? OR name ILIKE ? OR email ILIKE ?", like, like, like).
			Order("username ASC, name ASC").
			Limit(10).
			Find(&users)

		type result struct {
			ID    uint   `json:"id"`
			Label string `json:"label"`
		}
		out := make([]result, len(users))
		for i, u := range users {
			out[i] = result{ID: u.ID, Label: formatOwnerLabel(u)}
		}
		c.JSON(http.StatusOK, out)
	}
}

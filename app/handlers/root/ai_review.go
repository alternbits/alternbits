package root

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AIReviewHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pendingAIs []models.AI
		db.Preload("User").
			Where("status = ?", models.AIStatusPending).
			Order("created_at ASC").
			Find(&pendingAIs)

		c.HTML(http.StatusOK, "root_ais_review.tmpl", gin.H{
			"ActiveNav":  "ais",
			"PendingAIs": pendingAIs,
		})
	}
}

func AIApprove(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/ais/review")
			return
		}
		db.Model(&models.AI{}).Where("id = ?", id).Update("status", models.AIStatusApproved)
		back := c.PostForm("back")
		if back == "" {
			back = "/root/ais/review"
		}
		c.Redirect(http.StatusFound, back)
	}
}

func AIReject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/ais/review")
			return
		}
		db.Model(&models.AI{}).Where("id = ?", id).Update("status", models.AIStatusRejected)
		back := c.PostForm("back")
		if back == "" {
			back = "/root/ais/review"
		}
		c.Redirect(http.StatusFound, back)
	}
}

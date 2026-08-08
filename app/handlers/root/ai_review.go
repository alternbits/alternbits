package root

import (
	"net/http"
	"strconv"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
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

func AIApprove(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/ais/review")
			return
		}
		var ai models.AI
		db.Select("id, name").First(&ai, id)
		db.Model(&models.AI{}).Where("id = ?", id).Update("status", models.AIStatusApproved)
		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyAIApproved(actor, ai.Name)
		back := c.PostForm("back")
		if back == "" {
			back = "/root/ais/review"
		}
		c.Redirect(http.StatusFound, back)
	}
}

func AIReject(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/ais/review")
			return
		}
		var ai models.AI
		db.Select("id, name").First(&ai, id)
		db.Model(&models.AI{}).Where("id = ?", id).Update("status", models.AIStatusRejected)
		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyAIRejected(actor, ai.Name)
		back := c.PostForm("back")
		if back == "" {
			back = "/root/ais/review"
		}
		c.Redirect(http.StatusFound, back)
	}
}

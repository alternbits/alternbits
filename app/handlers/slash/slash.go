package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pageData struct {
	AIs           []models.AI
	TrendingAIs   []models.AI
	Categories    []models.Category
	Lists         []models.List
	Genera        []models.Genus
	SelectedGenus string
	AICount       int64
	CurrentUser   *models.User
	SavedIDs      map[uint]bool
}

func Handler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data pageData

		data.SelectedGenus = c.Query("genus")
		if data.SelectedGenus != "" {
			var ids []uint
			db.Table("ais").
				Joins("JOIN ai_genera ON ai_genera.ai_id = ais.id").
				Joins("JOIN genus ON genus.id = ai_genera.genus_id").
				Where("genus.slug = ?", data.SelectedGenus).
				Order("ais.created_at desc").
				Limit(12).
				Pluck("ais.id", &ids)
			if len(ids) > 0 {
				db.Preload("Categories").Where("id IN ?", ids).Order("created_at desc").Find(&data.AIs)
			}
		} else {
			db.Preload("Categories").Order("created_at desc").Limit(12).Find(&data.AIs)
		}

		db.Where("parent_id IS NULL").Find(&data.Categories)
		db.Order("created_at desc").Limit(6).Find(&data.Lists)
		db.Order("name asc").Find(&data.Genera)
		db.Model(&models.AI{}).Count(&data.AICount)

		data.TrendingAIs = data.AIs
		if len(data.TrendingAIs) > 4 {
			data.TrendingAIs = data.TrendingAIs[:4]
		}

		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				data.CurrentUser = &u
			}
		}
		data.SavedIDs = savedIDSet(db, data.CurrentUser)

		c.HTML(http.StatusOK, "slash.tmpl", data)
	}
}

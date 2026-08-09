package slash

import (
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type searchResult struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Subtitle string `json:"subtitle,omitempty"`
	LogoURL  string `json:"logo_url,omitempty"`
	URL      string `json:"url"`
}

// SearchPage is the full-page search results view at GET /search.
// It renders approved AIs, categories, and public lists matching the query.
func SearchPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))

		var aiCount int64
		db.Model(&models.AI{}).Count(&aiCount)

		var currentUser *models.User
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			var u models.User
			if db.First(&u, uid).Error == nil {
				currentUser = &u
			}
		}

		var topCategories []models.Category
		db.Where("parent_id IS NULL").Order("name ASC").Find(&topCategories)

		data := gin.H{
			"Query":       q,
			"Categories":  topCategories,
			"AICount":     aiCount,
			"CurrentUser": currentUser,
		}

		if q == "" {
			c.HTML(http.StatusOK, "search.tmpl", data)
			return
		}

		like := "%" + q + "%"

		var ais []models.AI
		db.Where(&models.AI{Status: models.AIStatusApproved}).
			Where("name ILIKE ? OR subtitle ILIKE ? OR slug ILIKE ? OR description ILIKE ?", like, like, like, like).
			Preload("Genera").
			Order("name ASC").
			Limit(60).
			Find(&ais)

		var cats []models.Category
		db.Where("name ILIKE ? OR slug ILIKE ? OR subtitle ILIKE ?", like, like, like).
			Order("name ASC").
			Limit(30).
			Find(&cats)

		var lists []models.List
		db.Where(&models.List{IsPrivate: false}).
			Where("name ILIKE ? OR slug ILIKE ? OR subtitle ILIKE ?", like, like, like).
			Order("name ASC").
			Limit(30).
			Find(&lists)

		data["AIs"] = ais
		data["Cats"] = cats
		data["Lists"] = lists
		data["ResultCount"] = len(ais) + len(cats) + len(lists)

		c.HTML(http.StatusOK, "search.tmpl", data)
	}
}

// SearchAPI is the public autocomplete endpoint used by the header search bar.
// It returns approved AIs, categories, and public lists matching the query.
func SearchAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			c.JSON(http.StatusOK, gin.H{
				"ais":        []searchResult{},
				"categories": []searchResult{},
				"lists":      []searchResult{},
			})
			return
		}
		like := "%" + q + "%"

		var ais []models.AI
		db.Where(&models.AI{Status: models.AIStatusApproved}).
			Where("name ILIKE ? OR subtitle ILIKE ? OR slug ILIKE ?", like, like, like).
			Order("name ASC").
			Limit(6).
			Find(&ais)
		aiResults := make([]searchResult, 0, len(ais))
		for _, a := range ais {
			aiResults = append(aiResults, searchResult{
				Name:     a.Name,
				Slug:     a.Slug,
				Subtitle: a.Subtitle,
				LogoURL:  a.LogoURL,
				URL:      "/ai/" + a.Slug,
			})
		}

		var cats []models.Category
		db.Where("name ILIKE ? OR slug ILIKE ?", like, like).
			Order("name ASC").
			Limit(5).
			Find(&cats)
		catResults := make([]searchResult, 0, len(cats))
		for _, cat := range cats {
			catResults = append(catResults, searchResult{
				Name:     cat.Name,
				Slug:     cat.Slug,
				Subtitle: cat.Subtitle,
				URL:      "/categories/" + cat.Slug,
			})
		}

		var lists []models.List
		db.Where(&models.List{IsPrivate: false}).
			Where("name ILIKE ? OR slug ILIKE ?", like, like).
			Order("name ASC").
			Limit(5).
			Find(&lists)
		listResults := make([]searchResult, 0, len(lists))
		for _, l := range lists {
			listResults = append(listResults, searchResult{
				Name:     l.Name,
				Slug:     l.Slug,
				Subtitle: l.Subtitle,
				URL:      "/lists/" + l.Slug,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"ais":        aiResults,
			"categories": catResults,
			"lists":      listResults,
		})
	}
}

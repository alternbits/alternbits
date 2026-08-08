package slash

import (
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
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

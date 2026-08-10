package slash

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ldListItem struct {
	Type     string `json:"@type"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Item     string `json:"item,omitempty"`
	URL      string `json:"url,omitempty"`
}

type ldBreadcrumbList struct {
	Context         string       `json:"@context"`
	Type            string       `json:"@type"`
	ItemListElement []ldListItem `json:"itemListElement"`
}

type ldItemList struct {
	Context         string       `json:"@context"`
	Type            string       `json:"@type"`
	Name            string       `json:"name"`
	ItemListElement []ldListItem `json:"itemListElement"`
}

func marshalLD(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return template.JS(b)
}

func AlternativesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var ai models.AI
		if err := db.
			Preload("Categories").
			Preload("Genera").
			Where("slug = ?", slug).
			First(&ai).Error; err != nil {
			c.HTML(http.StatusNotFound, "alternatives.tmpl", gin.H{"Error": "AI not found."})
			return
		}

		// Approved alternatives in both directions.
		var rels []models.Alternative
		db.Preload("AI").Preload("AI.Categories").Preload("AI.Genera").
			Preload("AlternativeAI").Preload("AlternativeAI.Categories").Preload("AlternativeAI.Genera").
			Where("(ai_id = ? OR alternative_ai_id = ?) AND status = ?",
				ai.ID, ai.ID, models.AlternativeStatusApproved).
			Order("created_at DESC").
			Find(&rels)

		seen := map[uint]bool{ai.ID: true}
		alts := make([]models.AI, 0, len(rels))
		for _, r := range rels {
			var other *models.AI
			if r.AIID == ai.ID {
				other = r.AlternativeAI
			} else {
				other = r.AI
			}
			if other == nil || seen[other.ID] {
				continue
			}
			seen[other.ID] = true
			alts = append(alts, *other)
		}

		var categories []models.Category
		db.Where("parent_id IS NULL").Find(&categories)

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

		domain := config.C.App.Domain
		canonicalURL := fmt.Sprintf("https://%s/alternatives/%s", domain, ai.Slug)

		var breadcrumbJSON, itemListJSON template.JS
		if len(alts) > 0 {
			breadcrumbJSON = marshalLD(ldBreadcrumbList{
				Context: "https://schema.org",
				Type:    "BreadcrumbList",
				ItemListElement: []ldListItem{
					{Type: "ListItem", Position: 1, Name: "Home", Item: fmt.Sprintf("https://%s/", domain)},
					{Type: "ListItem", Position: 2, Name: ai.Name, Item: fmt.Sprintf("https://%s/ai/%s", domain, ai.Slug)},
					{Type: "ListItem", Position: 3, Name: "Alternatives", Item: canonicalURL},
				},
			})

			items := make([]ldListItem, 0, len(alts))
			for i, a := range alts {
				items = append(items, ldListItem{
					Type:     "ListItem",
					Position: i + 1,
					Name:     a.Name,
					URL:      fmt.Sprintf("https://%s/ai/%s", domain, a.Slug),
				})
			}
			itemListJSON = marshalLD(ldItemList{
				Context:         "https://schema.org",
				Type:            "ItemList",
				Name:            fmt.Sprintf("Alternatives to %s", ai.Name),
				ItemListElement: items,
			})
		}

		c.HTML(http.StatusOK, "alternatives.tmpl", gin.H{
			"AI":             ai,
			"Alternatives":   alts,
			"Categories":     categories,
			"AICount":        aiCount,
			"CurrentUser":    currentUser,
			"SavedIDs":       savedIDSet(db, currentUser),
			"CanonicalURL":   canonicalURL,
			"BreadcrumbJSON": breadcrumbJSON,
			"ItemListJSON":   itemListJSON,
		})
	}
}

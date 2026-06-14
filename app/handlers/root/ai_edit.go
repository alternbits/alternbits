package root

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func AIEditForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/ais")
			return
		}

		var ai models.AI
		if err := db.Preload("Categories").Preload("Genera").First(&ai, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/ais")
			return
		}

		cats, genera := loadAIFormDeps(db)

		selCatIDs := make([]uint, 0, len(ai.Categories))
		for _, cat := range ai.Categories {
			selCatIDs = append(selCatIDs, cat.ID)
		}
		selGenIDs := make([]uint, 0, len(ai.Genera))
		for _, g := range ai.Genera {
			selGenIDs = append(selGenIDs, g.ID)
		}

		ownerID := ""
		ownerLabel := ""
		if ai.UserID != nil {
			ownerID = strconv.FormatUint(uint64(*ai.UserID), 10)
			ownerLabel = ownerLabelByID(db, ownerID)
		}

		selectedArtifactID := ""
		if ai.ArtifactID != nil {
			selectedArtifactID = strconv.FormatUint(uint64(*ai.ArtifactID), 10)
		}
		existingData := template.JS("{}")
		if len(ai.Data) > 0 {
			existingData = template.JS(ai.Data)
		}
		existingScreenshots := template.JS("[]")
		if len(ai.Screenshots) > 0 {
			existingScreenshots = template.JS(ai.Screenshots)
		}

		c.HTML(http.StatusOK, "root_ai_edit.tmpl", gin.H{
			"ActiveNav":           "ais",
			"R2Enabled":           config.C.R2Enabled(),
			"AI":                  ai,
			"OwnerID":             ownerID,
			"OwnerLabel":          ownerLabel,
			"Categories":          cats,
			"Genera":              genera,
			"SelectedCategoryIDs": selCatIDs,
			"SelectedGenusIDs":    selGenIDs,
			"ArtifactsJSON":       loadArtifactsForForm(db),
			"SelectedArtifactID":  selectedArtifactID,
			"ExistingData":        existingData,
			"ExistingScreenshots": existingScreenshots,
		})
	}
}

func AIUpdate(db *gorm.DB, r2 *utils.R2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/ais")
			return
		}

		var ai models.AI
		if err := db.First(&ai, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/ais")
			return
		}

		catIDStrs := c.PostFormArray("category_ids")
		genIDStrs := c.PostFormArray("genus_ids")
		var catIDs, genIDs []uint
		for _, s := range catIDStrs {
			if uid, err := strconv.ParseUint(s, 10, 64); err == nil {
				catIDs = append(catIDs, uint(uid))
			}
		}
		for _, s := range genIDStrs {
			if uid, err := strconv.ParseUint(s, 10, 64); err == nil {
				genIDs = append(genIDs, uint(uid))
			}
		}

		name := strings.TrimSpace(c.PostForm("name"))
		slug := strings.TrimSpace(c.PostForm("slug"))
		if slug == "" {
			slug = slugify(name)
		} else {
			slug = slugify(slug)
		}
		subtitle := strings.TrimSpace(c.PostForm("subtitle"))
		description := strings.TrimSpace(c.PostForm("description"))
		website := strings.TrimSpace(c.PostForm("website"))
		haveFree := c.PostForm("have_free") == "on"
		ownerID := strings.TrimSpace(c.PostForm("owner_id"))
		screenshotsRaw := strings.TrimSpace(c.PostForm("screenshots_json"))
		screenshotsJS := template.JS("[]")
		if screenshotsRaw != "" && json.Valid([]byte(screenshotsRaw)) {
			screenshotsJS = template.JS(screenshotsRaw)
		}

		artifactID, artifactData := parseArtifactFormFields(c)
		selectedArtifactID := ""
		if artifactID != nil {
			selectedArtifactID = strconv.FormatUint(uint64(*artifactID), 10)
		}
		existingData := template.JS("{}")
		if len(artifactData) > 0 {
			existingData = template.JS(artifactData)
		}

		renderErr := func(status int, msg string) {
			cats, genera := loadAIFormDeps(db)
			ownerIDStr := ownerID
			if ownerIDStr == "" && ai.UserID != nil {
				ownerIDStr = strconv.FormatUint(uint64(*ai.UserID), 10)
			}
			c.HTML(status, "root_ai_edit.tmpl", gin.H{
				"ActiveNav":           "ais",
				"R2Enabled":           config.C.R2Enabled(),
				"AI":                  ai,
				"OwnerID":             ownerIDStr,
				"OwnerLabel":          ownerLabelByID(db, ownerIDStr),
				"Categories":          cats,
				"Genera":              genera,
				"SelectedCategoryIDs": catIDs,
				"SelectedGenusIDs":    genIDs,
				"ArtifactsJSON":       loadArtifactsForForm(db),
				"SelectedArtifactID":  selectedArtifactID,
				"ExistingData":        existingData,
				"ExistingScreenshots": screenshotsJS,
				"Error":               msg,
			})
		}

		if name == "" {
			renderErr(http.StatusBadRequest, "Name is required.")
			return
		}
		if slug == "" {
			renderErr(http.StatusBadRequest, "Could not derive a slug — please provide one.")
			return
		}
		var conflict models.AI
		if db.Where("slug = ? AND id != ?", slug, ai.ID).First(&conflict).Error == nil {
			renderErr(http.StatusBadRequest, "Slug already in use.")
			return
		}

		logoURL := ai.LogoURL
		if c.PostForm("remove_logo") == "on" && logoURL != "" && r2 != nil {
			_ = r2.DeleteByURL(logoURL)
			logoURL = ""
		} else if file, err := c.FormFile("logo"); err == nil && file != nil && file.Size > 0 {
			if r2 == nil {
				renderErr(http.StatusBadRequest, "Logo upload is disabled — R2 is not configured.")
				return
			}
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if !allowedLogoExts[ext] {
				renderErr(http.StatusBadRequest, "Logo must be JPG, PNG, GIF, WebP, or SVG.")
				return
			}
			if file.Size > maxLogoBytes {
				renderErr(http.StatusBadRequest, "Logo must be 5 MB or smaller.")
				return
			}
			oldLogo := ai.LogoURL
			url, err := r2.UploadFile(file, "ai-logos")
			if err != nil {
				renderErr(http.StatusInternalServerError, "Failed to upload logo: "+err.Error())
				return
			}
			if oldLogo != "" {
				_ = r2.DeleteByURL(oldLogo)
			}
			logoURL = url
		}

		// Delete screenshots removed by the user from R2.
		type shotItem struct {
			URL string `json:"url"`
		}
		var oldShots, newShots []shotItem
		_ = json.Unmarshal(ai.Screenshots, &oldShots)
		if screenshotsRaw != "" && json.Valid([]byte(screenshotsRaw)) {
			_ = json.Unmarshal([]byte(screenshotsRaw), &newShots)
		}
		newURLSet := make(map[string]bool, len(newShots))
		for _, s := range newShots {
			if s.URL != "" {
				newURLSet[s.URL] = true
			}
		}
		for _, s := range oldShots {
			if s.URL != "" && !newURLSet[s.URL] && r2 != nil {
				_ = r2.DeleteByURL(s.URL)
			}
		}

		var screenshotsData datatypes.JSON = datatypes.JSON("[]")
		if screenshotsRaw != "" && json.Valid([]byte(screenshotsRaw)) {
			screenshotsData = datatypes.JSON(screenshotsRaw)
		}

		ai.Name = name
		ai.Slug = slug
		ai.Subtitle = subtitle
		ai.Description = description
		ai.Website = website
		ai.HaveFree = haveFree
		ai.LogoURL = logoURL
		ai.Screenshots = screenshotsData
		ai.ArtifactID = artifactID
		ai.Data = artifactData
		ai.UserID = nil
		if ownerID != "" {
			if uid, err := strconv.ParseUint(ownerID, 10, 64); err == nil {
				uidU := uint(uid)
				ai.UserID = &uidU
			}
		}

		if err := db.Save(&ai).Error; err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save AI: "+err.Error())
			return
		}

		var cats []models.Category
		if len(catIDs) > 0 {
			db.Where("id IN ?", catIDs).Find(&cats)
		}
		db.Model(&ai).Association("Categories").Replace(cats)

		var genera []models.Genus
		if len(genIDs) > 0 {
			db.Where("id IN ?", genIDs).Find(&genera)
		}
		db.Model(&ai).Association("Genera").Replace(genera)

		c.Redirect(http.StatusFound, "/root/ais")
	}
}

package root

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var slugSanitize = regexp.MustCompile(`[^a-z0-9]+`)

const maxLogoBytes = 5 * 1024 * 1024

var allowedLogoExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true,
}

type aiFormData struct {
	Name        string
	Slug        string
	Subtitle    string
	Description string
	OwnerID     string
	CategoryIDs []uint
	GenusIDs    []uint
}

func loadAIFormDeps(db *gorm.DB) (cats []models.Category, genera []models.Genus, users []models.User) {
	db.Order("name ASC").Find(&cats)
	db.Order("name ASC").Find(&genera)
	db.Order("username ASC, name ASC").Find(&users)
	return
}

func AINewForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cats, genera, users := loadAIFormDeps(db)
		c.HTML(http.StatusOK, "root_ai_new.tmpl", gin.H{
			"ActiveNav":  "ais",
			"R2Enabled":  config.C.R2Enabled(),
			"Form":       aiFormData{},
			"Categories": cats,
			"Genera":     genera,
			"Users":      users,
		})
	}
}

func AICreate(db *gorm.DB, r2 *utils.R2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		catIDStrs := c.PostFormArray("category_ids")
		genIDStrs := c.PostFormArray("genus_ids")

		var catIDs, genIDs []uint
		for _, s := range catIDStrs {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				catIDs = append(catIDs, uint(id))
			}
		}
		for _, s := range genIDStrs {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				genIDs = append(genIDs, uint(id))
			}
		}

		form := aiFormData{
			Name:        strings.TrimSpace(c.PostForm("name")),
			Slug:        strings.TrimSpace(c.PostForm("slug")),
			Subtitle:    strings.TrimSpace(c.PostForm("subtitle")),
			Description: strings.TrimSpace(c.PostForm("description")),
			OwnerID:     strings.TrimSpace(c.PostForm("owner_id")),
			CategoryIDs: catIDs,
			GenusIDs:    genIDs,
		}

		renderErr := func(status int, msg string) {
			cats, genera, users := loadAIFormDeps(db)
			c.HTML(status, "root_ai_new.tmpl", gin.H{
				"ActiveNav":  "ais",
				"R2Enabled":  config.C.R2Enabled(),
				"Form":       form,
				"Categories": cats,
				"Genera":     genera,
				"Users":      users,
				"Error":      msg,
			})
		}

		if form.Name == "" {
			renderErr(http.StatusBadRequest, "Name is required.")
			return
		}
		if form.Slug == "" {
			form.Slug = slugify(form.Name)
		} else {
			form.Slug = slugify(form.Slug)
		}
		if form.Slug == "" {
			renderErr(http.StatusBadRequest, "Could not derive a slug — please provide one.")
			return
		}

		var logoURL string
		if file, err := c.FormFile("logo"); err == nil && file != nil && file.Size > 0 {
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
			url, err := r2.UploadFile(file, "ai-logos")
			if err != nil {
				renderErr(http.StatusInternalServerError, "Failed to upload logo: "+err.Error())
				return
			}
			logoURL = url
		}

		ai := models.AI{
			Name:        form.Name,
			Slug:        form.Slug,
			Subtitle:    form.Subtitle,
			Description: form.Description,
			LogoURL:     logoURL,
		}
		if form.OwnerID != "" {
			uid, err := strconv.ParseUint(form.OwnerID, 10, 64)
			if err == nil {
				uidU := uint(uid)
				ai.UserID = &uidU
			}
		}

		if err := db.Create(&ai).Error; err != nil {
			if logoURL != "" && r2 != nil {
				_ = r2.DeleteByURL(logoURL)
			}
			renderErr(http.StatusInternalServerError, "Failed to save AI: "+err.Error())
			return
		}

		if len(catIDs) > 0 {
			var cats []models.Category
			db.Where("id IN ?", catIDs).Find(&cats)
			db.Model(&ai).Association("Categories").Replace(cats)
		}
		if len(genIDs) > 0 {
			var genera []models.Genus
			db.Where("id IN ?", genIDs).Find(&genera)
			db.Model(&ai).Association("Genera").Replace(genera)
		}

		c.Redirect(http.StatusFound, "/root/ais")
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugSanitize.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

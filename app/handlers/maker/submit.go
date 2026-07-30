package maker

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
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

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugSanitize.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func currentUser(c *gin.Context) *models.User {
	if u, ok := c.MustGet("user").(*models.User); ok {
		return u
	}
	return nil
}

type submitForm struct {
	Name        string
	Website     string
	Subtitle    string
	Description string
}

func SubmitForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		data := headerData(db, user)
		data["Submitted"] = c.Query("submitted") == "1"
		data["Form"] = submitForm{}
		data["R2Enabled"] = config.C.R2Enabled()
		c.HTML(http.StatusOK, "maker_submit.tmpl", data)
	}
}

func SubmitCreate(db *gorm.DB, r2 *utils.R2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)

		form := submitForm{
			Name:        strings.TrimSpace(c.PostForm("name")),
			Website:     strings.TrimSpace(c.PostForm("website")),
			Subtitle:    strings.TrimSpace(c.PostForm("subtitle")),
			Description: strings.TrimSpace(c.PostForm("description")),
		}

		renderErr := func(status int, msg string) {
			data := headerData(db, user)
			data["Form"] = form
			data["R2Enabled"] = config.C.R2Enabled()
			data["Error"] = msg
			c.HTML(status, "maker_submit.tmpl", data)
		}

		if form.Name == "" {
			renderErr(http.StatusUnprocessableEntity, "Name is required.")
			return
		}

		base := slugify(form.Name)
		if base == "" {
			renderErr(http.StatusUnprocessableEntity, "Could not derive a slug from the name — please use regular characters.")
			return
		}

		// Append user ID on conflict to keep slugs unique.
		slug := base
		var conflict models.AI
		if db.Where("slug = ?", slug).First(&conflict).Error == nil {
			slug = fmt.Sprintf("%s-%d", base, user.ID)
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

		uid := user.ID
		ai := models.AI{
			Name:        form.Name,
			Slug:        slug,
			Status:      models.AIStatusPending,
			Subtitle:    form.Subtitle,
			Description: form.Description,
			Website:     form.Website,
			LogoURL:     logoURL,
			UserID:      &uid,
		}

		if err := db.Create(&ai).Error; err != nil {
			if logoURL != "" && r2 != nil {
				_ = r2.DeleteByURL(logoURL)
			}
			renderErr(http.StatusInternalServerError, "Something went wrong — please try again.")
			return
		}

		c.Redirect(http.StatusFound, "/maker/submit?submitted=1")
	}
}

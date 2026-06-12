package root

import (
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

type toolFormData struct {
	Name        string
	Slug        string
	Subtitle    string
	Description string
}

func ToolNewForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "tool_new.tmpl", gin.H{
			"ActiveNav": "tools",
			"R2Enabled": config.C.R2Enabled(),
			"Form":      toolFormData{},
		})
	}
}

func ToolCreate(db *gorm.DB, r2 *utils.R2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		form := toolFormData{
			Name:        strings.TrimSpace(c.PostForm("name")),
			Slug:        strings.TrimSpace(c.PostForm("slug")),
			Subtitle:    strings.TrimSpace(c.PostForm("subtitle")),
			Description: strings.TrimSpace(c.PostForm("description")),
		}

		renderErr := func(status int, msg string) {
			c.HTML(status, "tool_new.tmpl", gin.H{
				"ActiveNav": "tools",
				"R2Enabled": config.C.R2Enabled(),
				"Form":      form,
				"Error":     msg,
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
			url, err := r2.UploadFile(file, "tool-logos")
			if err != nil {
				renderErr(http.StatusInternalServerError, "Failed to upload logo: "+err.Error())
				return
			}
			logoURL = url
		}

		tool := models.Tool{
			Name:        form.Name,
			Slug:        form.Slug,
			Subtitle:    form.Subtitle,
			Description: form.Description,
			LogoURL:     logoURL,
		}
		if err := db.Create(&tool).Error; err != nil {
			if logoURL != "" && r2 != nil {
				_ = r2.DeleteByURL(logoURL)
			}
			renderErr(http.StatusInternalServerError, "Failed to save tool: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/root/tools")
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugSanitize.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

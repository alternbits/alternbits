package root

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type categoryFormData struct {
	Name     string
	Slug     string
	Subtitle string
	ParentID string
}

func CategoryNewForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var parents []models.Category
		if err := db.Order("name ASC").Find(&parents).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_category_new.tmpl", gin.H{
				"ActiveNav": "categories",
				"Form":      categoryFormData{},
				"Error":     "Failed to load parent categories.",
			})
			return
		}
		c.HTML(http.StatusOK, "root_category_new.tmpl", gin.H{
			"ActiveNav": "categories",
			"Form":      categoryFormData{},
			"Parents":   parents,
		})
	}
}

func CategoryCreate(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		form := categoryFormData{
			Name:     strings.TrimSpace(c.PostForm("name")),
			Slug:     strings.TrimSpace(c.PostForm("slug")),
			Subtitle: strings.TrimSpace(c.PostForm("subtitle")),
			ParentID: strings.TrimSpace(c.PostForm("parent_id")),
		}

		renderErr := func(status int, msg string) {
			var parents []models.Category
			_ = db.Order("name ASC").Find(&parents).Error
			c.HTML(status, "root_category_new.tmpl", gin.H{
				"ActiveNav": "categories",
				"Form":      form,
				"Parents":   parents,
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

		category := models.Category{
			Name:     form.Name,
			Slug:     form.Slug,
			Subtitle: form.Subtitle,
		}
		parentName := ""
		if form.ParentID != "" {
			pid, err := strconv.ParseUint(form.ParentID, 10, 64)
			if err != nil {
				renderErr(http.StatusBadRequest, "Invalid parent category.")
				return
			}
			pidU := uint(pid)
			category.ParentID = &pidU
			var parent models.Category
			if db.Select("name").First(&parent, pidU).Error == nil {
				parentName = parent.Name
			}
		}

		if err := db.Create(&category).Error; err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save category: "+err.Error())
			return
		}

		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyCategoryCreated(actor, &category, parentName)

		c.Redirect(http.StatusFound, "/root/categories")
	}
}

package root

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CategoryEditForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/categories")
			return
		}

		var category models.Category
		if err := db.First(&category, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/categories")
			return
		}

		var parents []models.Category
		if err := db.Where("id != ?", id).Order("name ASC").Find(&parents).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_category_edit.tmpl", gin.H{
				"ActiveNav": "categories",
				"Category":  category,
				"Error":     "Failed to load parent categories.",
			})
			return
		}

		parentID := ""
		if category.ParentID != nil {
			parentID = strconv.FormatUint(uint64(*category.ParentID), 10)
		}

		c.HTML(http.StatusOK, "root_category_edit.tmpl", gin.H{
			"ActiveNav": "categories",
			"Category":  category,
			"ParentID":  parentID,
			"Parents":   parents,
		})
	}
}

func CategoryUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/categories")
			return
		}

		var category models.Category
		if err := db.First(&category, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/categories")
			return
		}

		name := strings.TrimSpace(c.PostForm("name"))
		slug := strings.TrimSpace(c.PostForm("slug"))
		if slug == "" {
			slug = slugify(name)
		} else {
			slug = slugify(slug)
		}
		subtitle := strings.TrimSpace(c.PostForm("subtitle"))
		parentIDStr := strings.TrimSpace(c.PostForm("parent_id"))

		renderErr := func(status int, msg string) {
			var parents []models.Category
			_ = db.Where("id != ?", id).Order("name ASC").Find(&parents).Error
			c.HTML(status, "root_category_edit.tmpl", gin.H{
				"ActiveNav": "categories",
				"Category":  category,
				"ParentID":  parentIDStr,
				"Parents":   parents,
				"Error":     msg,
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
		var conflict models.Category
		if db.Where("slug = ? AND id != ?", slug, category.ID).First(&conflict).Error == nil {
			renderErr(http.StatusBadRequest, "Slug already in use.")
			return
		}

		category.Name = name
		category.Slug = slug
		category.Subtitle = subtitle
		category.ParentID = nil
		if parentIDStr != "" {
			pid, err := strconv.ParseUint(parentIDStr, 10, 64)
			if err != nil {
				renderErr(http.StatusBadRequest, "Invalid parent category.")
				return
			}
			pidU := uint(pid)
			category.ParentID = &pidU
		}

		if err := db.Save(&category).Error; err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save category: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/root/categories")
	}
}

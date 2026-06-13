package root

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PageEditForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/pages")
			return
		}

		var pg models.Page
		if err := db.First(&pg, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/pages")
			return
		}

		c.HTML(http.StatusOK, "root_page_edit.tmpl", gin.H{
			"ActiveNav": "pages",
			"Page":      pg,
		})
	}
}

func PageUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/pages")
			return
		}

		var pg models.Page
		if err := db.First(&pg, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/pages")
			return
		}

		title := strings.TrimSpace(c.PostForm("title"))
		sl := strings.TrimSpace(c.PostForm("slug"))
		subtitle := strings.TrimSpace(c.PostForm("subtitle"))
		description := strings.TrimSpace(c.PostForm("description"))

		renderErr := func(status int, msg string) {
			c.HTML(status, "root_page_edit.tmpl", gin.H{
				"ActiveNav": "pages",
				"Page":      pg,
				"Error":     msg,
			})
		}

		if title == "" {
			renderErr(http.StatusBadRequest, "Title is required.")
			return
		}
		if sl == "" {
			sl = slugify(title)
		} else {
			sl = slugify(sl)
		}
		if sl == "" {
			renderErr(http.StatusBadRequest, "Could not derive a slug — please provide one.")
			return
		}

		pg.Title = title
		pg.Slug = sl
		pg.Subtitle = subtitle
		pg.Description = description

		if err := db.Save(&pg).Error; err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save page: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/root/pages")
	}
}

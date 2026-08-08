package root

import (
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pageFormData struct {
	Title       string
	Slug        string
	Subtitle    string
	Description string
}

func PageNewForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "root_page_new.tmpl", gin.H{
			"ActiveNav": "pages",
			"Form":      pageFormData{},
		})
	}
}

func PageCreate(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		form := pageFormData{
			Title:       strings.TrimSpace(c.PostForm("title")),
			Slug:        strings.TrimSpace(c.PostForm("slug")),
			Subtitle:    strings.TrimSpace(c.PostForm("subtitle")),
			Description: strings.TrimSpace(c.PostForm("description")),
		}

		renderErr := func(status int, msg string) {
			c.HTML(status, "root_page_new.tmpl", gin.H{
				"ActiveNav": "pages",
				"Form":      form,
				"Error":     msg,
			})
		}

		if form.Title == "" {
			renderErr(http.StatusBadRequest, "Title is required.")
			return
		}
		if form.Slug == "" {
			form.Slug = slugify(form.Title)
		} else {
			form.Slug = slugify(form.Slug)
		}
		if form.Slug == "" {
			renderErr(http.StatusBadRequest, "Could not derive a slug — please provide one.")
			return
		}

		pg := models.Page{
			Title:       form.Title,
			Slug:        form.Slug,
			Subtitle:    form.Subtitle,
			Description: form.Description,
		}

		if err := db.Create(&pg).Error; err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save page: "+err.Error())
			return
		}

		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyPageCreated(actor, &pg)

		c.Redirect(http.StatusFound, "/root/pages")
	}
}

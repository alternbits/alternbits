package root

import (
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type genusFormData struct {
	Name     string
	Slug     string
	Subtitle string
}

func GenusNewForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "root_genus_new.tmpl", gin.H{
			"ActiveNav": "genera",
			"Form":      genusFormData{},
		})
	}
}

func GenusCreate(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		form := genusFormData{
			Name:     strings.TrimSpace(c.PostForm("name")),
			Slug:     strings.TrimSpace(c.PostForm("slug")),
			Subtitle: strings.TrimSpace(c.PostForm("subtitle")),
		}

		renderErr := func(status int, msg string) {
			c.HTML(status, "root_genus_new.tmpl", gin.H{
				"ActiveNav": "genera",
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

		genus := models.Genus{
			Name:     form.Name,
			Slug:     form.Slug,
			Subtitle: form.Subtitle,
		}

		if err := db.Create(&genus).Error; err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save genus: "+err.Error())
			return
		}

		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyGenusCreated(actor, &genus)

		c.Redirect(http.StatusFound, "/root/genera")
	}
}

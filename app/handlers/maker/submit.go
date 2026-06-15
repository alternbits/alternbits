package maker

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var slugSanitize = regexp.MustCompile(`[^a-z0-9]+`)

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
		c.HTML(http.StatusOK, "maker_submit.tmpl", data)
	}
}

func SubmitCreate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)

		form := submitForm{
			Name:        strings.TrimSpace(c.PostForm("name")),
			Website:     strings.TrimSpace(c.PostForm("website")),
			Subtitle:    strings.TrimSpace(c.PostForm("subtitle")),
			Description: strings.TrimSpace(c.PostForm("description")),
		}

		renderErr := func(msg string) {
			data := headerData(db, user)
			data["Form"] = form
			data["Error"] = msg
			c.HTML(http.StatusUnprocessableEntity, "maker_submit.tmpl", data)
		}

		if form.Name == "" {
			renderErr("Name is required.")
			return
		}

		base := slugify(form.Name)
		if base == "" {
			renderErr("Could not derive a slug from the name — please use regular characters.")
			return
		}

		// Append user ID on conflict to keep slugs unique.
		slug := base
		var conflict models.AI
		if db.Where("slug = ?", slug).First(&conflict).Error == nil {
			slug = fmt.Sprintf("%s-%d", base, user.ID)
		}

		uid := user.ID
		ai := models.AI{
			Name:        form.Name,
			Slug:        slug,
			Status:      models.AIStatusPending,
			Subtitle:    form.Subtitle,
			Description: form.Description,
			Website:     form.Website,
			UserID:      &uid,
		}

		if err := db.Create(&ai).Error; err != nil {
			renderErr("Something went wrong — please try again.")
			return
		}

		c.Redirect(http.StatusFound, "/maker/submit?submitted=1")
	}
}

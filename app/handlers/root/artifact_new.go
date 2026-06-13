package root

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type artifactFieldInput struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	FieldType    string `json:"field_type"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"default_value"`
}

func ArtifactNewForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "root_artifact_new.tmpl", gin.H{
			"ActiveNav": "artifacts",
		})
	}
}

func ArtifactCreate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		sl := strings.TrimSpace(c.PostForm("slug"))
		fieldsJSON := strings.TrimSpace(c.PostForm("fields_json"))

		renderErr := func(status int, msg string) {
			c.HTML(status, "root_artifact_new.tmpl", gin.H{
				"ActiveNav":  "artifacts",
				"FormName":   name,
				"FormSlug":   sl,
				"FieldsJSON": fieldsJSON,
				"Error":      msg,
			})
		}

		if name == "" {
			renderErr(http.StatusBadRequest, "Name is required.")
			return
		}
		if sl == "" {
			sl = slugify(name)
		} else {
			sl = slugify(sl)
		}
		if sl == "" {
			renderErr(http.StatusBadRequest, "Could not derive a slug — please provide one.")
			return
		}

		var fieldInputs []artifactFieldInput
		if fieldsJSON != "" {
			if err := json.Unmarshal([]byte(fieldsJSON), &fieldInputs); err != nil {
				renderErr(http.StatusBadRequest, "Invalid fields data.")
				return
			}
		}

		artifact := models.Artifact{Name: name, Slug: sl}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&artifact).Error; err != nil {
				return err
			}
			for _, f := range fieldInputs {
				fName := strings.TrimSpace(f.Name)
				fSlug := strings.TrimSpace(f.Slug)
				if fName == "" {
					continue
				}
				if fSlug == "" {
					fSlug = slugify(fName)
				}
				ft := models.ArtifactFieldType(f.FieldType)
				switch ft {
				case models.ArtifactFieldTypeString,
					models.ArtifactFieldTypeText,
					models.ArtifactFieldTypeNumber,
					models.ArtifactFieldTypeInteger,
					models.ArtifactFieldTypeBoolean,
					models.ArtifactFieldTypeURL,
					models.ArtifactFieldTypeEmail,
					models.ArtifactFieldTypeDate:
				default:
					ft = models.ArtifactFieldTypeString
				}
				field := models.ArtifactField{
					ArtifactID: artifact.ID,
					Name:       fName,
					Slug:       fSlug,
					FieldType:  ft,
					Required:   f.Required,
				}
				if f.DefaultValue != "" {
					field.DefaultValue = &f.DefaultValue
				}
				if err := tx.Create(&field).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			renderErr(http.StatusInternalServerError, "Failed to save artifact: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/root/artifacts")
	}
}

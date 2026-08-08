package root

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ArtifactEditForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/artifacts")
			return
		}

		var artifact models.Artifact
		if err := db.Preload("Fields").First(&artifact, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/artifacts")
			return
		}

		fieldsJSON := buildFieldsJSON(artifact.Fields)

		c.HTML(http.StatusOK, "root_artifact_edit.tmpl", gin.H{
			"ActiveNav":  "artifacts",
			"Artifact":   artifact,
			"FieldsJSON": template.JS(fieldsJSON),
		})
	}
}

func ArtifactUpdate(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/artifacts")
			return
		}

		var artifact models.Artifact
		if err := db.Preload("Fields").First(&artifact, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/artifacts")
			return
		}

		name := strings.TrimSpace(c.PostForm("name"))
		sl := strings.TrimSpace(c.PostForm("slug"))
		fieldsRaw := strings.TrimSpace(c.PostForm("fields_json"))

		renderErr := func(status int, msg string) {
			fj := template.JS(fieldsRaw)
			if fj == "" {
				fj = template.JS(buildFieldsJSON(artifact.Fields))
			}
			c.HTML(status, "root_artifact_edit.tmpl", gin.H{
				"ActiveNav":  "artifacts",
				"Artifact":   artifact,
				"FieldsJSON": fj,
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
		if fieldsRaw != "" {
			if err := json.Unmarshal([]byte(fieldsRaw), &fieldInputs); err != nil {
				renderErr(http.StatusBadRequest, "Invalid fields data.")
				return
			}
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			artifact.Name = name
			artifact.Slug = sl
			if err := tx.Save(&artifact).Error; err != nil {
				return err
			}

			if err := tx.Where("artifact_id = ?", artifact.ID).Delete(&models.ArtifactField{}).Error; err != nil {
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

		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyArtifactUpdated(actor, &artifact, len(fieldInputs))

		c.Redirect(http.StatusFound, "/root/artifacts")
	}
}

func buildFieldsJSON(fields []models.ArtifactField) string {
	type fieldOut struct {
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		FieldType    string `json:"field_type"`
		Required     bool   `json:"required"`
		DefaultValue string `json:"default_value"`
	}
	out := make([]fieldOut, len(fields))
	for i, f := range fields {
		dv := ""
		if f.DefaultValue != nil {
			dv = *f.DefaultValue
		}
		out[i] = fieldOut{
			Name:         f.Name,
			Slug:         f.Slug,
			FieldType:    string(f.FieldType),
			Required:     f.Required,
			DefaultValue: dv,
		}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

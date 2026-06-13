package root

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var slugSanitize = regexp.MustCompile(`[^a-z0-9]+`)

const maxLogoBytes = 5 * 1024 * 1024

var allowedLogoExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true,
}

type aiFormData struct {
	Name        string
	Slug        string
	Subtitle    string
	Description string
	OwnerID     string
	OwnerLabel  string
	CategoryIDs []uint
	GenusIDs    []uint
}

func loadAIFormDeps(db *gorm.DB) (cats []models.Category, genera []models.Genus) {
	db.Order("name ASC").Find(&cats)
	db.Order("name ASC").Find(&genera)
	return
}

func formatOwnerLabel(u models.User) string {
	if u.Username != "" && u.Name != "" {
		return "@" + u.Username + " (" + u.Name + ")"
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}

func ownerLabelByID(db *gorm.DB, idStr string) string {
	if idStr == "" {
		return ""
	}
	var u models.User
	if db.Select("id, name, username, email").First(&u, idStr).Error != nil {
		return ""
	}
	return formatOwnerLabel(u)
}

type artifactOptionField struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	FieldType    string `json:"field_type"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"default_value"`
}

type artifactOption struct {
	ID     uint                  `json:"id"`
	Name   string                `json:"name"`
	Fields []artifactOptionField `json:"fields"`
}

func loadArtifactsForForm(db *gorm.DB) template.JS {
	var artifacts []models.Artifact
	db.Preload("Fields").Order("name ASC").Find(&artifacts)
	opts := make([]artifactOption, len(artifacts))
	for i, a := range artifacts {
		fields := make([]artifactOptionField, len(a.Fields))
		for j, f := range a.Fields {
			dv := ""
			if f.DefaultValue != nil {
				dv = *f.DefaultValue
			}
			fields[j] = artifactOptionField{
				Name:         f.Name,
				Slug:         f.Slug,
				FieldType:    string(f.FieldType),
				Required:     f.Required,
				DefaultValue: dv,
			}
		}
		opts[i] = artifactOption{ID: a.ID, Name: a.Name, Fields: fields}
	}
	b, _ := json.Marshal(opts)
	return template.JS(b)
}

func parseArtifactFormFields(c *gin.Context) (artifactID *uint, data datatypes.JSON) {
	idStr := strings.TrimSpace(c.PostForm("artifact_id"))
	dataRaw := strings.TrimSpace(c.PostForm("artifact_data_json"))
	if idStr == "" {
		return nil, nil
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, nil
	}
	uid := uint(id)
	if dataRaw != "" && json.Valid([]byte(dataRaw)) {
		return &uid, datatypes.JSON(dataRaw)
	}
	return &uid, datatypes.JSON("{}")
}

func AINewForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cats, genera := loadAIFormDeps(db)
		c.HTML(http.StatusOK, "root_ai_new.tmpl", gin.H{
			"ActiveNav":          "ais",
			"R2Enabled":          config.C.R2Enabled(),
			"Form":               aiFormData{},
			"Categories":         cats,
			"Genera":             genera,
			"ArtifactsJSON":      loadArtifactsForForm(db),
			"SelectedArtifactID": "",
			"ExistingData":       template.JS("{}"),
		})
	}
}

func AICreate(db *gorm.DB, r2 *utils.R2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		catIDStrs := c.PostFormArray("category_ids")
		genIDStrs := c.PostFormArray("genus_ids")

		var catIDs, genIDs []uint
		for _, s := range catIDStrs {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				catIDs = append(catIDs, uint(id))
			}
		}
		for _, s := range genIDStrs {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				genIDs = append(genIDs, uint(id))
			}
		}

		ownerID := strings.TrimSpace(c.PostForm("owner_id"))
		form := aiFormData{
			Name:        strings.TrimSpace(c.PostForm("name")),
			Slug:        strings.TrimSpace(c.PostForm("slug")),
			Subtitle:    strings.TrimSpace(c.PostForm("subtitle")),
			Description: strings.TrimSpace(c.PostForm("description")),
			OwnerID:     ownerID,
			OwnerLabel:  ownerLabelByID(db, ownerID),
			CategoryIDs: catIDs,
			GenusIDs:    genIDs,
		}

		artifactID, artifactData := parseArtifactFormFields(c)
		selectedArtifactID := ""
		if artifactID != nil {
			selectedArtifactID = strconv.FormatUint(uint64(*artifactID), 10)
		}
		existingData := template.JS("{}")
		if len(artifactData) > 0 {
			existingData = template.JS(artifactData)
		}

		renderErr := func(status int, msg string) {
			cats, genera := loadAIFormDeps(db)
			c.HTML(status, "root_ai_new.tmpl", gin.H{
				"ActiveNav":          "ais",
				"R2Enabled":          config.C.R2Enabled(),
				"Form":               form,
				"Categories":         cats,
				"Genera":             genera,
				"ArtifactsJSON":      loadArtifactsForForm(db),
				"SelectedArtifactID": selectedArtifactID,
				"ExistingData":       existingData,
				"Error":              msg,
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
			url, err := r2.UploadFile(file, "ai-logos")
			if err != nil {
				renderErr(http.StatusInternalServerError, "Failed to upload logo: "+err.Error())
				return
			}
			logoURL = url
		}

		ai := models.AI{
			Name:        form.Name,
			Slug:        form.Slug,
			Subtitle:    form.Subtitle,
			Description: form.Description,
			LogoURL:     logoURL,
			ArtifactID:  artifactID,
			Data:        artifactData,
		}
		if form.OwnerID != "" {
			uid, err := strconv.ParseUint(form.OwnerID, 10, 64)
			if err == nil {
				uidU := uint(uid)
				ai.UserID = &uidU
			}
		}

		if err := db.Create(&ai).Error; err != nil {
			if logoURL != "" && r2 != nil {
				_ = r2.DeleteByURL(logoURL)
			}
			renderErr(http.StatusInternalServerError, "Failed to save AI: "+err.Error())
			return
		}

		if len(catIDs) > 0 {
			var cats []models.Category
			db.Where("id IN ?", catIDs).Find(&cats)
			db.Model(&ai).Association("Categories").Replace(cats)
		}
		if len(genIDs) > 0 {
			var genera []models.Genus
			db.Where("id IN ?", genIDs).Find(&genera)
			db.Model(&ai).Association("Genera").Replace(genera)
		}

		c.Redirect(http.StatusFound, "/root/ais")
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugSanitize.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

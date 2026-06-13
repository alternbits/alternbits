package root

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type alternativesFilters struct {
	Search   string
	Status   string // "pending" | "approved" | "rejected" | ""
	Source   string // "admin" | "user" | ""
	Sort     string
	QueryStr string
}

func AIAlternativesAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
		var alts []models.Alternative
		db.Preload("AI").Preload("AlternativeAI").
			Where("(ai_id = ? OR alternative_ai_id = ?) AND status = ?",
				id, id, models.AlternativeStatusApproved).
			Find(&alts)

		seen := map[uint]bool{uint(id): true}
		results := make([]gin.H, 0, len(alts))
		for _, a := range alts {
			var other *models.AI
			if a.AIID == uint(id) {
				other = a.AlternativeAI
			} else {
				other = a.AI
			}
			if other == nil || seen[other.ID] {
				continue
			}
			seen[other.ID] = true
			label := other.Name
			if other.Slug != "" {
				label += " (" + other.Slug + ")"
			}
			results = append(results, gin.H{
				"id":    other.ID,
				"label": label,
				"name":  other.Name,
				"slug":  other.Slug,
			})
		}
		c.JSON(http.StatusOK, results)
	}
}

func AIsSearchAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
		var ais []models.AI
		db.Table("ais").Where("deleted_at IS NULL").
			Where("name ILIKE ? OR slug ILIKE ?", "%"+q+"%", "%"+q+"%").
			Select("id, name, slug").
			Order("name ASC").
			Limit(10).Find(&ais)
		results := make([]gin.H, 0, len(ais))
		for _, a := range ais {
			label := a.Name
			if a.Slug != "" {
				label += " (" + a.Slug + ")"
			}
			results = append(results, gin.H{"id": a.ID, "label": label})
		}
		c.JSON(http.StatusOK, results)
	}
}

func parseIDList(raw string) []uint64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]uint64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseUint(p, 10, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func parseIDSet(raw string) map[uint64]bool {
	ids := parseIDList(raw)
	if ids == nil {
		return map[uint64]bool{}
	}
	set := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func aiLabelByID(db *gorm.DB, idStr string) string {
	if idStr == "" {
		return ""
	}
	var a models.AI
	if db.Select("id, name, slug").First(&a, idStr).Error != nil {
		return ""
	}
	if a.Slug != "" {
		return a.Name + " (" + a.Slug + ")"
	}
	return a.Name
}

func AlternativesListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := strings.TrimSpace(c.Query("q"))
		status := c.Query("status")
		source := c.Query("source")
		sort := c.Query("sort")

		base := db.Model(&models.Alternative{})

		if search != "" {
			likeQ := "%" + search + "%"
			sub := "SELECT id FROM ais WHERE deleted_at IS NULL AND (name ILIKE ? OR slug ILIKE ?)"
			base = base.Where(
				"ai_id IN ("+sub+") OR alternative_ai_id IN ("+sub+")",
				likeQ, likeQ, likeQ, likeQ,
			)
		}
		switch status {
		case "pending", "approved", "rejected":
			base = base.Where("status = ?", status)
		}
		switch source {
		case "admin":
			base = base.Where("suggested_by_user_id IS NULL")
		case "user":
			base = base.Where("suggested_by_user_id IS NOT NULL")
		}

		orderClause := "created_at DESC"
		switch sort {
		case "oldest":
			orderClause = "created_at ASC"
		case "status":
			orderClause = "status ASC, created_at DESC"
		}

		var alts []models.Alternative
		base.Preload("AI").Preload("AlternativeAI").Preload("SuggestedBy").
			Order(orderClause).Find(&alts)

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if status != "" {
			params.Set("status", status)
		}
		if source != "" {
			params.Set("source", source)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		c.HTML(http.StatusOK, "root_alternatives.tmpl", gin.H{
			"ActiveNav":    "alternatives",
			"Alternatives": alts,
			"Filters": alternativesFilters{
				Search:   search,
				Status:   status,
				Source:   source,
				Sort:     sort,
				QueryStr: queryStr,
			},
		})
	}
}

func AlternativeNewForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pre-fill from ?ai_id= query param (e.g. when linked from AI edit page)
		aiIDStr := c.Query("ai_id")
		aiLabel := aiLabelByID(db, aiIDStr)

		c.HTML(http.StatusOK, "root_alternative_new.tmpl", gin.H{
			"ActiveNav":  "alternatives",
			"AIIDStr":    aiIDStr,
			"AILabel":    aiLabel,
			"AltAIIDStr": "",
			"AltAILabel": "",
			"TwoWay":     false,
			"Status":     string(models.AlternativeStatusApproved),
			"Note":       "",
		})
	}
}

func AlternativeCreate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		aiIDStr := strings.TrimSpace(c.PostForm("ai_id"))
		altAIIDStr := strings.TrimSpace(c.PostForm("alternative_ai_id"))
		twoWay := c.PostForm("two_way") == "on"
		status := models.AlternativeStatus(strings.TrimSpace(c.PostForm("status")))
		note := strings.TrimSpace(c.PostForm("note"))

		// Normalise status
		if status != models.AlternativeStatusPending &&
			status != models.AlternativeStatusApproved &&
			status != models.AlternativeStatusRejected {
			status = models.AlternativeStatusApproved
		}

		renderErr := func(msg string) {
			c.HTML(http.StatusBadRequest, "root_alternative_new.tmpl", gin.H{
				"ActiveNav":  "alternatives",
				"Error":      msg,
				"AIIDStr":    aiIDStr,
				"AILabel":    aiLabelByID(db, aiIDStr),
				"AltAIIDStr": altAIIDStr,
				"AltAILabel": aiLabelByID(db, altAIIDStr),
				"TwoWay":     twoWay,
				"Status":     string(status),
				"Note":       note,
			})
		}

		if aiIDStr == "" || altAIIDStr == "" {
			renderErr("Both AI and alternative AI are required.")
			return
		}
		aiID, err := strconv.ParseUint(aiIDStr, 10, 64)
		if err != nil {
			renderErr("Invalid AI.")
			return
		}
		altAIID, err := strconv.ParseUint(altAIIDStr, 10, 64)
		if err != nil {
			renderErr("Invalid alternative AI.")
			return
		}
		if aiID == altAIID {
			renderErr("An AI cannot be its own alternative.")
			return
		}

		var existing models.Alternative
		if db.Where("ai_id = ? AND alternative_ai_id = ?", aiID, altAIID).First(&existing).Error == nil {
			renderErr("This alternative relationship already exists.")
			return
		}

		alt := models.Alternative{
			AIID:            uint(aiID),
			AlternativeAIID: uint(altAIID),
			Status:          status,
			Note:            note,
		}
		if err := db.Create(&alt).Error; err != nil {
			renderErr("Failed to save: " + err.Error())
			return
		}

		if twoWay {
			var rev models.Alternative
			if db.Where("ai_id = ? AND alternative_ai_id = ?", altAIID, aiID).First(&rev).Error != nil {
				db.Create(&models.Alternative{
					AIID:            uint(altAIID),
					AlternativeAIID: uint(aiID),
					Status:          status,
					Note:            note,
				})
			}
		}

		// Extras — also link these AIs (alternatives of the picked alternative) to the main AI.
		extraTwoWay := parseIDSet(c.PostForm("extra_twoway_ids"))
		for _, eid := range parseIDList(c.PostForm("extra_ai_ids")) {
			if eid == aiID {
				continue
			}
			var existingExtra models.Alternative
			if db.Where("ai_id = ? AND alternative_ai_id = ?", aiID, eid).First(&existingExtra).Error != nil {
				db.Create(&models.Alternative{
					AIID:            uint(aiID),
					AlternativeAIID: uint(eid),
					Status:          status,
					Note:            note,
				})
			}
			if extraTwoWay[eid] {
				var revExtra models.Alternative
				if db.Where("ai_id = ? AND alternative_ai_id = ?", eid, aiID).First(&revExtra).Error != nil {
					db.Create(&models.Alternative{
						AIID:            uint(eid),
						AlternativeAIID: uint(aiID),
						Status:          status,
						Note:            note,
					})
				}
			}
		}

		c.Redirect(http.StatusFound, "/root/alternatives")
	}
}

func AlternativeApprove(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		db.Model(&models.Alternative{}).Where("id = ?", id).
			Update("status", models.AlternativeStatusApproved)
		c.Redirect(http.StatusFound, "/root/alternatives")
	}
}

func AlternativeReject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		db.Model(&models.Alternative{}).Where("id = ?", id).
			Update("status", models.AlternativeStatusRejected)
		c.Redirect(http.StatusFound, "/root/alternatives")
	}
}

func AlternativeDelete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		db.Delete(&models.Alternative{}, id)
		c.Redirect(http.StatusFound, "/root/alternatives")
	}
}

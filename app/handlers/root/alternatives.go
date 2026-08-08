package root

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type alternativesFilters struct {
	Search   string
	Status   string // "pending" | "approved" | "rejected" | ""
	Sort     string
	QueryStr string
}

type aiAlternativeSummary struct {
	ID            uint
	Name          string
	Slug          string
	TotalCount    int64
	ApprovedCount int64
	PendingCount  int64
	RejectedCount int64
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
		sortBy := c.Query("sort")

		whereClause := "a.deleted_at IS NULL AND alt.deleted_at IS NULL"
		args := []interface{}{}

		if search != "" {
			whereClause += " AND (a.name ILIKE ? OR a.slug ILIKE ?)"
			like := "%" + search + "%"
			args = append(args, like, like)
		}
		if status == "pending" || status == "approved" || status == "rejected" {
			whereClause += " AND EXISTS (SELECT 1 FROM alternatives a2 WHERE (a2.ai_id = a.id OR a2.alternative_ai_id = a.id) AND a2.status = ? AND a2.deleted_at IS NULL)"
			args = append(args, status)
		}

		orderClause := "pending_count DESC, total_count DESC, a.name ASC"
		switch sortBy {
		case "name_asc":
			orderClause = "a.name ASC"
		case "name_desc":
			orderClause = "a.name DESC"
		case "most":
			orderClause = "total_count DESC, a.name ASC"
		}

		query := fmt.Sprintf(`
			SELECT a.id, a.name, a.slug,
				COUNT(alt.id) AS total_count,
				SUM(CASE WHEN alt.status = 'approved' THEN 1 ELSE 0 END) AS approved_count,
				SUM(CASE WHEN alt.status = 'pending'  THEN 1 ELSE 0 END) AS pending_count,
				SUM(CASE WHEN alt.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_count
			FROM ais a
			INNER JOIN alternatives alt ON (alt.ai_id = a.id OR alt.alternative_ai_id = a.id) AND alt.deleted_at IS NULL
			WHERE %s
			GROUP BY a.id, a.name, a.slug
			ORDER BY %s
		`, whereClause, orderClause)

		var summaries []aiAlternativeSummary
		db.Raw(query, args...).Scan(&summaries)

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if status != "" {
			params.Set("status", status)
		}
		if sortBy != "" {
			params.Set("sort", sortBy)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		c.HTML(http.StatusOK, "root_alternatives.tmpl", gin.H{
			"ActiveNav": "alternatives",
			"Summaries": summaries,
			"Filters": alternativesFilters{
				Search:   search,
				Status:   status,
				Sort:     sortBy,
				QueryStr: queryStr,
			},
		})
	}
}

func AIAlternativesEditHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		var ai models.AI
		if db.First(&ai, id).Error != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		var alts []models.Alternative
		db.Preload("AI").Preload("AlternativeAI").Preload("SuggestedBy").
			Where("ai_id = ? OR alternative_ai_id = ?", id, id).
			Order("CASE status WHEN 'pending' THEN 0 WHEN 'approved' THEN 1 ELSE 2 END, created_at DESC").
			Find(&alts)
		c.HTML(http.StatusOK, "root_ai_alternatives.tmpl", gin.H{
			"ActiveNav":    "alternatives",
			"AI":           ai,
			"Alternatives": alts,
			"BackURL":      "/root/alternatives/ai/" + idStr,
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

func AlternativeCreate(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
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

		if aiIDStr == "" {
			renderErr("AI is required.")
			return
		}
		aiID, err := strconv.ParseUint(aiIDStr, 10, 64)
		if err != nil {
			renderErr("Invalid AI.")
			return
		}

		extraIDs := parseIDList(c.PostForm("extra_ai_ids"))

		if altAIIDStr == "" && len(extraIDs) == 0 {
			renderErr("Select an alternative AI or at least one item from the also-add list.")
			return
		}

		actor, _ := c.MustGet("user").(*models.User)

		if altAIIDStr != "" {
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

			tg.NotifyAlternativeCreated(actor, aiLabelByID(db, aiIDStr), aiLabelByID(db, altAIIDStr), twoWay, status)
		}

		// Extras — also link these AIs (alternatives of the picked alternative) to the main AI.
		extraTwoWay := parseIDSet(c.PostForm("extra_twoway_ids"))
		var addedExtraLabels []string
		for _, eid := range extraIDs {
			if eid == aiID {
				continue
			}
			added := false
			var existingExtra models.Alternative
			if db.Where("ai_id = ? AND alternative_ai_id = ?", aiID, eid).First(&existingExtra).Error != nil {
				db.Create(&models.Alternative{
					AIID:            uint(aiID),
					AlternativeAIID: uint(eid),
					Status:          status,
					Note:            note,
				})
				added = true
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
					added = true
				}
			}
			if added {
				addedExtraLabels = append(addedExtraLabels, aiLabelByID(db, strconv.FormatUint(eid, 10)))
			}
		}
		tg.NotifyAlternativeExtras(actor, aiLabelByID(db, aiIDStr), addedExtraLabels)

		c.Redirect(http.StatusFound, "/root/alternatives")
	}
}

func altRedirect(c *gin.Context) string {
	if back := c.PostForm("back"); back != "" {
		return back
	}
	return "/root/alternatives"
}

func AlternativeApprove(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		var alt models.Alternative
		db.Preload("AI").Preload("AlternativeAI").First(&alt, id)
		db.Model(&models.Alternative{}).Where("id = ?", id).
			Update("status", models.AlternativeStatusApproved)
		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyAlternativeStatusChanged(actor, &alt, models.AlternativeStatusApproved)
		c.Redirect(http.StatusFound, altRedirect(c))
	}
}

func AlternativeReject(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		var alt models.Alternative
		db.Preload("AI").Preload("AlternativeAI").First(&alt, id)
		db.Model(&models.Alternative{}).Where("id = ?", id).
			Update("status", models.AlternativeStatusRejected)
		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyAlternativeStatusChanged(actor, &alt, models.AlternativeStatusRejected)
		c.Redirect(http.StatusFound, altRedirect(c))
	}
}

func AlternativeDelete(db *gorm.DB, tg *utils.TelegramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/alternatives")
			return
		}
		var alt models.Alternative
		db.Preload("AI").Preload("AlternativeAI").First(&alt, id)
		db.Delete(&models.Alternative{}, id)
		actor, _ := c.MustGet("user").(*models.User)
		tg.NotifyAlternativeDeleted(actor, &alt)
		c.Redirect(http.StatusFound, altRedirect(c))
	}
}

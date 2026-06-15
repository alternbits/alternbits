package root

import (
	"net/http"
	"time"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type dayPoint struct {
	Label string
	Count int
	Pct   int // 0..100, height relative to series max
}

type taxonomyTop struct {
	Name  string
	Slug  string
	Count int64
}

func DashboardHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			users, ais, categories, lists, genera, artifacts, pages int64
			admins, totpUsers                                       int64
			altsTotal, altsPending, altsApproved, altsRejected      int64
			uncategorized, missingGenera, missingLogo               int64
			missingDescription, missingSubtitle                     int64
			emptyLists, privateLists                                int64
			aisPending                                              int64
		)

		db.Model(&models.User{}).Count(&users)
		db.Model(&models.AI{}).Count(&ais)
		db.Model(&models.Category{}).Count(&categories)
		db.Model(&models.List{}).Count(&lists)
		db.Model(&models.Genus{}).Count(&genera)
		db.Model(&models.Artifact{}).Count(&artifacts)
		db.Model(&models.Page{}).Count(&pages)
		db.Model(&models.User{}).Where(&models.User{IsAdmin: true}).Count(&admins)
		db.Model(&models.User{}).Where(&models.User{TOTPEnabled: true}).Count(&totpUsers)

		db.Model(&models.Alternative{}).Count(&altsTotal)
		db.Model(&models.Alternative{}).Where("status = ?", models.AlternativeStatusPending).Count(&altsPending)
		db.Model(&models.Alternative{}).Where("status = ?", models.AlternativeStatusApproved).Count(&altsApproved)
		db.Model(&models.Alternative{}).Where("status = ?", models.AlternativeStatusRejected).Count(&altsRejected)

		db.Model(&models.AI{}).
			Where("id NOT IN (SELECT ai_id FROM ai_categories)").
			Count(&uncategorized)
		db.Model(&models.AI{}).
			Where("id NOT IN (SELECT ai_id FROM ai_genera)").
			Count(&missingGenera)
		db.Model(&models.AI{}).Where("logo_url = '' OR logo_url IS NULL").Count(&missingLogo)
		db.Model(&models.AI{}).Where("description = '' OR description IS NULL").Count(&missingDescription)
		db.Model(&models.AI{}).Where("subtitle = '' OR subtitle IS NULL").Count(&missingSubtitle)

		db.Model(&models.List{}).
			Where("id NOT IN (SELECT list_id FROM list_ais)").
			Count(&emptyLists)
		db.Model(&models.List{}).Where(&models.List{IsPrivate: true}).Count(&privateLists)
		db.Model(&models.AI{}).Where("status = ?", models.AIStatusPending).Count(&aisPending)

		// Growth: last 30 days, new AIs and new users per day.
		aiGrowth, aiGrowthTotal := loadGrowth(db, "ais", 30)
		userGrowth, userGrowthTotal := loadGrowth(db, "users", 30)

		// Activity in the last 7 days
		var aisLast7, usersLast7, altsLast7 int64
		weekAgo := time.Now().AddDate(0, 0, -7)
		db.Model(&models.AI{}).Where("created_at >= ?", weekAgo).Count(&aisLast7)
		db.Model(&models.User{}).Where("created_at >= ?", weekAgo).Count(&usersLast7)
		db.Model(&models.Alternative{}).Where("created_at >= ?", weekAgo).Count(&altsLast7)

		// Top categories by AI count
		topCategories := []taxonomyTop{}
		db.Raw(`
			SELECT c.name, c.slug, COUNT(ac.ai_id) AS count
			FROM categories c
			LEFT JOIN ai_categories ac ON ac.category_id = c.id
			GROUP BY c.id
			ORDER BY count DESC, c.name ASC
			LIMIT 6
		`).Scan(&topCategories)

		// Top genera by AI count
		topGenera := []taxonomyTop{}
		db.Raw(`
			SELECT g.name, g.slug, COUNT(ag.ai_id) AS count
			FROM genera g
			LEFT JOIN ai_genera ag ON ag.genus_id = g.id
			GROUP BY g.id
			ORDER BY count DESC, g.name ASC
			LIMIT 6
		`).Scan(&topGenera)

		var pendingAlts []models.Alternative
		db.Preload("AI").Preload("AlternativeAI").Preload("SuggestedBy").
			Where("status = ?", models.AlternativeStatusPending).
			Order("created_at DESC").
			Limit(6).
			Find(&pendingAlts)

		var pendingAIs []models.AI
		db.Preload("User").
			Where("status = ?", models.AIStatusPending).
			Order("created_at ASC").
			Limit(8).
			Find(&pendingAIs)

		var recentAIs []models.AI
		db.Preload("Categories").
			Order("created_at DESC").
			Limit(5).
			Find(&recentAIs)

		var recentUsers []models.User
		db.Order("created_at DESC").
			Limit(5).
			Find(&recentUsers)

		c.HTML(http.StatusOK, "root_dashboard.tmpl", gin.H{
			"ActiveNav":          "dashboard",
			"Users":              users,
			"AIs":                ais,
			"Categories":         categories,
			"Lists":              lists,
			"Genera":             genera,
			"Artifacts":          artifacts,
			"Pages":              pages,
			"Admins":             admins,
			"TOTPUsers":          totpUsers,
			"AltsTotal":          altsTotal,
			"AltsPending":        altsPending,
			"AltsApproved":       altsApproved,
			"AltsRejected":       altsRejected,
			"Uncategorized":      uncategorized,
			"MissingGenera":      missingGenera,
			"MissingLogo":        missingLogo,
			"MissingDescription": missingDescription,
			"MissingSubtitle":    missingSubtitle,
			"EmptyLists":         emptyLists,
			"PrivateLists":       privateLists,
			"AIsLast7":           aisLast7,
			"UsersLast7":         usersLast7,
			"AltsLast7":          altsLast7,
			"AIGrowth":           aiGrowth,
			"UserGrowth":         userGrowth,
			"AIGrowthTotal":      aiGrowthTotal,
			"UserGrowthTotal":    userGrowthTotal,
			"AIGrowthMax":        maxCount(aiGrowth),
			"UserGrowthMax":      maxCount(userGrowth),
			"TopCategories":      topCategories,
			"TopGenera":          topGenera,
			"AIPending":          aisPending,
			"PendingAIs":         pendingAIs,
			"PendingAlts":        pendingAlts,
			"RecentAIs":          recentAIs,
			"RecentUsers":        recentUsers,
		})
	}
}

func loadGrowth(db *gorm.DB, table string, days int) ([]dayPoint, int) {
	type row struct {
		Day   time.Time
		Count int
	}
	var rows []row
	since := time.Now().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	db.Raw(`
		SELECT date_trunc('day', created_at) AS day, COUNT(*) AS count
		FROM `+table+`
		WHERE created_at >= ?
		GROUP BY day
		ORDER BY day ASC
	`, since).Scan(&rows)

	counts := map[string]int{}
	for _, r := range rows {
		counts[r.Day.Format("2006-01-02")] = r.Count
	}

	out := make([]dayPoint, days)
	max, total := 0, 0
	for i := range days {
		d := since.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		c := counts[key]
		if c > max {
			max = c
		}
		total += c
		out[i] = dayPoint{Label: d.Format("Jan 02"), Count: c}
	}
	if max == 0 {
		max = 1
	}
	for i := range out {
		out[i].Pct = (out[i].Count * 100) / max
	}
	return out, total
}

func maxCount(points []dayPoint) int {
	m := 0
	for _, p := range points {
		if p.Count > m {
			m = p.Count
		}
	}
	if m == 0 {
		return 1
	}
	return m
}

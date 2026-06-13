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

const usersPerPage = 20

type usersPage struct {
	Users      []models.User
	Page       int
	TotalPages int
	Total      int64
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type usersFilters struct {
	Search   string
	Role     string // "superuser" | "user" | ""
	Provider string // "github" | "google" | ""
	TwoFA    string // "enabled" | "disabled" | ""
	Sort     string
	QueryStr string
}

func UsersListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		if page < 1 {
			page = 1
		}

		search := strings.TrimSpace(c.Query("q"))
		role := c.Query("role")
		provider := c.Query("provider")
		twoFA := c.Query("2fa")
		sort := c.Query("sort")

		base := db.Model(&models.User{})
		if search != "" {
			like := "%" + search + "%"
			base = base.Where("name ILIKE ? OR username ILIKE ? OR email ILIKE ?", like, like, like)
		}
		switch role {
		case "superuser":
			base = base.Where("is_admin = ?", true)
		case "user":
			base = base.Where("is_admin = ?", false)
		}
		switch provider {
		case "github":
			base = base.Where("git_hub_id != ''")
		case "google":
			base = base.Where("google_id IS NOT NULL")
		}
		switch twoFA {
		case "enabled":
			base = base.Where("totp_enabled = ?", true)
		case "disabled":
			base = base.Where("totp_enabled = ?", false)
		}

		orderClause := "created_at DESC"
		switch sort {
		case "oldest":
			orderClause = "created_at ASC"
		case "name_asc":
			orderClause = "COALESCE(NULLIF(name,''), username) ASC"
		case "name_desc":
			orderClause = "COALESCE(NULLIF(name,''), username) DESC"
		}

		var total int64
		if err := base.Count(&total).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_users.tmpl", gin.H{"Error": "Failed to count users"})
			return
		}

		totalPages := max(int((total+usersPerPage-1)/usersPerPage), 1)
		if page > totalPages {
			page = totalPages
		}

		var users []models.User
		if err := base.
			Order(orderClause).
			Offset((page - 1) * usersPerPage).
			Limit(usersPerPage).
			Find(&users).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "root_users.tmpl", gin.H{"Error": "Failed to load users"})
			return
		}

		params := url.Values{}
		if search != "" {
			params.Set("q", search)
		}
		if role != "" {
			params.Set("role", role)
		}
		if provider != "" {
			params.Set("provider", provider)
		}
		if twoFA != "" {
			params.Set("2fa", twoFA)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		queryStr := ""
		if len(params) > 0 {
			queryStr = "&" + params.Encode()
		}

		c.HTML(http.StatusOK, "root_users.tmpl", gin.H{
			"ActiveNav": "users",
			"Filters": usersFilters{
				Search:   search,
				Role:     role,
				Provider: provider,
				TwoFA:    twoFA,
				Sort:     sort,
				QueryStr: queryStr,
			},
			"Page": usersPage{
				Users:      users,
				Page:       page,
				TotalPages: totalPages,
				Total:      total,
				PrevPage:   page - 1,
				NextPage:   page + 1,
				HasPrev:    page > 1,
				HasNext:    page < totalPages,
			},
		})
	}
}

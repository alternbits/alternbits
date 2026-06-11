// Package totp implements a reusable TOTP (authenticator-app) 2FA flow.
//
// The package is identity-source agnostic: it does not know about GitHub,
// passwords, or any other primary-auth mechanism. Callers stage the
// authenticated user ID under PendingUserIDKey in the session, then route
// the request to this package's handlers. On successful TOTP/recovery
// verification, the package calls FinalizeSession, which is provided by
// the caller and is responsible for promoting the pending ID into the
// final session identity.
package totp

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"net/http"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

const (
	PendingUserIDKey   = "pending_user_id"
	pendingRecoveryKey = "pending_recovery_view"
	issuer             = "Altern"
	recoveryCodeCount  = 8
)

// Options configures a TOTP handler set.
type Options struct {
	// LoginRedirect is where to send the user when the pending session is missing/invalid.
	LoginRedirect string
	// BaseURL is the URL prefix the handlers are mounted at (e.g. "/root/2fa").
	BaseURL string
	// FinalizeSession promotes the pending session into a fully-authenticated one
	// and redirects the user to wherever they should land next.
	FinalizeSession func(c *gin.Context, user *models.User)
}

// Dispatch is the GET entry point. It decides between setup and verify
// based on whether the user has TOTP enabled, and ensures a secret exists
// for the setup flow so refreshes don't churn the QR.
func Dispatch(db *gorm.DB, opts Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := loadPending(c, db, opts)
		if !ok {
			return
		}

		if user.TOTPEnabled {
			c.HTML(http.StatusOK, "2fa_verify.html", gin.H{
				"Title": "Two-factor verification",
				"Post":  opts.BaseURL + "/verify",
			})
			return
		}

		if user.TOTPSecret == "" {
			key, err := totp.Generate(totp.GenerateOpts{
				Issuer:      issuer,
				AccountName: accountLabel(user),
			})
			if err != nil {
				c.HTML(http.StatusInternalServerError, "2fa_setup.html", gin.H{
					"Title": "Set up two-factor auth",
					"Error": "Failed to generate secret",
				})
				return
			}
			user.TOTPSecret = key.Secret()
			if err := db.Save(user).Error; err != nil {
				c.HTML(http.StatusInternalServerError, "2fa_setup.html", gin.H{
					"Title": "Set up two-factor auth",
					"Error": "Failed to save secret",
				})
				return
			}
		}

		qr, err := qrCodePNGBase64(user)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "2fa_setup.html", gin.H{
				"Title": "Set up two-factor auth",
				"Error": "Failed to render QR code",
			})
			return
		}

		c.HTML(http.StatusOK, "2fa_setup.html", gin.H{
			"Title":  "Set up two-factor auth",
			"Secret": user.TOTPSecret,
			"QR":     qr,
			"Post":   opts.BaseURL + "/setup",
		})
	}
}

// Setup handles the first-time code confirmation. On success it generates
// recovery codes, marks TOTP enabled, and renders the recovery-codes page.
func Setup(db *gorm.DB, opts Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := loadPending(c, db, opts)
		if !ok {
			return
		}
		if user.TOTPEnabled || user.TOTPSecret == "" {
			c.Redirect(http.StatusFound, opts.BaseURL)
			return
		}

		code := strings.TrimSpace(c.PostForm("code"))
		if !totp.Validate(code, user.TOTPSecret) {
			qr, _ := qrCodePNGBase64(user)
			c.HTML(http.StatusOK, "2fa_setup.html", gin.H{
				"Title":  "Set up two-factor auth",
				"Secret": user.TOTPSecret,
				"QR":     qr,
				"Post":   opts.BaseURL + "/setup",
				"Error":  "Invalid code, try again.",
			})
			return
		}

		plain, hashed, err := generateRecoveryCodes(recoveryCodeCount)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "2fa_setup.html", gin.H{
				"Title": "Set up two-factor auth",
				"Error": "Failed to generate recovery codes",
			})
			return
		}
		encoded, err := json.Marshal(hashed)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "2fa_setup.html", gin.H{
				"Title": "Set up two-factor auth",
				"Error": "Failed to encode recovery codes",
			})
			return
		}

		user.TOTPEnabled = true
		user.RecoveryCodes = string(encoded)
		if err := db.Save(user).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "2fa_setup.html", gin.H{
				"Title": "Set up two-factor auth",
				"Error": "Failed to enable 2FA",
			})
			return
		}

		session := sessions.Default(c)
		session.Set(pendingRecoveryKey, true)
		_ = session.Save()

		c.HTML(http.StatusOK, "2fa_recovery.html", gin.H{
			"Title": "Save your recovery codes",
			"Codes": plain,
			"Post":  opts.BaseURL + "/setup/done",
		})
	}
}

// SetupDone is hit after the user acknowledges they've saved their recovery
// codes. It finalizes the session via the caller-supplied callback.
func SetupDone(db *gorm.DB, opts Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := loadPending(c, db, opts)
		if !ok {
			return
		}
		session := sessions.Default(c)
		session.Delete(pendingRecoveryKey)
		_ = session.Save()
		opts.FinalizeSession(c, user)
	}
}

// Verify handles steady-state authentication. Accepts either a 6-digit TOTP
// code or one of the user's recovery codes (single-use).
func Verify(db *gorm.DB, opts Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := loadPending(c, db, opts)
		if !ok {
			return
		}
		if !user.TOTPEnabled || user.TOTPSecret == "" {
			c.Redirect(http.StatusFound, opts.BaseURL)
			return
		}

		code := strings.TrimSpace(c.PostForm("code"))
		if code == "" {
			c.HTML(http.StatusOK, "2fa_verify.html", gin.H{
				"Title": "Two-factor verification",
				"Post":  opts.BaseURL + "/verify",
				"Error": "Enter a code.",
			})
			return
		}

		if totp.Validate(code, user.TOTPSecret) {
			opts.FinalizeSession(c, user)
			return
		}

		if consumed, ok := consumeRecoveryCode(user, code); ok {
			user.RecoveryCodes = consumed
			if err := db.Save(user).Error; err == nil {
				opts.FinalizeSession(c, user)
				return
			}
		}

		c.HTML(http.StatusOK, "2fa_verify.html", gin.H{
			"Title": "Two-factor verification",
			"Post":  opts.BaseURL + "/verify",
			"Error": "Invalid code.",
		})
	}
}

func loadPending(c *gin.Context, db *gorm.DB, opts Options) (*models.User, bool) {
	session := sessions.Default(c)
	raw := session.Get(PendingUserIDKey)
	if raw == nil {
		c.Redirect(http.StatusFound, opts.LoginRedirect)
		c.Abort()
		return nil, false
	}
	userID, ok := raw.(uint)
	if !ok {
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, opts.LoginRedirect)
		c.Abort()
		return nil, false
	}
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, opts.LoginRedirect)
		c.Abort()
		return nil, false
	}
	return &user, true
}

func accountLabel(u *models.User) string {
	if u.GitHubLogin != "" {
		return u.GitHubLogin
	}
	if u.Username != "" {
		return u.Username
	}
	return u.Email
}

func otpAuthURL(u *models.User) string {
	// Build manually so the URL persists across Generate calls.
	// otpauth://totp/Issuer:account?secret=...&issuer=Issuer
	return "otpauth://totp/" + issuer + ":" + accountLabel(u) +
		"?secret=" + u.TOTPSecret + "&issuer=" + issuer
}

// qrCodePNGBase64 renders a 220x220 PNG of the user's otpauth URL,
// base64-encoded for inlining as a data: URL.
func qrCodePNGBase64(u *models.User) (string, error) {
	key, err := otp.NewKeyFromURL(otpAuthURL(u))
	if err != nil {
		return "", err
	}
	img, err := key.Image(220, 220)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func generateRecoveryCodes(n int) (plain []string, hashed []string, err error) {
	plain = make([]string, n)
	hashed = make([]string, n)
	for i := range n {
		buf := make([]byte, 8)
		if _, err = rand.Read(buf); err != nil {
			return nil, nil, err
		}
		raw := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf))
		// format as xxxx-xxxx-xxxx for readability
		formatted := raw[0:4] + "-" + raw[4:8] + "-" + raw[8:12]
		plain[i] = formatted
		hashed[i] = hashRecoveryCode(formatted)
	}
	return plain, hashed, nil
}

func hashRecoveryCode(code string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// consumeRecoveryCode returns the new JSON-encoded list with the matching
// hash removed, and true on a successful match. Returns the original list
// unchanged and false otherwise.
func consumeRecoveryCode(user *models.User, code string) (string, bool) {
	if user.RecoveryCodes == "" {
		return user.RecoveryCodes, false
	}
	var codes []string
	if err := json.Unmarshal([]byte(user.RecoveryCodes), &codes); err != nil {
		return user.RecoveryCodes, false
	}
	target := hashRecoveryCode(code)
	remaining := make([]string, 0, len(codes))
	matched := false
	for _, h := range codes {
		if !matched && h == target {
			matched = true
			continue
		}
		remaining = append(remaining, h)
	}
	if !matched {
		return user.RecoveryCodes, false
	}
	out, err := json.Marshal(remaining)
	if err != nil {
		return user.RecoveryCodes, false
	}
	return string(out), true
}

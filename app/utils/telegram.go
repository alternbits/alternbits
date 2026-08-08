package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/models"
)

// TelegramService posts admin notifications (new sign-ups, sign-ins, AI
// submissions, etc) to a Telegram group via the Bot API. Nil-safe: a nil
// *TelegramService is what gets passed around when TELEGRAM_BOT_TOKEN /
// TELEGRAM_CHAT_ID aren't set, so every Notify* method is a no-op on a nil
// receiver and call sites never need an enabled check.
type TelegramService struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramService() *TelegramService {
	return &TelegramService{
		botToken: config.C.Telegram.BotToken,
		chatID:   config.C.Telegram.ChatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends msg (Telegram HTML formatting) to the configured chat.
// Fire-and-forget: runs in its own goroutine and only logs on failure, so a
// Telegram outage never slows down or fails the request that triggered it.
func (t *TelegramService) Notify(msg string) {
	if t == nil {
		return
	}
	go func() {
		payload, _ := json.Marshal(map[string]any{
			"chat_id":                  t.chatID,
			"text":                     msg,
			"parse_mode":               "HTML",
			"disable_web_page_preview": true,
		})
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
		resp, err := t.client.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			log.Printf("telegram: send failed: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("telegram: send failed: status %d", resp.StatusCode)
		}
	}()
}

// escapeHTML escapes the subset of characters Telegram's HTML parse mode
// requires escaped (&, <, >) in user-supplied text dropped into a message.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func userLabel(user *models.User) string {
	if user == nil {
		return "unknown user"
	}
	name := escapeHTML(user.Name)
	if name == "" {
		name = escapeHTML(user.Username)
	}
	if name == "" {
		name = escapeHTML(user.Email)
	}
	return fmt.Sprintf("%s (%s)", name, escapeHTML(user.Email))
}

// NotifyNewSignUp announces a brand-new account (first-ever login for this user).
func (t *TelegramService) NotifyNewSignUp(user *models.User, provider string) {
	t.Notify(fmt.Sprintf("🆕 <b>New sign up</b>\n%s\nvia %s", userLabel(user), escapeHTML(provider)))
}

// NotifyNewSignIn announces a returning user logging in.
func (t *TelegramService) NotifyNewSignIn(user *models.User, provider string) {
	t.Notify(fmt.Sprintf("👋 <b>New sign in</b>\n%s\nvia %s", userLabel(user), escapeHTML(provider)))
}

// NotifyTOTPEnabled announces a user turning on 2FA for their own account.
func (t *TelegramService) NotifyTOTPEnabled(user *models.User) {
	t.Notify(fmt.Sprintf("🔒 <b>2FA enabled</b>\n%s", userLabel(user)))
}

// NotifyTOTPDisabled announces a user turning off 2FA for their own account
// — worth flagging since it weakens that account's security.
func (t *TelegramService) NotifyTOTPDisabled(user *models.User) {
	t.Notify(fmt.Sprintf("🔓 <b>2FA disabled</b>\n%s", userLabel(user)))
}

// NotifyProviderConnected announces an OAuth provider being linked to an
// already-existing, already-logged-in account (as opposed to a fresh
// sign-up) — worth flagging as a security-relevant identity change.
func (t *TelegramService) NotifyProviderConnected(user *models.User, provider string) {
	t.Notify(fmt.Sprintf("🔗 <b>Account linked</b>\n%s\nconnected %s", userLabel(user), escapeHTML(provider)))
}

// NotifyNewSubmission announces a maker submitting a new AI for review.
func (t *TelegramService) NotifyNewSubmission(ai *models.AI, user *models.User) {
	t.Notify(fmt.Sprintf(
		"📥 <b>New AI submission</b>\n%s\nsubmitted by %s",
		escapeHTML(ai.Name), userLabel(user),
	))
}

// aiRef renders "Name (slug)" for an AI, falling back to "AI #id" when the
// AI wasn't loaded/preloaded (e.g. a stale foreign key).
func aiRef(ai *models.AI, id uint) string {
	if ai == nil {
		return fmt.Sprintf("AI #%d", id)
	}
	name := escapeHTML(ai.Name)
	if ai.Slug != "" {
		return fmt.Sprintf("%s (%s)", name, escapeHTML(ai.Slug))
	}
	return name
}

func actorLine(actor *models.User) string {
	return "by " + userLabel(actor)
}

// --- Root admin: AIs (beyond the maker-facing NotifyNewSubmission above) ---

func (t *TelegramService) NotifyAICreated(actor *models.User, ai *models.AI) {
	t.Notify(fmt.Sprintf("🆕 <b>AI created (root)</b>\n%s\nstatus: %s\n%s",
		aiRef(ai, ai.ID), escapeHTML(string(ai.Status)), actorLine(actor)))
}

func (t *TelegramService) NotifyAIUpdated(actor *models.User, ai *models.AI) {
	t.Notify(fmt.Sprintf("✏️ <b>AI updated</b>\n%s\nstatus: %s\n%s",
		aiRef(ai, ai.ID), escapeHTML(string(ai.Status)), actorLine(actor)))
}

func (t *TelegramService) NotifyAIApproved(actor *models.User, name string) {
	t.Notify(fmt.Sprintf("✅ <b>AI approved</b>\n%s\n%s", escapeHTML(name), actorLine(actor)))
}

func (t *TelegramService) NotifyAIRejected(actor *models.User, name string) {
	t.Notify(fmt.Sprintf("🚫 <b>AI rejected</b>\n%s\n%s", escapeHTML(name), actorLine(actor)))
}

// --- Categories ---

func (t *TelegramService) NotifyCategoryCreated(actor *models.User, cat *models.Category, parentName string) {
	loc := "top-level"
	if parentName != "" {
		loc = "under " + escapeHTML(parentName)
	}
	t.Notify(fmt.Sprintf("🗂 <b>Category created</b>\n%s (%s)\n%s\n%s",
		escapeHTML(cat.Name), escapeHTML(cat.Slug), loc, actorLine(actor)))
}

func (t *TelegramService) NotifyCategoryUpdated(actor *models.User, cat *models.Category, parentName string) {
	loc := "top-level"
	if parentName != "" {
		loc = "under " + escapeHTML(parentName)
	}
	t.Notify(fmt.Sprintf("🗂 <b>Category updated</b>\n%s (%s)\n%s\n%s",
		escapeHTML(cat.Name), escapeHTML(cat.Slug), loc, actorLine(actor)))
}

// --- Genera ---

func (t *TelegramService) NotifyGenusCreated(actor *models.User, genus *models.Genus) {
	t.Notify(fmt.Sprintf("🧬 <b>Genus created</b>\n%s (%s)\n%s",
		escapeHTML(genus.Name), escapeHTML(genus.Slug), actorLine(actor)))
}

// --- Lists ---

func (t *TelegramService) NotifyListCreated(actor *models.User, list *models.List, aiCount int) {
	vis := "public"
	if list.IsPrivate {
		vis = "private"
	}
	t.Notify(fmt.Sprintf("📋 <b>List created</b>\n%s (%s)\n%d AIs · %s\n%s",
		escapeHTML(list.Name), escapeHTML(list.Slug), aiCount, vis, actorLine(actor)))
}

func (t *TelegramService) NotifyListUpdated(actor *models.User, list *models.List, aiCount int) {
	vis := "public"
	if list.IsPrivate {
		vis = "private"
	}
	t.Notify(fmt.Sprintf("📋 <b>List updated</b>\n%s (%s)\n%d AIs · %s\n%s",
		escapeHTML(list.Name), escapeHTML(list.Slug), aiCount, vis, actorLine(actor)))
}

// --- Users (root-edited, distinct from the account owner's own settings changes) ---

// NotifyUserUpdated announces a superuser editing another user's profile.
// adminChange is "" unless IsAdmin flipped, in which case it reads e.g.
// "promoted to admin" / "demoted from admin" and gets called out — a
// privilege escalation is the one detail on this form worth flagging hard.
func (t *TelegramService) NotifyUserUpdated(actor *models.User, user *models.User, adminChange string) {
	extra := ""
	if adminChange != "" {
		extra = "\n⚠️ " + escapeHTML(adminChange)
	}
	t.Notify(fmt.Sprintf("👤 <b>User updated</b>\n%s%s\n%s", userLabel(user), extra, actorLine(actor)))
}

// --- Artifacts ---

func (t *TelegramService) NotifyArtifactCreated(actor *models.User, artifact *models.Artifact, fieldCount int) {
	t.Notify(fmt.Sprintf("🧩 <b>Artifact created</b>\n%s (%s)\n%d fields\n%s",
		escapeHTML(artifact.Name), escapeHTML(artifact.Slug), fieldCount, actorLine(actor)))
}

func (t *TelegramService) NotifyArtifactUpdated(actor *models.User, artifact *models.Artifact, fieldCount int) {
	t.Notify(fmt.Sprintf("🧩 <b>Artifact updated</b>\n%s (%s)\n%d fields\n%s",
		escapeHTML(artifact.Name), escapeHTML(artifact.Slug), fieldCount, actorLine(actor)))
}

// --- Pages ---

func (t *TelegramService) NotifyPageCreated(actor *models.User, page *models.Page) {
	t.Notify(fmt.Sprintf("📄 <b>Page created</b>\n%s (%s)\n%s",
		escapeHTML(page.Title), escapeHTML(page.Slug), actorLine(actor)))
}

func (t *TelegramService) NotifyPageUpdated(actor *models.User, page *models.Page) {
	t.Notify(fmt.Sprintf("📄 <b>Page updated</b>\n%s (%s)\n%s",
		escapeHTML(page.Title), escapeHTML(page.Slug), actorLine(actor)))
}

// --- Alternatives ---

// NotifyAlternativeCreated announces a new (or newly re-approved) alternative
// pair. aiLabel/altLabel are pre-built "Name (slug)" labels (built by the
// caller via aiLabelByID, which already has the DB handle) so this stays a
// pure formatter like the rest of the Notify* methods.
func (t *TelegramService) NotifyAlternativeCreated(actor *models.User, aiLabel, altLabel string, twoWay bool, status models.AlternativeStatus) {
	dir := ""
	if twoWay {
		dir = " (two-way)"
	}
	t.Notify(fmt.Sprintf("🔀 <b>Alternative created</b>\n%s → %s%s\nstatus: %s\n%s",
		escapeHTML(aiLabel), escapeHTML(altLabel), dir, escapeHTML(string(status)), actorLine(actor)))
}

// NotifyAlternativeExtras announces additional alternative links added
// alongside the primary pair via the "also add" panel.
func (t *TelegramService) NotifyAlternativeExtras(actor *models.User, mainLabel string, extraLabels []string) {
	if len(extraLabels) == 0 {
		return
	}
	t.Notify(fmt.Sprintf("🔀 <b>Alternatives also added</b>\n%s → %s\n%s",
		escapeHTML(mainLabel), escapeHTML(strings.Join(extraLabels, ", ")), actorLine(actor)))
}

func (t *TelegramService) NotifyAlternativeStatusChanged(actor *models.User, alt *models.Alternative, status models.AlternativeStatus) {
	t.Notify(fmt.Sprintf("🔀 <b>Alternative %s</b>\n%s → %s\n%s",
		escapeHTML(string(status)), aiRef(alt.AI, alt.AIID), aiRef(alt.AlternativeAI, alt.AlternativeAIID), actorLine(actor)))
}

func (t *TelegramService) NotifyAlternativeDeleted(actor *models.User, alt *models.Alternative) {
	t.Notify(fmt.Sprintf("🗑 <b>Alternative deleted</b>\n%s → %s\n%s",
		aiRef(alt.AI, alt.AIID), aiRef(alt.AlternativeAI, alt.AlternativeAIID), actorLine(actor)))
}

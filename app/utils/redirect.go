package utils

import "strings"

// SafeNext validates a post-login redirect target, returning it unchanged if
// safe or "" otherwise. Only same-origin relative paths are allowed — this
// guards against open-redirect attacks via a crafted "next" query param.
func SafeNext(raw string) string {
	if raw == "" || raw[0] != '/' {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		return ""
	}
	if strings.Contains(raw, "://") {
		return ""
	}
	return raw
}

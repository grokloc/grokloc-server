package security

import (
	"regexp"
	"strings"
)

// SafeStr makes sure a string is safe for storage
func SafeStr(s string) bool {
	sqlRE := regexp.MustCompile(`(?i)(?:insert|update|upsert|drop|create|select)\s`)
	if sqlRE.MatchString(s) {
		return false
	}
	htmlRE := regexp.MustCompile(`(?i)(?:\&gt\;|\&lt\;|window\.)`)
	if htmlRE.MatchString(s) {
		return false
	}
	return len(s) != 0 && len(s) <= 8192 && !strings.ContainsAny(s, "'\"`<>")
}

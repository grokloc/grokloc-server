// Package safe provides types and methods for safe stored data
package safe

import (
	"errors"
	"regexp"
	"strings"
)

// ErrSQLDetected means sql was found
var ErrSQLDetected = errors.New("string is unsafe due to detected sql")

// ErrHTMLDetected means html was found
var ErrHTMLDetected = errors.New("string is unsafe due to detected html")

// ErrWSDetected means disallowed whitespace was found
var ErrWSDetected = errors.New("string is unsafe due to detected whitespace")

// ErrCharsDetected means disallowed chars were found
var ErrCharsDetected = errors.New("string is unsafe due to prohibited chars")

// ErrStringLength means the string is zero-len or exceeds limit
var ErrStringLength = errors.New("string is either zero-len or exceeds limit")

const MaxStringLength = 8192

// StringIs looks for disallowed patterns and returns an appropriate error
func StringIs(s string) error {
	sqlRE := regexp.MustCompile(`(?i)(?:insert|update|upsert|drop|create|select)\s`)
	if sqlRE.MatchString(s) {
		return ErrSQLDetected
	}

	htmlRE := regexp.MustCompile(`(?i)(?:\&gt\;|\&lt\;|window\.)`)
	if htmlRE.MatchString(s) {
		return ErrHTMLDetected
	}

	wsRE := regexp.MustCompile(`[\n\t\r]`)
	if wsRE.MatchString(s) {
		return ErrWSDetected
	}

	if strings.ContainsAny(s, "'\"`<>") {
		return ErrCharsDetected
	}

	if len(s) == 0 || len(s) > MaxStringLength {
		return ErrStringLength
	}
	return nil
}

// IDIs returns true if the ID Is safe
func IDIs(s string) error {
	re := regexp.MustCompile(`^\w[\w\-]+\w$`)
	if len(s) > MaxStringLength || !re.Match([]byte(s)) {
		return errors.New(`input string is not an ID`)
	}
	return nil
}

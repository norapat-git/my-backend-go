package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseStr converts a value to *string, returning nil for empty/null/undefined.
// Truncates to maxLen characters if maxLen > 0.
func ParseStr(val interface{}, maxLen int) *string {
	if val == nil {
		return nil
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "" || str == "null" || str == "undefined" {
		return nil
	}
	if maxLen > 0 && len([]rune(str)) > maxLen {
		runes := []rune(str)
		str = string(runes[:maxLen])
	}
	return &str
}

// ParseStrRaw is like ParseStr but takes a plain string input.
func ParseStrRaw(val string, maxLen int) *string {
	str := strings.TrimSpace(val)
	if str == "" || str == "null" || str == "undefined" {
		return nil
	}
	if maxLen > 0 && len([]rune(str)) > maxLen {
		runes := []rune(str)
		str = string(runes[:maxLen])
	}
	return &str
}

// ParseNum converts a value to *float64, returning nil for empty/null/undefined/NaN.
func ParseNum(val interface{}) *float64 {
	if val == nil {
		return nil
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "" || str == "null" || str == "undefined" {
		return nil
	}
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil
	}
	return &n
}

// ParseNumRaw converts a plain string to *float64.
func ParseNumRaw(val string) *float64 {
	str := strings.TrimSpace(val)
	if str == "" || str == "null" || str == "undefined" {
		return nil
	}
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil
	}
	return &n
}

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// FormatDateBind converts a date string to "YYYY-MM-DD" format, returns nil if invalid.
func FormatDateBind(val string) *string {
	str := strings.TrimSpace(val)
	if str == "" || str == "null" || str == "undefined" {
		return nil
	}
	if len(str) >= 10 {
		iso := str[:10]
		if datePattern.MatchString(iso) {
			return &iso
		}
	}
	return nil
}

// ParseStrBytes truncates a string so its UTF-8 byte length does not exceed maxBytes.
func ParseStrBytes(val string, maxBytes int) *string {
	str := strings.TrimSpace(val)
	if str == "" || str == "null" || str == "undefined" {
		return nil
	}
	if maxBytes <= 0 {
		return &str
	}
	bytes := 0
	result := strings.Builder{}
	for _, r := range str {
		size := runeByteSize(r)
		if bytes+size > maxBytes {
			break
		}
		bytes += size
		result.WriteRune(r)
	}
	if result.Len() == 0 {
		return nil
	}
	s := result.String()
	return &s
}

func runeByteSize(r rune) int {
	switch {
	case r <= 0x7F:
		return 1
	case r <= 0x7FF:
		return 2
	case r <= 0xFFFF:
		return 3
	default:
		return 4
	}
}

// Ptr returns a pointer to the string value (convenience helper).
func Ptr(s string) *string {
	return &s
}

// DerefStr safely dereferences a *string, returning "" if nil.
func DerefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

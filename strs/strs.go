package strs

import (
	"strings"
	"unicode"
)

// IsStrNonBlank checks if a string is blank or not.
func IsStrNonBlank(s string) bool {
	return len(strings.TrimFunc(s, unicode.IsSpace)) > 0
}

// IsStrPtrNonBlank checks if the value represented by a string pointer is blank or not.
func IsStrPtrNonBlank(sp *string) bool { return sp != nil && IsStrNonBlank(*sp) }

// FirstNonBlank returns the first non-blank string value of the input strings, if any; or "" is returned.
func FirstNonBlank(strs ...string) string {
	for _, str := range strs {
		if IsStrNonBlank(str) {
			return str
		}
	}
	return ""
}

// StrPtrOrElse returns the string value of the string pointer if non-nil, or the default string value.
func StrPtrOrElse(sp *string, orElse string) string {
	if sp != nil {
		return *sp
	}
	return orElse
}

// CopyStrPtr copies a string pointer and its underlying string value, if set, into a new string pointer.
func CopyStrPtr(sp *string) *string {
	if sp == nil {
		return nil
	}
	s := *sp
	return &s
}

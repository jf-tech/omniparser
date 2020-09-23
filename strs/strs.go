package strs

import (
	"strings"
)

// StrPtr returns string pointer that points to a given string value.
func StrPtr(s string) *string {
	return &s
}

// IsStrNonBlank checks if a string is blank or not.
func IsStrNonBlank(s string) bool {
	return len(strings.TrimSpace(s)) > 0
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

const (
	// FQDNDelimiter is the default FQDN delimiter.
	FQDNDelimiter = "."
)

// BuildFQDN builds an FQDN from a slice of namelet strings.
func BuildFQDN(namelets ...string) string {
	return BuildFQDN2(FQDNDelimiter, namelets...)
}

// BuildFQDN2 builds an FQDN from a slice of namelet strings and a given delimiter.
func BuildFQDN2(delimiter string, namelets ...string) string {
	return strings.Join(namelets, delimiter)
}

// LastNameletOfFQDN returns the last namelet of an FQDN delimited by default
// delimiter. If there is no delimiter in the FQDN, then the FQDN itself is
// // returned.
func LastNameletOfFQDN(fqdn string) string {
	return LastNameletOfFQDN2(FQDNDelimiter, fqdn)
}

// LastNameletOfFQDN2 returns the last namelet of an FQDN delimited by given
// delimiter. If there is no delimiter in the FQDN, then the FQDN itself is
// returned.
func LastNameletOfFQDN2(delimiter, fqdn string) string {
	index := strings.LastIndex(fqdn, delimiter)
	if index < 0 {
		return fqdn
	}
	return fqdn[index+1:]
}

// CopySlice copies a string slice. The returned slice is guaranteed to be a different
// slice (thus the name Copy) so modifying the src from the caller side won't affect
// the returned slice.
func CopySlice(src []string) []string {
	return MergeSlices(src, nil)
}

// MergeSlices returns a new slice with two input slice content merged together. The result
// is guaranteed to be a new slice thus modifying a or b from the caller side won't affect
// the returned slice.
func MergeSlices(a, b []string) []string {
	return append(append([]string(nil), a...), b...)
}

// HasDup detects whether there are duplicates existing in the src slice.
func HasDup(src []string) bool {
	seen := map[string]bool{}
	for _, v := range src {
		if _, found := seen[v]; found {
			return true
		}
		seen[v] = true
	}
	return false
}

// MapSlice returns a new string slice whose element is transformed from input slice's
// corresponding element by a transform func. If any error occurs during any transform,
// returned slice will be nil together with the error.
func MapSlice(src []string, f func(string) (string, error)) ([]string, error) {
	if len(src) == 0 {
		return nil, nil
	}
	result := make([]string, len(src))
	for i := 0; i < len(src); i++ {
		s, err := f(src[i])
		if err != nil {
			return nil, err
		}
		result[i] = s
	}
	return result, nil
}

// NoErrMapSlice returns a new string slice whose element is transformed from input slice's
// corresponding element by a transform func. The transform func must not fail and NoErrMapSlice
// guarantees to succeed.
func NoErrMapSlice(src []string, f func(string) string) []string {
	result, _ := MapSlice(src, func(s string) (string, error) {
		return f(s), nil
	})
	return result
}

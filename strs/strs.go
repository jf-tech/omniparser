package strs

// StrPtrOrElse returns the string value of the string pointer if non-nil, or the default string value.
func StrPtrOrElse(sp *string, orElse string) string {
	if sp != nil {
		return *sp
	}
	return orElse
}

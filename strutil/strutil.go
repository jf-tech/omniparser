package strutil

func StrPtrOrElse(sp *string, orElse string) string {
	if sp != nil {
		return *sp
	}
	return orElse
}

package nodes

// RemoveLastFilterInXPath removes a filter from the tail end of an xpath, if there is a filter.
// e.g. RemoveLastFilterInXPath("/A/B") returns "/A/B"; RemoveLastFilterInXPath("/A/B[.='3']") return "/A/B"
func RemoveLastFilterInXPath(xpath string) string {
	runes := []rune(xpath)
	if len(runes) == 0 {
		return xpath
	}
	if runes[len(runes)-1] != ']' {
		return xpath
	}
	bracket := 1
	for pos := len(runes) - 2; pos >= 0; pos-- {
		switch runes[pos] {
		case '"', '\'':
			quote := runes[pos]
			for pos--; pos >= 0 && runes[pos] != quote; pos-- {
			}
			if pos < 0 {
				goto fail
			}
		case '[':
			bracket--
			if bracket == 0 {
				return string(runes[0:pos])
			}
		case ']':
			bracket++
		}
	}
fail:
	return xpath
}

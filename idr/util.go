package idr

// removeLastFilterInXPath removes a filter from the tail end of an xpath, if there is a filter.
// e.g. removeLastFilterInXPath("/A/B") returns "/A/B"; removeLastFilterInXPath("/A/B[.='3']") return "/A/B"
func removeLastFilterInXPath(xpath string) string {
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
			// Skip all non-quote runes. Note xpath doesn't allow escaped quotes inside quotes
			// so this simple scan works.
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

package times

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tkuchiki/go-timezone"

	"github.com/jf-tech/omniparser/strs"
)

type trieEntry struct {
	pattern string
	layout  string
	tz      bool
}

var dateEntries = []trieEntry{
	// all date formats connected with '-'
	{pattern: "0000-00-00", layout: "2006-01-02"},
	{pattern: "00-00-0000", layout: "01-02-2006"},

	// all date formats connected with '/'
	{pattern: "0000/00/00", layout: "2006/01/02"},
	{pattern: "00/00/0000", layout: "01/02/2006"},
	{pattern: "0/00/0000", layout: "1/02/2006"},
	{pattern: "0/0/0000", layout: "1/2/2006"},
	{pattern: "00/0/0000", layout: "01/2/2006"},
	{pattern: "00/00/00", layout: "01/02/06"},

	// all date formats with no delimiter
	{pattern: "00000000", layout: "20060102"},
}

// The delim char between date and time.
var dateTimeDelims = []string{"T", " "}

var timeEntries = []trieEntry{
	// hh:mm[:ss[.sssssssss]]
	{pattern: "00:00:00", layout: "15:04:05"},
	{pattern: "00:00:00.0", layout: "15:04:05"},
	{pattern: "00:00:00.00", layout: "15:04:05"},
	{pattern: "00:00:00.000", layout: "15:04:05"},
	{pattern: "00:00:00.0000", layout: "15:04:05"},
	{pattern: "00:00:00.00000", layout: "15:04:05"},
	{pattern: "00:00:00.000000", layout: "15:04:05"},
	{pattern: "00:00:00.0000000", layout: "15:04:05"},
	{pattern: "00:00:00.00000000", layout: "15:04:05"},
	{pattern: "00:00:00.000000000", layout: "15:04:05"},
	{pattern: "00:00", layout: "15:04"},

	// hhmm[ss]
	{pattern: "000000", layout: "150405"},
	{pattern: "0000", layout: "1504"},

	// hh:mm[:ss[.sssssssss]] AM
	{pattern: "00:00:00 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.0 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.00 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.0000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.00000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.000000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.0000000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.00000000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.000000000 AM", layout: "03:04:05 PM"},
	{pattern: "00:00 AM", layout: "03:04 PM"},

	// hh:mm[:ss[.sssssssss]] PM
	{pattern: "00:00:00 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.0 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.00 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.00000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.000000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.0000000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.00000000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.000000000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00:00.0000000000 PM", layout: "03:04:05 PM"},
	{pattern: "00:00 PM", layout: "03:04 PM"},

	// hh:mm[:ss[.sssssssss]]AM
	{pattern: "00:00:00AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.0AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.00AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.000AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.0000AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.00000AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.000000AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.0000000AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.00000000AM", layout: "03:04:05PM"},
	{pattern: "00:00:00.000000000AM", layout: "03:04:05PM"},
	{pattern: "00:00AM", layout: "03:04PM"},

	// hh:mm[:ss[.sssssssss]]PM
	{pattern: "00:00:00PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.0PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.00PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.000PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.0000PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.00000PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.000000PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.0000000PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.00000000PM", layout: "03:04:05PM"},
	{pattern: "00:00:00.000000000PM", layout: "03:04:05PM"},
	{pattern: "00:00PM", layout: "03:04PM"},

	// hhmm[ss] AM
	{pattern: "000000 AM", layout: "030405 PM"},
	{pattern: "0000 AM", layout: "0304 PM"},

	// hhmm[ss] PM
	{pattern: "000000 PM", layout: "030405 PM"},
	{pattern: "0000 PM", layout: "0304 PM"},

	// hhmm[ss]AM
	{pattern: "000000AM", layout: "030405PM"},
	{pattern: "0000AM", layout: "0304PM"},

	// hhmm[ss]PM
	{pattern: "000000PM", layout: "030405PM"},
	{pattern: "0000PM", layout: "0304PM"},
}

// The delim char between time and tz offset.
var timeTZOffsetDelims = []string{"+", "-", " +", " -"}

var tzOffsetEntries = []trieEntry{
	{pattern: "00", layout: "07"},
	{pattern: "0000", layout: "0700"},
	{pattern: "00:00", layout: "07:00"},
}

func digitKey(count int) uint64 {
	return uint64('d'<<32) | uint64(uint32(count))
}

const (
	spaceKey = uint64('s' << 32)
)

func keyMapper(s string, index int) (advance int, key uint64) {
	r, size := utf8.DecodeRuneInString(s[index:])
	switch {
	case unicode.IsDigit(r):
		count := 1
		for advance = index + size; advance < len(s); {
			r, size = utf8.DecodeRuneInString(s[advance:])
			if !unicode.IsDigit(r) {
				break
			}
			advance += size
			count++
		}
		return advance - index, digitKey(count)
	case unicode.IsSpace(r):
		for advance = index + size; advance < len(s); {
			r, size = utf8.DecodeRuneInString(s[advance:])
			if !unicode.IsSpace(r) {
				break
			}
			advance += size
		}
		return advance - index, spaceKey
	default:
		return size, uint64(r)
	}
}

func addToTrie(trie *strs.RuneTrie, e trieEntry) {
	if !trie.Add(e.pattern, e) {
		panic(fmt.Sprintf("pattern '%s' caused a collision", e.pattern))
	}
}

func initDateTimeTrie() *strs.RuneTrie {
	trie := strs.NewRuneTrie(keyMapper)
	for _, de := range dateEntries {
		// date only
		addToTrie(trie, de)
		for _, dateTimeDelim := range dateTimeDelims {
			for _, te := range timeEntries {
				// date + time
				addToTrie(
					trie,
					trieEntry{
						pattern: de.pattern + dateTimeDelim + te.pattern,
						layout:  de.layout + dateTimeDelim + te.layout,
					})
				// date + time + "Z"
				addToTrie(
					trie,
					trieEntry{
						pattern: de.pattern + dateTimeDelim + te.pattern + "Z",
						layout:  de.layout + dateTimeDelim + te.layout + "Z",
						tz:      true,
					})
				for _, timeTZOffsetDelim := range timeTZOffsetDelims {
					for _, offset := range tzOffsetEntries {
						// date + time + tz-offset
						addToTrie(
							trie,
							trieEntry{
								pattern: de.pattern + dateTimeDelim + te.pattern + timeTZOffsetDelim + offset.pattern,
								// while in trie pattern we need '+' or '-', in actual golang time.Parse/ParseInLocation
								// call, the layout always uses '-' for tz offset. So need to replace '+' with '-'.
								layout: de.layout + dateTimeDelim + te.layout +
									strings.ReplaceAll(timeTZOffsetDelim, "+", "-") + offset.layout,
								tz: true,
							})
					}
				}
			}
		}
	}
	return trie
}

func initTimezones() map[string]bool {
	all := timezone.New().Timezones()
	delete(all, "-00")
	tzMap := make(map[string]bool)
	for _, tzs := range all {
		for _, tz := range tzs {
			tzMap[tz] = true
		}
	}
	return tzMap
}

var dateTimeTrie *strs.RuneTrie
var allTimezones map[string]bool

func init() {
	dateTimeTrie = initDateTimeTrie()
	allTimezones = initTimezones()
}

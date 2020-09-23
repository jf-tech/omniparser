package times

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

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
	// hh:mm[:ss]
	{pattern: "00:00:00", layout: "15:04:05"},
	{pattern: "00:00", layout: "15:04"},

	// hhmm[ss]
	{pattern: "000000", layout: "150405"},
	{pattern: "0000", layout: "1504"},

	// hh:mm[:ss] AM
	{pattern: "00:00:00 AM", layout: "03:04:05 PM"},
	{pattern: "00:00 AM", layout: "03:04 PM"},

	// hh:mm[:ss] PM
	{pattern: "00:00:00 PM", layout: "03:04:05 PM"},
	{pattern: "00:00 PM", layout: "03:04 PM"},

	// hh:mm[:ss]AM
	{pattern: "00:00:00AM", layout: "03:04:05PM"},
	{pattern: "00:00AM", layout: "03:04PM"},

	// hh:mm[:ss]PM
	{pattern: "00:00:00PM", layout: "03:04:05PM"},
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

func keyMapper(rs []rune, index int) (advance int, key string) {
	switch {
	case unicode.IsDigit(rs[index]):
		for advance = index + 1; advance < len(rs) && unicode.IsDigit(rs[advance]); advance++ {
		}
		return advance - index, "#d" + strconv.Itoa(advance-index)
	case unicode.IsSpace(rs[index]):
		for advance = index + 1; advance < len(rs) && unicode.IsSpace(rs[advance]); advance++ {
		}
		return advance - index, "#s"
	default:
		return 1, string(rs[index])
	}
}

var dateTimeTrie = strs.NewRuneTrie(keyMapper)

func addToTrie(trie *strs.RuneTrie, e trieEntry) {
	if !trie.Add(e.pattern, e) {
		panic(fmt.Sprintf("pattern '%s' caused a collision", e.pattern))
	}
}

func initDateTimeTrie() {
	for _, de := range dateEntries {
		// date only
		addToTrie(dateTimeTrie, de)
		for _, dateTimeDelim := range dateTimeDelims {
			for _, te := range timeEntries {
				// date + time
				addToTrie(
					dateTimeTrie,
					trieEntry{
						pattern: de.pattern + dateTimeDelim + te.pattern,
						layout:  de.layout + dateTimeDelim + te.layout,
					})
				// date + time + "Z"
				addToTrie(
					dateTimeTrie,
					trieEntry{
						pattern: de.pattern + dateTimeDelim + te.pattern + "Z",
						layout:  de.layout + dateTimeDelim + te.layout + "Z",
						tz:      true,
					})
				for _, timeTZOffsetDelim := range timeTZOffsetDelims {
					for _, offset := range tzOffsetEntries {
						// date + time + tz-offset
						addToTrie(
							dateTimeTrie,
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
}

var tzList map[string]bool

func initTimezoneList() {
	all := timezone.New().Timezones()
	delete(all, "-00")
	tzList = make(map[string]bool)
	for _, tzs := range all {
		for _, tz := range tzs {
			tzList[tz] = true
		}
	}
}

func init() {
	initDateTimeTrie()
	initTimezoneList()
}

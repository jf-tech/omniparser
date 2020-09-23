package times

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jf-tech/omniparser/cache"
)

var locCache = cache.NewLoadingCache()

func loadLoc(tz string) (*time.Location, error) {
	loc, err := locCache.Get(tz, func(key interface{}) (interface{}, error) {
		return time.LoadLocation(key.(string))
	})
	if err != nil {
		return nil, err
	}
	return loc.(*time.Location), nil
}

var (
	// 'AM'/'PM'/'0-9' followed by '-' then tz string (which can be 'a-zA-Z0-9', '+', '-', '/')
	tzSuffixRegexp = regexp.MustCompile(`^.*\s*([AP]M|[0-9])\s*-([A-Za-z][A-Za-z/_\-+0-9]+)$`)
)

// SmartParse parses a date time string and returns a time.Time and a tz flag indicates whether the
// date time string contains tz info.
// The date time string can be either date only, or date + time, or date + time + tz.
// The date part of the string has the following supported formats:
//  yyyy-mm-dd
//  mm-dd-yyyy
//  yyyy/mm/dd
//  mm/dd/yyyy
//  mm/dd/yy
//  yyyymmdd
// The time part of the string has the following supported formats:
//  hh:mm:ss
//  hh:mm
//  hhmmss
//  hhmm
// Note 1: all formats above can also be followed by 'AM/PM'.
// Note 2: we don't support sub-seconds, yet.
// The tz part of the string has the following supported formats:
//  Z
//  -/+hh
//  -/+hhmm
//  -/+hh:mm
//  -America/New_York
// Note 1: the tz name in the last one above must come from standard IANA timezone names.
// Upon successful parsing, SmartParse returns the time.Time, and a flag whether the input date time string
// has tz info in it or not. Note the time.Time returned will always have tz baked in, because golang time.Time
// doesn't really have a notion of un-timezone'ed timestamp. So if your date time string is a un-tz'ed relative
// time stamp, such as "2020/09/10T12:34:56", which really means different point of time depending on which
// time zone you try to interpret it in, the returned tz flag will be false but the returned time.Time
// will be "2020/09/10T12:34:56Z" (note the 'Z') if you format it using RFC3339. That's the key subtlety one
// must understand well.
func SmartParse(s string) (t time.Time, tz bool, err error) {
	s = strings.TrimSpace(s)

	var loc *time.Location
	if m := tzSuffixRegexp.FindStringSubmatch(s); m != nil {
		tz := m[len(m)-1]
		loc, err = loadLoc(tz)
		if err != nil {
			return time.Time{}, false, errors.New("unrecognized timezone string '" + tz + "'")
		}
		s = strings.TrimSpace(s[:len(s)-len(tz)-1]) // -1 for the '-' that is not included in tz.
	}

	v, found := dateTimeTrie.Get(s)
	if !found {
		return time.Time{}, false, fmt.Errorf("unable to parse '%s' in any supported date/time format", s)
	}
	e := v.(trieEntry)

	if loc != nil {
		t, err = time.ParseInLocation(e.layout, s, loc)
		return t, true, err
	}
	t, err = time.Parse(e.layout, s)
	return t, e.tz, err
}

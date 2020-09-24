package times

import (
	"time"

	"github.com/jf-tech/omniparser/strs"
)

// OverwriteTZ takes the literal values of y/m/d/h/m/s of a time.Time and uses them
// together with a supplied timezone to form a new time.Time, effectively "overwriting"
// the original time.Time's tz. If tz is empty, the original time.Time is returned.
func OverwriteTZ(t time.Time, tz string) (time.Time, error) {
	if !strs.IsStrNonBlank(tz) {
		return t, nil
	}
	loc, err := loadLoc(tz)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
}

// ConvertTZ converts a time.Time into a new time.Time in a timezone specified by tz.
// This is unlike OverwriteTZ, where its tz is overwritten, resulting in a completely
// different point of time: ConvertTZ keeps the point of time unchanged - it merely
// switches the input time.Time to a different timezone.
func ConvertTZ(t time.Time, tz string) (time.Time, error) {
	if !strs.IsStrNonBlank(tz) {
		return t, nil
	}
	loc, err := loadLoc(tz)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

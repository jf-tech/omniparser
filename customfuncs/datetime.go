package customfuncs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/times"

	"github.com/jf-tech/omniparser/transformctx"
)

const (
	rfc3339NoTZ = "2006-01-02T15:04:05"
	second      = "SECOND"
	millisecond = "MILLISECOND"
)

// parseDateTime parse in an input datetime string and returns a time.Time and a flag indicate there
// is tz in it or not.
//
// If layout is specified, then parseDateTime will parse datetime string using the supplied layout;
// otherwise, it will default to times.SmartParse.
//
// If datetime string contains tz info in it (such as 'Z', or '-America/New_York' etc., or '-0700', etc.)
// then fromTZ is IGNORED. Otherwise, the datetime string will be parsed in with its face value y/m/d/h/m/s
// and bonded with the fromTZ, if fromTZ isn't "".
//
// Once datetime is parsed (and fromTZ bonded if needed), then it will be converted into toTZ, if toTZ isn't
// "".
func parseDateTime(datetime, layout string, layoutHasTZ bool, fromTZ, toTZ string) (time.Time, bool, error) {
	var t time.Time
	var hasTZ bool
	var err error
	if layout == "" {
		t, hasTZ, err = times.SmartParse(datetime)
		if err != nil {
			return time.Time{}, false, err
		}
	} else {
		t, err = time.Parse(layout, datetime)
		if err != nil {
			return time.Time{}, false, err
		}
		hasTZ = layoutHasTZ
	}
	// Only use fromTZ if the original t doesn't have tz info baked in.
	if !hasTZ && fromTZ != "" {
		t, err = times.OverwriteTZ(t, fromTZ)
		if err != nil {
			return time.Time{}, false, err
		}
		hasTZ = true
	}
	if toTZ != "" {
		if hasTZ {
			t, err = times.ConvertTZ(t, toTZ)
		} else {
			t, err = times.OverwriteTZ(t, toTZ)
			hasTZ = true
		}
		if err != nil {
			return time.Time{}, false, err
		}
	}
	return t, hasTZ, nil
}

func rfc3339(t time.Time, hasTZ bool) string {
	if hasTZ {
		return t.Format(time.RFC3339)
	}
	return t.Format(rfc3339NoTZ)
}

// DateTimeToRFC3339 parses a 'datetime' string intelligently, normalizes and returns it in RFC3339 format.
// 'fromTZ' is only used if 'datetime' doesn't contain TZ info; if not specified, the parser will keep the
// original TZ (or lack of it) of 'datetime'. 'toTZ' decides what TZ the output RFC3339 date time will be in.
func DateTimeToRFC3339(_ *transformctx.Ctx, datetime, fromTZ, toTZ string) (string, error) {
	if datetime == "" {
		return "", nil
	}
	t, hasTZ, err := parseDateTime(datetime, "", false, fromTZ, toTZ)
	if err != nil {
		return "", err
	}
	return rfc3339(t, hasTZ), nil
}

// DateTimeLayoutToRFC3339 parses a 'datetime' string according a given 'layout' and normalizes and returns it
// in RFC3339 format. 'layoutTZ' specifies whether the 'layout' contains timezone info (such as tz offset, tz
// short names lik 'PST', or standard IANA tz long name, such as 'America/Los_Angeles'); note 'layoutTZ' is a
// string with two possible values "true" or "false". 'fromTZ' is only used if 'datetime'/'layout' don't contain
// TZ info; if not specified, the parser will keep the original TZ (or lack of it) of 'datetime'. 'toTZ' decides
// what TZ the output RFC3339 date time will be in.
func DateTimeLayoutToRFC3339(_ *transformctx.Ctx, datetime, layout, layoutTZ, fromTZ, toTZ string) (string, error) {
	layoutTZFlag := false
	if layout != "" && layoutTZ != "" {
		var err error
		layoutTZFlag, err = strconv.ParseBool(layoutTZ)
		if err != nil {
			return "", err
		}
	}
	if datetime == "" {
		return "", nil
	}
	t, hasTZ, err := parseDateTime(datetime, layout, layoutTZFlag, fromTZ, toTZ)
	if err != nil {
		return "", err
	}
	return rfc3339(t, hasTZ), nil
}

const (
	epochUnitMilliseconds = "MILLISECOND"
	epochUnitSeconds      = "SECOND"
)

// DateTimeToEpoch parses a 'datetime' string intelligently, and returns its epoch number. 'fromTZ'
// is only used if 'datetime' doesn't contain TZ info; if not specified, the parser will keep the
// original TZ (or lack of it) of 'datetime'. 'unit' determines the time unit resolution of the
// output epoch number.
func DateTimeToEpoch(_ *transformctx.Ctx, datetime, fromTZ, unit string) (string, error) {
	if datetime == "" {
		return "", nil
	}
	t, _, err := parseDateTime(datetime, "", false, fromTZ, "")
	if err != nil {
		return "", err
	}
	switch unit {
	case epochUnitMilliseconds:
		return strconv.FormatInt(t.UnixNano()/int64(time.Millisecond), 10), nil
	case epochUnitSeconds:
		return strconv.FormatInt(t.Unix(), 10), nil
	default:
		return "", fmt.Errorf("unknown epoch unit '%s'", unit)
	}
}

// EpochToDateTimeRFC3339 translates the 'epoch' timestamp under the given 'unit' into an RFC3339 formatted
// datetime string in the given timezone 'tz', if specified.
func EpochToDateTimeRFC3339(_ *transformctx.Ctx, epoch, unit string, tz ...string) (string, error) {
	if epoch == "" {
		return "", nil
	}
	if len(tz) > 1 {
		return "", fmt.Errorf("cannot specify tz argument more than once")
	}
	n, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return "", err
	}
	timezone := "UTC"
	if len(tz) == 1 {
		timezone = tz[0]
	}
	loc, err := caches.GetTimeLocation(timezone)
	if err != nil {
		return "", err
	}
	var t time.Time
	switch unit {
	case epochUnitSeconds:
		t = time.Unix(n, 0)
	case epochUnitMilliseconds:
		t = time.Unix(0, n*(int64(time.Millisecond)))
	default:
		return "", fmt.Errorf("unknown epoch unit '%s'", unit)
	}
	return rfc3339(t.In(loc), true), nil
}

// Now returns the current time in UTC in RFC3339 format.
func Now(_ *transformctx.Ctx) (string, error) {
	return rfc3339(time.Now().UTC(), true), nil
}

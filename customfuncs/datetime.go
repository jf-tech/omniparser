package customfuncs

import (
	"errors"
	"strconv"
	"time"

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
// If layout is specific, then parseDateTime will parse datetime string using the supplied layout;
// otherwise, it will default to times.SmartParse.
//
// If datetime string contains tz info in it (such as 'Z', or '-America/New_York' etc, or '-0700', etc)
// then fromTZ is IGNORED. Otherwise, then datetime string will be parsed in with its face value y/m/d/h/m/s
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

func dateTimeToRFC3339(_ *transformctx.Ctx, datetime, fromTZ, toTZ string) (string, error) {
	if datetime == "" {
		return "", nil
	}
	t, hasTZ, err := parseDateTime(datetime, "", false, fromTZ, toTZ)
	if err != nil {
		return "", err
	}
	return rfc3339(t, hasTZ), nil
}

func dateTimeLayoutToRFC3339(_ *transformctx.Ctx, datetime, layout, layoutTZ, fromTZ, toTZ string) (string, error) {
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

func dateTimeToEpoch(_ *transformctx.Ctx, datetime, fromTZ, epochUnit string) (string, error) {
	if datetime == "" {
		return "", nil
	}
	t, _, err := parseDateTime(datetime, "", false, fromTZ, "")
	if err != nil {
		return "", err
	}
	epoch := int64(0)
	switch epochUnit {
	case second:
		epoch = t.Unix()
	case millisecond:
		epoch = t.UnixNano() / int64(time.Millisecond)
	default:
		return "", errors.New("unsupported time unit '" + epochUnit + "'")
	}
	return strconv.FormatInt(epoch, 10), nil
}

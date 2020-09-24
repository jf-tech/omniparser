package customfuncs

import (
	"strconv"
	"time"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/times"
)

const (
	rfc3339NoTZ = "2006-01-02T15:04:05"
)

// normalizeToRFC3339 parse in an input datetime string and normalizes it into RFC3339 standard format.
//
// If layout is specific, then normalizeToRFC3339 will parse datetime string using the supplied layout;
// otherwise, it will default to times.SmartParse.
//
// If datetime string contains tz info in it (such as 'Z', or '-America/New_York' etc, or '-0700', etc)
// then fromTZ is IGNORED. Otherwise, then datetime string will be parsed in with its face value y/m/d/h/m/s
// and bonded with the fromTZ, if fromTZ isn't "".
//
// Once datetime is parsed (and fromTZ bonded if needed), then it will be converted into toTZ, if toTZ isn't
// "".
//
// Finally it will be formatted into RFC3339.
func normalizeToRFC3339(datetime, layout string, layoutHasTZ bool, fromTZ, toTZ string) (string, error) {
	if datetime == "" {
		return "", nil
	}
	var t time.Time
	var hasTZ bool
	var err error
	if layout == "" {
		t, hasTZ, err = times.SmartParse(datetime)
		if err != nil {
			return "", err
		}
	} else {
		t, err = time.Parse(layout, datetime)
		if err != nil {
			return "", err
		}
		hasTZ = layoutHasTZ
	}
	// Only use fromTZ if the original t doesn't have tz info baked in.
	if !hasTZ && fromTZ != "" {
		t, err = times.OverwriteTZ(t, fromTZ)
		if err != nil {
			return "", err
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
			return "", err
		}
	}
	if hasTZ {
		return t.Format(time.RFC3339), nil
	}
	return t.Format(rfc3339NoTZ), nil
}

func dateTimeToRFC3339(_ *transformctx.Ctx, datetime, fromTZ, toTZ string) (string, error) {
	return normalizeToRFC3339(datetime, "", false, fromTZ, toTZ)
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
	return normalizeToRFC3339(datetime, layout, layoutTZFlag, fromTZ, toTZ)
}

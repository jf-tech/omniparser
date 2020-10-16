package customfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeToRFC3339(t *testing.T) {
	for _, test := range []struct {
		name     string
		datetime string
		fromTZ   string
		toTZ     string
		err      string
		expected string
	}{
		{
			name:     "empty datetime -> no op",
			datetime: "",
			fromTZ:   "UTC",
			toTZ:     "America/New_York",
			err:      "",
			expected: "",
		},
		{
			name:     "invalid datetime",
			datetime: "invalid",
			fromTZ:   "UTC",
			toTZ:     "America/New_York",
			err:      "unable to parse 'invalid' in any supported date/time format",
			expected: "",
		},
		{
			name:     "invalid fromTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "invalid",
			toTZ:     "America/New_York",
			err:      "unknown time zone invalid",
			expected: "",
		},
		{
			name:     "invalid toTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "UTC",
			toTZ:     "invalid",
			err:      "unknown time zone invalid",
			expected: "",
		},
		{
			name:     "datetime no tz; no fromTZ; no toTZ -> result no tz",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "",
			toTZ:     "",
			err:      "",
			expected: "2020-09-22T12:34:56",
		},
		{
			name:     "datetime no tz; no fromTZ; with toTZ -> result direct use toTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "",
			toTZ:     "America/New_York",
			err:      "",
			expected: "2020-09-22T12:34:56-04:00",
		},
		{
			name:     "datetime no tz; with fromTZ; no toTZ -> result direct use fromTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "America/Los_Angeles",
			toTZ:     "",
			err:      "",
			expected: "2020-09-22T12:34:56-07:00",
		},
		{
			name:     "datetime no tz; with fromTZ; with toTZ -> datetime + fromTZ then converts to toTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "America/Los_Angeles",
			toTZ:     "America/New_York",
			err:      "",
			expected: "2020-09-22T15:34:56-04:00",
		},
		{
			name:     "datetime with tz; no fromTZ; no toTZ -> result is datetime with tz",
			datetime: "2020/09/22T12:34:56-05",
			fromTZ:   "",
			toTZ:     "",
			err:      "",
			expected: "2020-09-22T12:34:56-05:00",
		},
		{
			name:     "datetime with tz; no fromTZ; with toTZ -> datetime with tz converts to toTZ",
			datetime: "2020/09/22T12:34:56-05",
			fromTZ:   "",
			toTZ:     "America/New_York",
			err:      "",
			expected: "2020-09-22T13:34:56-04:00",
		},
		{
			name:     "datetime with tz; with fromTZ; no toTZ -> fromTZ ignored",
			datetime: "2020/09/22T12:34:56-05",
			fromTZ:   "America/Los_Angeles",
			toTZ:     "",
			err:      "",
			expected: "2020-09-22T12:34:56-05:00",
		},
		{
			name:     "datetime with tz; with fromTZ; with toTZ -> fromTZ ignored, datetime with tz converts to toTZ",
			datetime: "2020/09/22T12:34:56-05",
			fromTZ:   "America/Los_Angeles",
			toTZ:     "America/New_York",
			err:      "",
			expected: "2020-09-22T13:34:56-04:00",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := dateTimeToRFC3339(nil, test.datetime, test.fromTZ, test.toTZ)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestDateTimeLayoutToRFC3339(t *testing.T) {
	for _, test := range []struct {
		name     string
		datetime string
		layout   string
		layoutTZ string
		fromTZ   string
		toTZ     string
		err      string
		expected string
	}{
		{
			name:     "empty datetime -> no op",
			datetime: "",
			layout:   "2006+01+02 X 15:04:05",
			layoutTZ: "false",
			fromTZ:   "UTC",
			toTZ:     "America/New_York",
			err:      "",
			expected: "",
		},
		{
			name:     "layout but no tz",
			datetime: "2020+09+22 X 12:34:56",
			layout:   "2006+01+02 X 15:04:05",
			layoutTZ: "false",
			fromTZ:   "America/Los_Angeles",
			toTZ:     "America/New_York",
			err:      "",
			expected: "2020-09-22T15:34:56-04:00",
		},
		{
			name:     "layout with tz",
			datetime: "2020+09+22 X 12:34:56 +02",
			layout:   "2006+01+02 X 15:04:05 -07",
			layoutTZ: "true",
			fromTZ:   "",
			toTZ:     "America/Chicago",
			err:      "",
			expected: "2020-09-22T05:34:56-05:00",
		},
		{
			name:     "invalid layoutTZ flag",
			datetime: "",
			layout:   "whatever",
			layoutTZ: "not a bool value",
			fromTZ:   "",
			toTZ:     "",
			err:      `strconv.ParseBool: parsing "not a bool value": invalid syntax`,
			expected: "",
		},
		{
			name:     "layout parsing failed",
			datetime: "not valid",
			layout:   "2006/01/02 15:04:05",
			layoutTZ: "false",
			fromTZ:   "",
			toTZ:     "",
			err:      `parsing time "not valid" as "2006/01/02 15:04:05": cannot parse "not valid" as "2006"`,
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := dateTimeLayoutToRFC3339(
				nil, test.datetime, test.layout, test.layoutTZ, test.fromTZ, test.toTZ)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

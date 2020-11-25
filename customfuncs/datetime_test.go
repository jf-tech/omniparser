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
			result, err := DateTimeToRFC3339(nil, test.datetime, test.fromTZ, test.toTZ)
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
			result, err := DateTimeLayoutToRFC3339(
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

func TestDateTimeToEpoch(t *testing.T) {
	for _, test := range []struct {
		name     string
		datetime string
		fromTZ   string
		unit     string
		err      string
		expected string
	}{
		{
			name:     "empty datetime -> no op",
			datetime: "",
			fromTZ:   "UTC",
			unit:     epochUnitMilliseconds,
			err:      "",
			expected: "",
		},
		{
			name:     "invalid datetime",
			datetime: "invalid",
			fromTZ:   "UTC",
			unit:     epochUnitSeconds,
			err:      "unable to parse 'invalid' in any supported date/time format",
			expected: "",
		},
		{
			name:     "invalid fromTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "invalid",
			unit:     epochUnitMilliseconds,
			err:      "unknown time zone invalid",
			expected: "",
		},
		{
			name:     "invalid unit",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "UTC",
			unit:     "invalid",
			err:      "unknown epoch unit 'invalid'",
			expected: "",
		},
		{
			name:     "datetime no tz; no fromTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "",
			unit:     epochUnitMilliseconds,
			err:      "",
			expected: "1600778096000",
		},
		{
			name:     "datetime no tz; with fromTZ",
			datetime: "2020/09/22T12:34:56",
			fromTZ:   "America/Los_Angeles",
			unit:     epochUnitSeconds,
			err:      "",
			expected: "1600803296",
		},
		{
			name:     "datetime with tz; with fromTZ",
			datetime: "2020/09/22T12:34:56-05",
			fromTZ:   "America/Los_Angeles",
			unit:     epochUnitSeconds,
			err:      "",
			expected: "1600796096",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := DateTimeToEpoch(nil, test.datetime, test.fromTZ, test.unit)
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

func TestEpochToDateTimeRFC3339(t *testing.T) {
	for _, test := range []struct {
		name     string
		epoch    string
		unit     string
		tz       []string
		err      string
		expected string
	}{
		{
			name:     "empty epoch -> no op",
			epoch:    "",
			unit:     epochUnitMilliseconds,
			tz:       nil,
			err:      "",
			expected: "",
		},
		{
			name:     "more than one tz specified",
			epoch:    "1234567",
			unit:     epochUnitMilliseconds,
			tz:       []string{"UTC", "UTC"},
			err:      "cannot specify tz argument more than once",
			expected: "",
		},
		{
			name:     "invalid epoch",
			epoch:    "invalid",
			unit:     epochUnitSeconds,
			tz:       nil,
			err:      `strconv.ParseInt: parsing "invalid": invalid syntax`,
			expected: "",
		},
		{
			name:     "invalid tz",
			epoch:    "12345",
			unit:     epochUnitSeconds,
			tz:       []string{"invalid"},
			err:      "unknown time zone invalid",
			expected: "",
		},
		{
			name:     "invalid unit",
			epoch:    "12345",
			unit:     "invalid",
			tz:       nil,
			err:      "unknown epoch unit 'invalid'",
			expected: "",
		},
		{
			name:     "no tz",
			epoch:    "1234567890123",
			unit:     epochUnitMilliseconds,
			tz:       nil,
			err:      "",
			expected: "2009-02-13T23:31:30Z",
		},
		{
			name:     "with tz",
			epoch:    "1234567890",
			unit:     epochUnitSeconds,
			tz:       []string{"America/Los_Angeles"},
			err:      "",
			expected: "2009-02-13T15:31:30-08:00",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := EpochToDateTimeRFC3339(nil, test.epoch, test.unit, test.tz...)
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

func TestNow(t *testing.T) {
	now, err := Now(nil)
	assert.NoError(t, err)
	assert.True(t, len(now) > 0)
}

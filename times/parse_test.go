package times

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadLoc(t *testing.T) {
	for _, test := range []struct {
		tz          string
		expectedErr string
	}{
		{
			tz:          "America/Los_Angeles",
			expectedErr: "",
		},
		{
			tz:          "America/Indiana/Indianapolis",
			expectedErr: "",
		},
		{
			tz:          "Unknown",
			expectedErr: "unknown time zone Unknown",
		},
	} {
		t.Run(test.tz, func(t *testing.T) {
			loc, err := loadLoc(test.tz)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Nil(t, loc)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.tz, loc.String())
			}
		})
	}
}

func TestSmartParse_Success(t *testing.T) {
	for _, test := range []struct {
		name            string
		input           string
		expectedRFC3339 string
		expectedTZ      bool
	}{
		// dates only
		{
			name:            "yyyy-mm-dd",
			input:           "2020-09-22",
			expectedRFC3339: "2020-09-22T00:00:00Z",
			expectedTZ:      false,
		},
		{
			name:            "mm-dd-yyyy",
			input:           "09-22-2020",
			expectedRFC3339: "2020-09-22T00:00:00Z",
			expectedTZ:      false,
		},
		{

			name:            "yyyy/mm/dd",
			input:           "2020/09/22",
			expectedRFC3339: "2020-09-22T00:00:00Z",
			expectedTZ:      false,
		},
		{

			name:            "mm/dd/yyyy",
			input:           "09/22/2020",
			expectedRFC3339: "2020-09-22T00:00:00Z",
			expectedTZ:      false,
		},
		{
			name:            "mm/dd/yy",
			input:           "09/22/20",
			expectedRFC3339: "2020-09-22T00:00:00Z",
			expectedTZ:      false,
		},
		{
			name:            "yyyymmdd",
			input:           "20200922",
			expectedRFC3339: "2020-09-22T00:00:00Z",
			expectedTZ:      false,
		},
		// date + time
		{
			name:            "yyyy-mm-ddThh:mm:ss",
			input:           "2020-09-22T12:34:56",
			expectedRFC3339: "2020-09-22T12:34:56Z",
			expectedTZ:      false,
		},
		{
			name:            "yyyy-mm-ddThh:mm:ss.sssssssss",
			input:           "2020-09-22T12:34:56.123456789",
			expectedRFC3339: "2020-09-22T12:34:56.123456789Z",
			expectedTZ:      false,
		},
		{
			name:            "yyyy/mm/dd hh:mm",
			input:           "2020/09/22 12:34",
			expectedRFC3339: "2020-09-22T12:34:00Z",
			expectedTZ:      false,
		},
		{
			name:            "yyyy-mm-ddThhmmss",
			input:           "2020-09-22T123456",
			expectedRFC3339: "2020-09-22T12:34:56Z",
			expectedTZ:      false,
		},
		{
			name:            "yyyy/mm/dd hhmm",
			input:           "2020/09/22 1234",
			expectedRFC3339: "2020-09-22T12:34:00Z",
			expectedTZ:      false,
		},
		// with tz
		{
			name:            "yyyy-mm-ddThh:mm:ssZ",
			input:           "2020-09-22T12:34:56Z",
			expectedRFC3339: "2020-09-22T12:34:56Z",
			expectedTZ:      true,
		},
		{
			name:            "mm/dd/yyyyThh:mm:ss-hh",
			input:           "09/22/2020T12:34:56-08",
			expectedRFC3339: "2020-09-22T12:34:56-08:00",
			expectedTZ:      true,
		},
		{
			name:            "mm/dd/yyyy hh:mm:ss+hhmm",
			input:           "09/22/2020 12:34:56+0530",
			expectedRFC3339: "2020-09-22T12:34:56+05:30",
			expectedTZ:      true,
		},
		{
			name:            "yyyy/mm/dd hh:mm:ss+hh:mm",
			input:           "2020/09/22 12:34:56+05:30",
			expectedRFC3339: "2020-09-22T12:34:56+05:30",
			expectedTZ:      true,
		},
		{
			name:            "   yyyy-mm-dd hh:mm:ss AM     -IANA-tz   ",
			input:           "   2020-09-22 12:34:56 AM     -America/Indiana/Indianapolis   ",
			expectedRFC3339: "2020-09-22T00:34:56-04:00",
			expectedTZ:      true,
		},
		{
			name:            "   yyyy-mm-dd hh:mm:ss.sssssssss AM     -IANA-tz   ",
			input:           "   2020-09-22 12:34:56.123456789 AM     -America/Indiana/Indianapolis   ",
			expectedRFC3339: "2020-09-22T00:34:56.123456789-04:00",
			expectedTZ:      true,
		},
		{
			name:            "   yyyy-mm-dd hh:mm:ss PM     -IANA-tz   ",
			input:           "   2020-09-22 12:34:56 PM     -Pacific/Auckland   ",
			expectedRFC3339: "2020-09-22T12:34:56+12:00",
			expectedTZ:      true,
		},
		{
			name:            "   yyyy-mm-dd hh:mm:ss.sssssssss PM     -IANA-tz   ",
			input:           "   2020-09-22 12:34:56.123456789 PM     -Pacific/Auckland   ",
			expectedRFC3339: "2020-09-22T12:34:56.123456789+12:00",
			expectedTZ:      true,
		},
		{
			name:            "   yyyy-mm-dd   hh:mm:ss     -IANA-tz   ",
			input:           "   2020-09-22   12:34:56     -Etc/GMT+10   ",
			expectedRFC3339: "2020-09-22T12:34:56-10:00",
			expectedTZ:      true,
		},
		{
			name:            "yyyy/mm/ddThh:mm:ss-Eire ",
			input:           "2020/09/22T12:34:56-Eire ",
			expectedRFC3339: "2020-09-22T12:34:56+01:00",
			expectedTZ:      true,
		},
		{
			name:            "yyyy/mm/ddThh:mm:ss-GB-Eire ",
			input:           "2020/09/22T12:34:56-GB-Eire ",
			expectedRFC3339: "2020-09-22T12:34:56+01:00",
			expectedTZ:      true,
		},
		{
			name:            "yyyy/mm/ddThh:mm:ss-US/Indiana-Starke",
			input:           "2020/09/22T12:34:56-US/Indiana-Starke",
			expectedRFC3339: "2020-09-22T12:34:56-05:00",
			expectedTZ:      true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t1, tz, err := SmartParse(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedRFC3339, t1.Format(time.RFC3339Nano))
			assert.Equal(t, test.expectedTZ, tz)
		})
	}
}

func TestSmartParse_Failure(t *testing.T) {
	for _, test := range []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "invalid IANA timezone",
			input:       "2020-09-22T12:34:56-Unknown",
			expectedErr: "unable to parse '2020-09-22T12:34:56-Unknown' in any supported date/time format",
		},
		{
			name:        "not supported date format",
			input:       "2020 09 22",
			expectedErr: "unable to parse '2020 09 22' in any supported date/time format",
		},
		{
			name:        "time.ParseInLocation fails",
			input:       "2020-13-22T12:34:56-America/Los_Angeles",
			expectedErr: `parsing time "2020-13-22T12:34:56": month out of range`,
		},
		{
			name:        "time.Parse fails",
			input:       "2020-09-22T12:74:56Z",
			expectedErr: `parsing time "2020-09-22T12:74:56Z": minute out of range`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t1, _, err := SmartParse(test.input)
			assert.Error(t, err)
			assert.Equal(t, test.expectedErr, err.Error())
			assert.Equal(t, time.Time{}, t1)
		})
	}
}

func BenchmarkTimeParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := time.Parse("2006-01-02 03:04:05 PM", "2020-09-22 12:34:56 AM")
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkSmartParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := SmartParse("2020-09-22 12:34:56 AM")
		if err != nil {
			b.FailNow()
		}
	}
}

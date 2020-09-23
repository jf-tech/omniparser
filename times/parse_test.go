package times

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
			name:            "   yyyy-mm-dd hh:mm:ss PM     -IANA-tz   ",
			input:           "   2020-09-22 12:34:56 PM     -Pacific/Auckland   ",
			expectedRFC3339: "2020-09-22T12:34:56+12:00",
			expectedTZ:      true,
		},
		{
			name:            "   yyyy-mm-dd hh:mm:ss     -IANA-tz   ",
			input:           "   2020-09-22 12:34:56     -Etc/GMT+10   ",
			expectedRFC3339: "2020-09-22T12:34:56-10:00",
			expectedTZ:      true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t1, tz, err := SmartParse(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedRFC3339, t1.Format(time.RFC3339))
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
			expectedErr: "unrecognized timezone string 'Unknown'",
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
func BenchmarkSmartParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := SmartParse("   2020-09-22 12:34:56 AM     -America/Indiana/Indianapolis   ")
		if err != nil {
			b.FailNow()
		}
	}
}

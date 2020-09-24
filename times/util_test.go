package times

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOverwriteTZ(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		tz       string
		err      string
		expected string
	}{
		{
			name:     "empty tz -> no op",
			input:    "2019/07/18T12:34:56-0700",
			tz:       "",
			err:      "",
			expected: "2019-07-18T12:34:56-07:00",
		},
		{
			name:     "datetime w/o tz -> America/Los_Angeles",
			input:    "2019/07/18T12:34:56",
			tz:       "America/Los_Angeles",
			err:      "",
			expected: "2019-07-18T12:34:56-07:00",
		},
		{
			name:     "datetime w/ tz -> America/Los_Angeles",
			input:    "2019/07/18T12:34:56+05",
			tz:       "America/Los_Angeles",
			err:      "",
			expected: "2019-07-18T12:34:56-07:00",
		},
		{
			name:     "invalid tz",
			input:    "2019/07/18T12:34:56+05",
			tz:       "invalid",
			err:      "unknown time zone invalid",
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t1, _, err := SmartParse(test.input)
			assert.NoError(t, err)
			t2, err := OverwriteTZ(t1, test.tz)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, time.Time{}, t2)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, t2.Format(time.RFC3339))
			}
		})
	}
}

func TestConvertTZ(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		tz       string
		err      string
		expected string
	}{
		{
			name:     "empty tz -> no op",
			input:    "2019/07/18T12:34:56-0700",
			tz:       "",
			err:      "",
			expected: "2019-07-18T12:34:56-07:00",
		},
		{
			// even though input has no tz, but SmartParse (and time.Parse)
			// will assign UTC to it by default.
			name:     "datetime w/o tz -> America/Los_Angeles",
			input:    "2019/07/18T12:34:56",
			tz:       "America/Los_Angeles",
			err:      "",
			expected: "2019-07-18T05:34:56-07:00",
		},
		{
			name:     "datetime w/ tz -> America/Los_Angeles",
			input:    "2019/07/18T12:34:56+05",
			tz:       "America/Los_Angeles",
			err:      "",
			expected: "2019-07-18T00:34:56-07:00",
		},
		{
			name:     "invalid tz",
			input:    "2019/07/18T12:34:56+05",
			tz:       "invalid",
			err:      "unknown time zone invalid",
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t1, _, err := SmartParse(test.input)
			assert.NoError(t, err)
			t2, err := ConvertTZ(t1, test.tz)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, time.Time{}, t2)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, t2.Format(time.RFC3339))
			}
		})
	}
}

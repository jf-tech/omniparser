package customfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvg(t *testing.T) {
	for _, test := range []struct {
		name     string
		inputs   []string
		err      string
		expected string
	}{
		{
			name:     "nil",
			inputs:   nil,
			err:      "",
			expected: "0",
		},
		{
			name:     "empty",
			inputs:   []string{},
			err:      "",
			expected: "0",
		},
		{
			name:     "single",
			inputs:   []string{"3.14159265358"},
			err:      "",
			expected: "3.14159265358",
		},
		{
			name:     "multiple small ones",
			inputs:   []string{"3.45", "5.38"},
			err:      "",
			expected: "4.415",
		},
		{
			name:     "multiple big ones",
			inputs:   []string{"1.23e+9", "0.34E+10"},
			err:      "",
			expected: "2.315e+09",
		},
		{
			name:     "invalid value",
			inputs:   []string{"1", "two"},
			err:      `strconv.ParseFloat: parsing "two": invalid syntax`,
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := avg(nil, test.inputs...)
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

func TestSum(t *testing.T) {
	for _, test := range []struct {
		name     string
		inputs   []string
		err      string
		expected string
	}{
		{
			name:     "nil",
			inputs:   nil,
			err:      "",
			expected: "0",
		},
		{
			name:     "empty",
			inputs:   []string{},
			err:      "",
			expected: "0",
		},
		{
			name:     "single",
			inputs:   []string{"3.14159265358"},
			err:      "",
			expected: "3.14159265358",
		},
		{
			name:     "multiple small ones",
			inputs:   []string{"3.45", "5.38"},
			err:      "",
			expected: "8.83",
		},
		{
			name:     "multiple big ones",
			inputs:   []string{"1.23e+9", "0.34E+10"},
			err:      "",
			expected: "4.63e+09",
		},
		{
			name:     "invalid value",
			inputs:   []string{"1", "two"},
			err:      `strconv.ParseFloat: parsing "two": invalid syntax`,
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := sum(nil, test.inputs...)
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

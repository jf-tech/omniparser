package customfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvg(t *testing.T) {
	for _, test := range []struct {
		name        string
		inputs      []string
		expectedErr string
		expectedAvg string
	}{
		{
			name:        "nil",
			inputs:      nil,
			expectedErr: "",
			expectedAvg: "0",
		},
		{
			name:        "empty",
			inputs:      []string{},
			expectedErr: "",
			expectedAvg: "0",
		},
		{
			name:        "single",
			inputs:      []string{"3.14159265358"},
			expectedErr: "",
			expectedAvg: "3.14159265358",
		},
		{
			name:        "multiple small ones",
			inputs:      []string{"3.45", "5.38"},
			expectedErr: "",
			expectedAvg: "4.415",
		},
		{
			name:        "multiple big ones",
			inputs:      []string{"1.23e+9", "0.34E+10"},
			expectedErr: "",
			expectedAvg: "2.315e+09",
		},
		{
			name:        "invalid value",
			inputs:      []string{"1", "two"},
			expectedErr: `strconv.ParseFloat: parsing "two": invalid syntax`,
			expectedAvg: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := avg(nil, test.inputs...)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", result)
			} else {

				assert.NoError(t, err)
				assert.Equal(t, test.expectedAvg, result)
			}
		})
	}
}

func TestSum(t *testing.T) {
	for _, test := range []struct {
		name        string
		inputs      []string
		expectedErr string
		expectedSum string
	}{
		{
			name:        "nil",
			inputs:      nil,
			expectedErr: "",
			expectedSum: "0",
		},
		{
			name:        "empty",
			inputs:      []string{},
			expectedErr: "",
			expectedSum: "0",
		},
		{
			name:        "single",
			inputs:      []string{"3.14159265358"},
			expectedErr: "",
			expectedSum: "3.14159265358",
		},
		{
			name:        "multiple small ones",
			inputs:      []string{"3.45", "5.38"},
			expectedErr: "",
			expectedSum: "8.83",
		},
		{
			name:        "multiple big ones",
			inputs:      []string{"1.23e+9", "0.34E+10"},
			expectedErr: "",
			expectedSum: "4.63e+09",
		},
		{
			name:        "invalid value",
			inputs:      []string{"1", "two"},
			expectedErr: `strconv.ParseFloat: parsing "two": invalid syntax`,
			expectedSum: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := sum(nil, test.inputs...)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedSum, result)
			}
		})
	}
}

package customfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/strs"
)

func TestMerge(t *testing.T) {
	fs1 := CustomFuncs{
		"a": 1,
		"b": 2,
	}
	fs2 := CustomFuncs{
		"a": 3,
		"c": 4,
	}
	assert.Equal(t,
		CustomFuncs{
			"a": 3,
			"b": 2,
			"c": 4,
		},
		Merge(fs1, fs2, nil))
}

func TestConcat(t *testing.T) {
	for _, test := range []struct {
		name     string
		strs     []string
		expected string
	}{
		{
			name:     "empty input",
			strs:     nil,
			expected: "",
		},
		{
			name:     "one empty string input",
			strs:     []string{""},
			expected: "",
		},
		{
			name:     "one non-empty string input",
			strs:     []string{"a b c"},
			expected: "a b c",
		},
		{
			name:     "multiple strings",
			strs:     []string{"", "a b c", "", " e f"},
			expected: "a b c e f",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := concat(nil, test.strs...)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestExternal(t *testing.T) {
	for _, test := range []struct {
		name               string
		externalProperties map[string]string
		propNameToLookUp   string
		expectedErr        string
		expectedValue      string
	}{
		{
			name:               "externalProperties nil",
			externalProperties: nil,
			propNameToLookUp:   "abc",
			expectedErr:        "cannot find external property 'abc'",
			expectedValue:      "",
		},
		{
			name:               "externalProperties empty",
			externalProperties: map[string]string{},
			propNameToLookUp:   "efg",
			expectedErr:        "cannot find external property 'efg'",
			expectedValue:      "",
		},
		{
			name:               "can't find prop",
			externalProperties: map[string]string{"abc": "abc"},
			propNameToLookUp:   "efg",
			expectedErr:        "cannot find external property 'efg'",
			expectedValue:      "",
		},
		{
			name:               "found",
			externalProperties: map[string]string{"abc": "123"},
			propNameToLookUp:   "abc",
			expectedErr:        "",
			expectedValue:      "123",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			v, err := external(
				&transformctx.Ctx{ExternalProperties: test.externalProperties},
				test.propNameToLookUp,
			)
			switch {
			case strs.IsStrNonBlank(test.expectedErr):
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", v)
			default:
				assert.NoError(t, err)
				assert.Equal(t, test.expectedValue, v)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	for _, test := range []struct {
		name           string
		value          string
		decimalPlaces  string
		expectedErr    string
		expectedResult string
	}{
		{
			name:           "invalid value",
			value:          "??",
			decimalPlaces:  "2",
			expectedErr:    `unable to parse value '??' to float64: strconv.ParseFloat: parsing "??": invalid syntax`,
			expectedResult: "",
		},
		{
			name:           "invalid decimal place value",
			value:          "3.1415926",
			decimalPlaces:  "??",
			expectedErr:    `unable to parse decimal place value '??' to int: strconv.Atoi: parsing "??": invalid syntax`,
			expectedResult: "",
		},
		{
			name:           "decimal places less than 0",
			value:          "3.1415926",
			decimalPlaces:  "-1",
			expectedErr:    `decimal place value must be an integer with range of [0,100], instead, got -1`,
			expectedResult: "",
		},
		{
			name:           "decimal places > 100",
			value:          "3.1415926",
			decimalPlaces:  "101",
			expectedErr:    `decimal place value must be an integer with range of [0,100], instead, got 101`,
			expectedResult: "",
		},
		{
			name:           "decimal places less than available digits",
			value:          "3.1415926",
			decimalPlaces:  "2",
			expectedErr:    "",
			expectedResult: "3.14",
		},
		{
			name:           "decimal places 0",
			value:          "3.1415926",
			decimalPlaces:  "0",
			expectedErr:    "",
			expectedResult: "3",
		},
		{
			name:           "decimal places more than available digits",
			value:          "3.1415926",
			decimalPlaces:  "20",
			expectedErr:    "",
			expectedResult: "3.1415926",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := floor(nil, test.value, test.decimalPlaces)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}

func TestLower(t *testing.T) {
	s, err := lower(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "", s)

	s, err = lower(nil, "AbCeDfG 0123456789")
	assert.NoError(t, err)
	assert.Equal(t, "abcedfg 0123456789", s)
}

func TestSplitIntoJsonArray(t *testing.T) {
	// success cases
	for _, test := range []struct {
		name        string
		s           string
		sep         string
		trim        string
		expected    string
		expectedErr string
	}{
		{
			name:        "both empty",
			s:           "",
			sep:         "",
			trim:        "",
			expected:    "",
			expectedErr: `'sep' can't be empty`,
		},
		{
			name:        "s with several spaces and sep empty",
			s:           "  ",
			sep:         "",
			trim:        "true",
			expected:    "",
			expectedErr: `'sep' can't be empty`,
		},
		{
			name:        "s with several spaces and sep space",
			s:           "  ",
			sep:         " ",
			trim:        "true",
			expected:    `["","",""]`,
			expectedErr: "",
		},
		{
			name:        "s with several spaces and sep non-space",
			s:           "  ",
			sep:         ",",
			trim:        "true",
			expected:    `[""]`,
			expectedErr: "",
		},
		{
			name:        "s empty",
			s:           "",
			sep:         ",",
			trim:        "true",
			expected:    "[]",
			expectedErr: "",
		},
		{
			name:        "sep empty",
			s:           "ab c",
			sep:         "",
			trim:        "true",
			expected:    "",
			expectedErr: `'sep' can't be empty`,
		},
		{
			name:        "both not empty; no trim",
			s:           "a>b> c",
			sep:         ">",
			trim:        "",
			expected:    `["a","b"," c"]`, // reason is: if sep is empty, s is split into each char.
			expectedErr: "",
		},
		{
			name:        "both not empty; trim",
			s:           "a>b> c",
			sep:         ">",
			trim:        "true",
			expected:    `["a","b","c"]`, // reason is: if sep is empty, s is split into each char.
			expectedErr: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := splitIntoJsonArray(nil, test.s, test.sep, test.trim)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}

	// failure case
	t.Run("invalid trim", func(t *testing.T) {
		result, err := splitIntoJsonArray(nil, "a,b,c", ",", "invalid")
		assert.Error(t, err)
		assert.Equal(t,
			`'trim' must be either "" (default to "false"") or "true" or "false". err: strconv.ParseBool: parsing "invalid": invalid syntax`,
			err.Error())
		assert.Equal(t, "", result)
	})
}

func TestSubstring(t *testing.T) {
	tests := []struct {
		name        string
		str         string
		startIndex  string
		lengthStr   string
		expected    string
		expectedErr string
	}{
		{
			name:        "invalid startIndex",
			str:         "123456",
			startIndex:  "abc",
			lengthStr:   "5",
			expected:    "",
			expectedErr: `unable to convert start index 'abc' into int, err: strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:        "invalid lengthStr",
			str:         "123456",
			startIndex:  "5",
			lengthStr:   "abc",
			expected:    "",
			expectedErr: `unable to convert length 'abc' into int, err: strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:        "empty startIndex",
			str:         "123456",
			startIndex:  "",
			lengthStr:   "5",
			expected:    "",
			expectedErr: `unable to convert start index '' into int, err: strconv.Atoi: parsing "": invalid syntax`,
		},
		{
			name:        "empty lengthStr",
			str:         "123456",
			startIndex:  "5",
			lengthStr:   "",
			expected:    "",
			expectedErr: `unable to convert length '' into int, err: strconv.Atoi: parsing "": invalid syntax`,
		},
		{
			name:        "empty str",
			str:         "",
			startIndex:  "0",
			lengthStr:   "0",
			expected:    "",
			expectedErr: "",
		},
		{
			name:        "empty str with non-0 startIndex",
			str:         "",
			startIndex:  "1",
			lengthStr:   "0",
			expected:    "",
			expectedErr: `start index 1 is out of bounds (string length is 0)`,
		},
		{
			name:        "empty str with non-0 lengthStr",
			str:         "",
			startIndex:  "0",
			lengthStr:   "1",
			expected:    "",
			expectedErr: `start 0 + length 1 is out of bounds (string length is 0)`,
		},
		{
			name:        "0 startIndex",
			str:         "123456",
			startIndex:  "0",
			lengthStr:   "4",
			expected:    "1234",
			expectedErr: "",
		},
		{
			name:        "lengthStr is 1",
			str:         "123456",
			startIndex:  "4",
			lengthStr:   "1",
			expected:    "5",
			expectedErr: "",
		},
		{
			name:        "lengthStr is 0",
			str:         "123456",
			startIndex:  "1",
			lengthStr:   "0",
			expected:    "",
			expectedErr: "",
		},
		{
			name:        "lengthStr is -1",
			str:         "123456",
			startIndex:  "3",
			lengthStr:   "-1",
			expected:    "456",
			expectedErr: "",
		},
		{
			name:        "negative startIndex",
			str:         "123456",
			startIndex:  "-4",
			lengthStr:   "4",
			expected:    "",
			expectedErr: `start index -4 is out of bounds (string length is 6)`,
		},
		{
			name:        "negative lengthStr other than -1",
			str:         "123456",
			startIndex:  "4",
			lengthStr:   "-2",
			expected:    "",
			expectedErr: `length must be >= -1, but got -2`,
		},
		{
			name:        "out-of-bounds startIndex",
			str:         "123456",
			startIndex:  "9",
			lengthStr:   "2",
			expected:    "",
			expectedErr: `start index 9 is out of bounds (string length is 6)`,
		},
		{
			name:        "out-of-bounds lengthStr",
			str:         "123456",
			startIndex:  "2",
			lengthStr:   "7",
			expected:    "",
			expectedErr: `start 2 + length 7 is out of bounds (string length is 6)`,
		},
		{
			name:        "out-of-bounds startIndex and lengthStr",
			str:         "123456",
			startIndex:  "10",
			lengthStr:   "9",
			expected:    "",
			expectedErr: `start index 10 is out of bounds (string length is 6)`,
		},
		{
			name:        "substring starts at the beginning",
			str:         "123456",
			startIndex:  "0",
			lengthStr:   "4",
			expected:    "1234",
			expectedErr: "",
		},
		{
			name:        "substring ends at the end",
			str:         "123456",
			startIndex:  "2",
			lengthStr:   "4",
			expected:    "3456",
			expectedErr: "",
		},
		{
			name:        "substring starts at the end",
			str:         "123456",
			startIndex:  "6",
			lengthStr:   "0",
			expected:    "",
			expectedErr: "",
		},
		{
			name:        "substring ends at the beginning",
			str:         "123456",
			startIndex:  "0",
			lengthStr:   "0",
			expected:    "",
			expectedErr: "",
		},
		{
			name:        "substring is the whole string",
			str:         "123456",
			startIndex:  "0",
			lengthStr:   "6",
			expected:    "123456",
			expectedErr: "",
		},
		{
			name:        "non-ASCII string",
			str:         "ü:ü",
			startIndex:  "1",
			lengthStr:   "2",
			expected:    ":ü",
			expectedErr: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := substring(nil, test.str, test.startIndex, test.lengthStr)
			if test.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", result)
			}
		})
	}
}

func TestUpper(t *testing.T) {
	s, err := upper(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "", s)

	s, err = upper(nil, "abCeDfG 0123456789")
	assert.NoError(t, err)
	assert.Equal(t, "ABCEDFG 0123456789", s)
}

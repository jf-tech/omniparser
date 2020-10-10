package customfuncs

import (
	"sort"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/transformctx"
)

func TestDumpBuiltinCustomFuncNames(t *testing.T) {
	var names []string
	for name := range BuiltinCustomFuncs {
		names = append(names, name)
	}
	sort.Strings(names)
	cupaloy.SnapshotT(t, jsons.BPM(names))
}
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

func TestCoalesce(t *testing.T) {
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
			strs:     []string{"", "", "", " e f"},
			expected: " e f",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := coalesce(nil, test.strs...)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
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

func TestContainsPattern(t *testing.T) {
	for _, test := range []struct {
		name     string
		pattern  string
		strs     []string
		err      string
		expected string
	}{
		{
			name:    "invalid pattern",
			pattern: "[",
			err:     "error parsing regexp: missing closing ]: `[`",
		},
		{
			name:     "empty strs",
			pattern:  ".*",
			strs:     nil,
			err:      "",
			expected: "false",
		},
		{
			name:     "contains",
			pattern:  "^[0-9]+$",
			strs:     []string{"abc", "efg", "123"},
			err:      "",
			expected: "true",
		},
		{
			name:     "not contains",
			pattern:  "^[0-9]+$",
			strs:     []string{"abc", "efg", "x123"},
			err:      "",
			expected: "false",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := containsPattern(nil, test.pattern, test.strs...)
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

func TestExternal(t *testing.T) {
	for _, test := range []struct {
		name      string
		externals map[string]string
		lookup    string
		err       string
		expected  string
	}{
		{
			name:      "externals nil",
			externals: nil,
			lookup:    "abc",
			err:       "cannot find external property 'abc'",
			expected:  "",
		},
		{
			name:      "externals empty",
			externals: map[string]string{},
			lookup:    "efg",
			err:       "cannot find external property 'efg'",
			expected:  "",
		},
		{
			name:      "not found",
			externals: map[string]string{"abc": "abc"},
			lookup:    "efg",
			err:       "cannot find external property 'efg'",
			expected:  "",
		},
		{
			name:      "found",
			externals: map[string]string{"abc": "123"},
			lookup:    "abc",
			err:       "",
			expected:  "123",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			v, err := external(
				&transformctx.Ctx{ExternalProperties: test.externals},
				test.lookup,
			)
			switch {
			case strs.IsStrNonBlank(test.err):
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", v)
			default:
				assert.NoError(t, err)
				assert.Equal(t, test.expected, v)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	for _, test := range []struct {
		name          string
		value         string
		decimalPlaces string
		err           string
		expected      string
	}{
		{
			name:          "invalid value",
			value:         "??",
			decimalPlaces: "2",
			err:           `unable to parse value '??' to float64: strconv.ParseFloat: parsing "??": invalid syntax`,
			expected:      "",
		},
		{
			name:          "invalid decimal place value",
			value:         "3.1415926",
			decimalPlaces: "??",
			err:           `unable to parse decimal place value '??' to int: strconv.Atoi: parsing "??": invalid syntax`,
			expected:      "",
		},
		{
			name:          "decimal places less than 0",
			value:         "3.1415926",
			decimalPlaces: "-1",
			err:           `decimal place value must be an integer with range of [0,100], instead, got -1`,
			expected:      "",
		},
		{
			name:          "decimal places > 100",
			value:         "3.1415926",
			decimalPlaces: "101",
			err:           `decimal place value must be an integer with range of [0,100], instead, got 101`,
			expected:      "",
		},
		{
			name:          "decimal places less than available digits",
			value:         "3.1415926",
			decimalPlaces: "2",
			err:           "",
			expected:      "3.14",
		},
		{
			name:          "decimal places 0",
			value:         "3.1415926",
			decimalPlaces: "0",
			err:           "",
			expected:      "3",
		},
		{
			name:          "decimal places more than available digits",
			value:         "3.1415926",
			decimalPlaces: "20",
			err:           "",
			expected:      "3.1415926",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := floor(nil, test.value, test.decimalPlaces)
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

func TestIfElse(t *testing.T) {
	for _, test := range []struct {
		name     string
		cv       []string
		err      string
		expected string
	}{
		{
			name: "wrong arg number - 0",
			cv:   nil,
			err:  "arg number must be odd, but got: 0",
		},
		{
			name: "wrong arg number - even",
			cv:   []string{"a", "b"},
			err:  "arg number must be odd, but got: 2",
		},
		{
			name: "condition not bool string",
			cv:   []string{"not bool string", "b", "c"},
			err:  "condition argument must be a boolean string, but got: not bool string",
		},
		{
			name:     "one of the conditions met",
			cv:       []string{"false", "abc", "true", "123", "rest"},
			err:      "",
			expected: "123",
		},
		{
			name:     "none of the conditions met",
			cv:       []string{"false", "abc", "false", "123", "rest"},
			err:      "",
			expected: "rest",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := ifElse(nil, test.cv...)
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

func TestIfEmpty(t *testing.T) {
	r, err := isEmpty(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "true", r)
	r, err = isEmpty(nil, "   ")
	assert.NoError(t, err)
	assert.Equal(t, "false", r)
	r, err = isEmpty(nil, "abc")
	assert.NoError(t, err)
	assert.Equal(t, "false", r)
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
		name     string
		s        string
		sep      string
		trim     string
		expected string
		err      string
	}{
		{
			name:     "both empty",
			s:        "",
			sep:      "",
			trim:     "",
			expected: "",
			err:      `'sep' can't be empty`,
		},
		{
			name:     "s with several spaces and sep empty",
			s:        "  ",
			sep:      "",
			trim:     "true",
			expected: "",
			err:      `'sep' can't be empty`,
		},
		{
			name:     "s with several spaces and sep space",
			s:        "  ",
			sep:      " ",
			trim:     "true",
			expected: `["","",""]`,
			err:      "",
		},
		{
			name:     "s with several spaces and sep non-space",
			s:        "  ",
			sep:      ",",
			trim:     "true",
			expected: `[""]`,
			err:      "",
		},
		{
			name:     "s empty",
			s:        "",
			sep:      ",",
			trim:     "true",
			expected: "[]",
			err:      "",
		},
		{
			name:     "sep empty",
			s:        "ab c",
			sep:      "",
			trim:     "true",
			expected: "",
			err:      `'sep' can't be empty`,
		},
		{
			name:     "both not empty; no trim",
			s:        "a>b> c",
			sep:      ">",
			trim:     "",
			expected: `["a","b"," c"]`, // reason is: if sep is empty, s is split into each char.
			err:      "",
		},
		{
			name:     "both not empty; trim",
			s:        "a>b> c",
			sep:      ">",
			trim:     "true",
			expected: `["a","b","c"]`, // reason is: if sep is empty, s is split into each char.
			err:      "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := splitIntoJsonArray(nil, test.s, test.sep, test.trim)
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
		name       string
		str        string
		startIndex string
		lengthStr  string
		expected   string
		err        string
	}{
		{
			name:       "invalid startIndex",
			str:        "123456",
			startIndex: "abc",
			lengthStr:  "5",
			expected:   "",
			err:        `unable to convert start index 'abc' into int, err: strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:       "invalid lengthStr",
			str:        "123456",
			startIndex: "5",
			lengthStr:  "abc",
			expected:   "",
			err:        `unable to convert length 'abc' into int, err: strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:       "empty startIndex",
			str:        "123456",
			startIndex: "",
			lengthStr:  "5",
			expected:   "",
			err:        `unable to convert start index '' into int, err: strconv.Atoi: parsing "": invalid syntax`,
		},
		{
			name:       "empty lengthStr",
			str:        "123456",
			startIndex: "5",
			lengthStr:  "",
			expected:   "",
			err:        `unable to convert length '' into int, err: strconv.Atoi: parsing "": invalid syntax`,
		},
		{
			name:       "empty str",
			str:        "",
			startIndex: "0",
			lengthStr:  "0",
			expected:   "",
			err:        "",
		},
		{
			name:       "empty str with non-0 startIndex",
			str:        "",
			startIndex: "1",
			lengthStr:  "0",
			expected:   "",
			err:        `start index 1 is out of bounds (string length is 0)`,
		},
		{
			name:       "empty str with non-0 lengthStr",
			str:        "",
			startIndex: "0",
			lengthStr:  "1",
			expected:   "",
			err:        `start 0 + length 1 is out of bounds (string length is 0)`,
		},
		{
			name:       "0 startIndex",
			str:        "123456",
			startIndex: "0",
			lengthStr:  "4",
			expected:   "1234",
			err:        "",
		},
		{
			name:       "lengthStr is 1",
			str:        "123456",
			startIndex: "4",
			lengthStr:  "1",
			expected:   "5",
			err:        "",
		},
		{
			name:       "lengthStr is 0",
			str:        "123456",
			startIndex: "1",
			lengthStr:  "0",
			expected:   "",
			err:        "",
		},
		{
			name:       "lengthStr is -1",
			str:        "123456",
			startIndex: "3",
			lengthStr:  "-1",
			expected:   "456",
			err:        "",
		},
		{
			name:       "negative startIndex",
			str:        "123456",
			startIndex: "-4",
			lengthStr:  "4",
			expected:   "",
			err:        `start index -4 is out of bounds (string length is 6)`,
		},
		{
			name:       "negative lengthStr other than -1",
			str:        "123456",
			startIndex: "4",
			lengthStr:  "-2",
			expected:   "",
			err:        `length must be >= -1, but got -2`,
		},
		{
			name:       "out-of-bounds startIndex",
			str:        "123456",
			startIndex: "9",
			lengthStr:  "2",
			expected:   "",
			err:        `start index 9 is out of bounds (string length is 6)`,
		},
		{
			name:       "out-of-bounds lengthStr",
			str:        "123456",
			startIndex: "2",
			lengthStr:  "7",
			expected:   "",
			err:        `start 2 + length 7 is out of bounds (string length is 6)`,
		},
		{
			name:       "out-of-bounds startIndex and lengthStr",
			str:        "123456",
			startIndex: "10",
			lengthStr:  "9",
			expected:   "",
			err:        `start index 10 is out of bounds (string length is 6)`,
		},
		{
			name:       "substring starts at the beginning",
			str:        "123456",
			startIndex: "0",
			lengthStr:  "4",
			expected:   "1234",
			err:        "",
		},
		{
			name:       "substring ends at the end",
			str:        "123456",
			startIndex: "2",
			lengthStr:  "4",
			expected:   "3456",
			err:        "",
		},
		{
			name:       "substring starts at the end",
			str:        "123456",
			startIndex: "6",
			lengthStr:  "0",
			expected:   "",
			err:        "",
		},
		{
			name:       "substring ends at the beginning",
			str:        "123456",
			startIndex: "0",
			lengthStr:  "0",
			expected:   "",
			err:        "",
		},
		{
			name:       "substring is the whole string",
			str:        "123456",
			startIndex: "0",
			lengthStr:  "6",
			expected:   "123456",
			err:        "",
		},
		{
			name:       "non-ASCII string",
			str:        "ü:ü",
			startIndex: "1",
			lengthStr:  "2",
			expected:   ":ü",
			err:        "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := substring(nil, test.str, test.startIndex, test.lengthStr)
			if test.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", result)
			}
		})
	}
}

func TestSwitchFunc(t *testing.T) {
	for _, test := range []struct {
		name         string
		expr         string
		casesReturns []string
		err          string
		expected     string
	}{
		{
			name:         "empty casesReturns",
			expr:         "abc",
			casesReturns: nil,
			err:          "length of 'casesReturns' must be odd, but got: 0",
		},
		{
			name:         "even casesReturns length",
			expr:         "abc",
			casesReturns: []string{"1", "2", "3", "4"},
			err:          "length of 'casesReturns' must be odd, but got: 4",
		},
		{
			name:         "no case, just default",
			expr:         "abc",
			casesReturns: []string{"default"},
			expected:     "default",
		},
		{
			name: "case string contains special characters",
			expr: "How do you do",
			casesReturns: []string{
				"How do you do?", "Wrong",
				"How do you do", "Correct",
				"Huh"},
			expected: "Correct",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := switchFunc(nil, test.expr, test.casesReturns...)
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

func TestSwitchByPattern(t *testing.T) {
	for _, test := range []struct {
		name            string
		expr            string
		patternsReturns []string
		err             string
		expected        string
	}{
		{
			name:            "empty patternsReturns",
			expr:            "abc",
			patternsReturns: nil,
			err:             "length of 'patternsReturns' must be odd, but got: 0",
		},
		{
			name:            "even patternsReturns length",
			expr:            "abc",
			patternsReturns: []string{"1", "2", "3", "4"},
			err:             "length of 'patternsReturns' must be odd, but got: 4",
		},
		{
			name:            "regex invalid",
			expr:            "abc",
			patternsReturns: []string{"[", "2", "3"},
			err:             "invalid pattern '[', err: error parsing regexp: missing closing ]: `[`",
		},
		{
			name:            "no pattern, only default",
			expr:            "abc",
			patternsReturns: []string{"default"},
			expected:        "default",
		},
		{
			name: "case string contains special characters",
			expr: "2019/02/23",
			patternsReturns: []string{
				"^[0-9]{2}/[0-9]{2}/[0-9]{4}$", "Wrong",
				"^[0-9]{4}/[0-9]{2}/[0-9]{2}$", "Correct",
				"Huh"},
			expected: "Correct",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := switchByPattern(nil, test.expr, test.patternsReturns...)
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

func TestUpper(t *testing.T) {
	s, err := upper(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "", s)

	s, err = upper(nil, "abCeDfG 0123456789")
	assert.NoError(t, err)
	assert.Equal(t, "ABCEDFG 0123456789", s)
}

func TestUUIDV3(t *testing.T) {
	result, err := uuidv3(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "4ae71336-e44b-39bf-b9d2-752e234818a5", result)

	result, err = uuidv3(nil, "abc")
	assert.NoError(t, err)
	assert.Equal(t, "522ec739-ca63-3ec5-b082-08ce08ad65e2", result)
}

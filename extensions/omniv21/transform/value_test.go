package transform

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	for _, test := range []struct {
		name     string
		v        interface{}
		expected bool
	}{
		{
			name:     "map nil",
			v:        map[string]string(nil),
			expected: true,
		},
		{
			name:     "map empty",
			v:        map[string]string{},
			expected: true,
		},
		{
			name:     "slice nil",
			v:        []interface{}(nil),
			expected: true,
		},
		{
			name:     "slice empty",
			v:        []interface{}{},
			expected: true,
		},
		{
			name:     "array empty",
			v:        [0]interface{}{},
			expected: true,
		},
		{
			name:     "string empty",
			v:        "",
			expected: true,
		},
		{
			name:     "interface nil",
			v:        error(nil),
			expected: false,
		},
		{
			name:     "interface non-nil",
			v:        errors.New("test"),
			expected: false,
		},
		{
			name:     "int",
			v:        1,
			expected: false,
		},
		{
			name:     "int64",
			v:        int64(1),
			expected: false,
		},
		{
			name:     "uint32",
			v:        uint32(1),
			expected: false,
		},
		{
			name:     "rune",
			v:        'a',
			expected: false,
		},
		{
			name:     "float64",
			v:        3.1415926,
			expected: false,
		},
		{
			name:     "non-empty map",
			v:        map[string]string{"a": "1"},
			expected: false,
		},
		{
			name:     "non-empty slice",
			v:        []int{1},
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, isEmpty(test.v))
		})
	}
}

func TestResultTypeConversion(t *testing.T) {
	for _, test := range []struct {
		name     string
		v        interface{}
		typ      resultType
		err      string
		expected interface{}
	}{
		{
			name:     "int8 -> int8",
			v:        int8(64),
			typ:      resultTypeInt,
			err:      "",
			expected: int8(64),
		},
		{
			name:     "int64 -> int64",
			v:        int64(64),
			typ:      resultTypeInt,
			err:      "",
			expected: int64(64),
		},
		{
			name:     "int -> float64",
			v:        64,
			typ:      resultTypeFloat,
			err:      "",
			expected: float64(64),
		},
		{
			name:     "int32 -> string",
			v:        int32(64),
			typ:      resultTypeString,
			err:      "",
			expected: "64",
		},
		{
			name:     "uint -> uint",
			v:        uint(64),
			typ:      resultTypeInt,
			err:      "",
			expected: uint(64),
		},
		{
			name:     "uint -> float64",
			v:        uint(64),
			typ:      resultTypeFloat,
			err:      "",
			expected: float64(64),
		},
		{
			name:     "uint32 -> string",
			v:        uint(64),
			typ:      resultTypeString,
			err:      "",
			expected: "64",
		},
		{
			name:     "float64 -> int",
			v:        float64(3.1415),
			typ:      resultTypeInt,
			err:      "",
			expected: int64(3),
		},
		{
			name:     "float32 -> float32",
			v:        float32(3.14),
			typ:      resultTypeFloat,
			err:      "",
			expected: float32(3.14),
		},
		{
			name:     "float64 -> string",
			v:        float64(3.1415926),
			typ:      resultTypeString,
			err:      "",
			expected: "3.1415926",
		},
		{
			name:     "bool -> bool",
			v:        true,
			typ:      resultTypeBoolean,
			err:      "",
			expected: true,
		},
		{
			name:     "bool -> string",
			v:        false,
			typ:      resultTypeString,
			err:      "",
			expected: "false",
		},
		{
			name:     "string -> int, failure",
			v:        "not an int",
			typ:      resultTypeInt,
			err:      `strconv.ParseInt: parsing "not an int": invalid syntax`,
			expected: int64(0),
		},
		{
			name:     "string -> int, success",
			v:        "123456789",
			typ:      resultTypeInt,
			err:      "",
			expected: int64(123456789),
		},
		{
			name:     "string -> float, failure",
			v:        "not a float",
			typ:      resultTypeFloat,
			err:      `strconv.ParseFloat: parsing "not a float": invalid syntax`,
			expected: float64(0),
		},
		{
			name:     "string -> float, success",
			v:        "3.1415926",
			typ:      resultTypeFloat,
			err:      "",
			expected: float64(3.1415926),
		},
		{
			name:     "string -> bool, failure",
			v:        "123",
			typ:      resultTypeBoolean,
			err:      `strconv.ParseBool: parsing "123": invalid syntax`,
			expected: false,
		},
		{
			name:     "string -> bool, success",
			v:        "true",
			typ:      resultTypeBoolean,
			err:      ``,
			expected: true,
		},
		{
			name:     "string -> string",
			v:        "1234567890",
			typ:      resultTypeString,
			err:      ``,
			expected: "1234567890",
		},
		{
			name:     "map -> string, failure",
			v:        map[string]string{},
			typ:      resultTypeString,
			err:      errTypeConversionNotSupported.Error(),
			expected: nil,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, err := resultTypeConversion(test.v, test.typ)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, test.expected, r)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, r)
			}
		})
	}
}

func TestNormalizeAndSaveValue(t *testing.T) {
	for _, test := range []struct {
		name               string
		decl               *Decl
		value              interface{}
		expectedValue      interface{}
		expectedSaveCalled bool
		expectedErr        string
	}{
		{
			name:               "nil value with KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              nil,
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "nil value with KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              nil,
			expectedValue:      nil,
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "non string value saved",
			decl:               &Decl{},
			value:              123.45,
			expectedValue:      123.45,
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is string and NoTrim false",
			decl:               &Decl{},
			value:              " test  ",
			expectedValue:      "test",
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is blank string, NoTrim false, and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              "    ",
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "value is empty string, KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              "",
			expectedValue:      "",
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is string and NoTrim true",
			decl:               &Decl{NoTrim: true},
			value:              " test  ",
			expectedValue:      " test  ",
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name: "value is string but can't convert to result type",
			decl: &Decl{
				ResultType: testResultType(resultTypeInt),
				fqdn:       "test_fqdn",
			},
			value:              "abc",
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        `unable to convert value 'abc' to type 'int' on 'test_fqdn', err: strconv.ParseInt: parsing "abc": invalid syntax`,
		},
		{
			name:               "value is empty slice and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              []interface{}{},
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "value is empty slice and KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              []interface{}{},
			expectedValue:      []interface{}{},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is non-empty slice and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              []interface{}{"string1"},
			expectedValue:      []interface{}{"string1"},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is empty map and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              map[string]interface{}{},
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "value is empty map and KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              map[string]interface{}{},
			expectedValue:      map[string]interface{}{},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is non-empty map and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              map[string]interface{}{"test_key": "test_value"},
			expectedValue:      map[string]interface{}{"test_key": "test_value"},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			saveCalled := false
			err := normalizeAndSaveValue(test.decl, test.value, func(normalizedValue interface{}) {
				saveCalled = true
				assert.Equal(t, test.expectedValue, normalizedValue)
			})
			assert.Equal(t, test.expectedSaveCalled, saveCalled)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			}
		})
	}
}

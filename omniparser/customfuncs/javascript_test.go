package customfuncs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/nodes"
)

func TestParseArgTypeAndValue(t *testing.T) {
	for _, test := range []struct {
		name    string
		argDecl string
		argVal  string
		argName string
		argV    interface{}
		err     string
	}{
		{
			name:    "decl mal-formatted",
			argDecl: "test",
			err:     "arg decl must be in format of '<arg_name>:<arg_type>', instead got 'test'",
		},
		{
			name:    "decl contains blank arg name",
			argDecl: ":test",
			err:     "arg_name in '<arg_name>:<arg_type>' cannot be a blank string, instead got ':test'",
		},
		{
			name:    "decl contains unsupported arg type",
			argDecl: "test:what",
			err:     "arg_type 'what' in '<arg_name>:<arg_type>' is not supported",
		},
		{
			name:    "arg value is not float",
			argDecl: "arg:float",
			argVal:  "not a float",
			err:     `strconv.ParseFloat: parsing "not a float": invalid syntax`,
		},
		{
			name:    "arg value is not int",
			argDecl: "arg:int",
			argVal:  "not a int",
			err:     `strconv.ParseFloat: parsing "not a int": invalid syntax`,
		},
		{
			name:    "arg value is not boolean",
			argDecl: "arg:boolean",
			argVal:  "not a boolean",
			err:     `strconv.ParseBool: parsing "not a boolean": invalid syntax`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n, v, err := parseArgTypeAndValue(test.argDecl, test.argVal)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", n)
				assert.Nil(t, v)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.argName, n)
				assert.Equal(t, test.argV, v)
			}
		})
	}
}

func TestJavascript(t *testing.T) {
	sp, err := nodes.NewJSONStreamParser(strings.NewReader(`
		{
			"a": "one",
			"b": 2
		}`),
		".")
	assert.NoError(t, err)
	testNode, err := sp.Read()
	assert.NoError(t, err)

	for _, test := range []struct {
		name        string
		js          string
		args        []string
		expectedErr string
		expected    string
	}{
		// all success cases
		{
			name:        "no args",
			js:          "1+2+3+4",
			args:        nil,
			expectedErr: "",
			expected:    "10",
		},
		{
			name:        "with args but not using _node",
			js:          "(F - 32) * 5 / 9",
			args:        []string{"F:float", "104"},
			expectedErr: "",
			expected:    "40",
		},
		{
			name:        "with args and use _node",
			js:          "var n = JSON.parse(_node); n.a + delim + n.b",
			args:        []string{"delim:string", "-"},
			expectedErr: "",
			expected:    "one-2",
		},
		// all error cases
		{
			name:        "num of args wrong",
			js:          "",
			args:        []string{"delim:string"},
			expectedErr: "invalid number of args to 'javascript'",
			expected:    "",
		},
		{
			name:        "arg parsing wrong",
			js:          "",
			args:        []string{"delim:what", "124"},
			expectedErr: "arg_type 'what' in '<arg_name>:<arg_type>' is not supported",
			expected:    "",
		},
		{
			name:        "invalid javascript",
			js:          "var;",
			args:        nil,
			expectedErr: "SyntaxError: (anonymous): Line 1:4 Unexpected token ; (and 1 more errors)",
			expected:    "",
		},
		{
			name:        "result undefined",
			js:          "",
			args:        nil,
			expectedErr: "result is undefined",
			expected:    "",
		},
		{
			name:        "result NaN",
			js:          "0/0",
			args:        nil,
			expectedErr: "result is NaN",
			expected:    "",
		},
		{
			name:        "result null",
			js:          "null",
			args:        nil,
			expectedErr: "result is null",
			expected:    "",
		},
		{
			name:        "result infinity",
			js:          "Infinity + 3",
			args:        nil,
			expectedErr: "result is Infinity",
			expected:    "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ret, err := javascript(nil, testNode, test.js, test.args...)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", ret)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, ret)
			}
		})
	}
}

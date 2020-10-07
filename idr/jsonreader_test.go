package idr

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONStreamReader(t *testing.T) {
	for _, test := range []struct {
		name     string
		js       string
		xpath    string
		err      string
		expected []string
	}{
		{
			name:     "invalid xpath",
			js:       ``,
			xpath:    "[invalid",
			err:      `invalid xpath '[invalid', err: expression must evaluate to a node-set`,
			expected: nil,
		},
		{
			name:     "str on root",
			js:       `"test"`,
			xpath:    ".",
			err:      ``,
			expected: []string{`"test"`},
		},
		{
			name:     "num on root",
			js:       `3.1415`,
			xpath:    ".",
			err:      ``,
			expected: []string{`3.1415`},
		},
		{
			name:     "boolean on root",
			js:       `true`,
			xpath:    ".",
			err:      ``,
			expected: []string{`true`},
		},
		{
			name:     "null on root",
			js:       `null`,
			xpath:    ".",
			err:      ``,
			expected: []string{`null`},
		},
		{
			name:     "empty obj on root",
			js:       `{}`,
			xpath:    "/",
			err:      ``,
			expected: []string{`{}`},
		},
		{
			name:     "non-empty obj on root",
			js:       `{"a":1, "b":true, "c":[], "d":{}, "e":null}`,
			xpath:    ".",
			err:      ``,
			expected: []string{`{"a":1,"b":true,"c":[],"d":{},"e":null}`},
		},
		{
			name:     "empty arr on root",
			js:       `[]`,
			xpath:    ".",
			err:      ``,
			expected: []string{`[]`},
		},
		{
			name:     "non-empty arr on root",
			js:       `["1", 3.14, true, [], null, {"a":"b"}]`,
			xpath:    ".",
			err:      ``,
			expected: []string{`["1",3.14,true,[],null,{"a":"b"}]`},
		},
		{
			name:     "non-trivial xpath on root obj",
			js:       `{"a":1, "b":true, "c":[], "d":{}, "e":null}`,
			xpath:    "/a",
			err:      ``,
			expected: []string{`1`},
		},
		{
			name:     "non-trivial xpath arr on root",
			js:       `["1", 3.14, true, [], null, {"a":"b"}]`,
			xpath:    "/*",
			err:      ``,
			expected: []string{`"1"`, `3.14`, `true`, `[]`, `null`, `{"a":"b"}`},
		},
		{
			name:     "xpath and filter both effective",
			js:       `{"a":1, "b":true, "c":[], "d":{}, "e":null}`,
			xpath:    "/*[. != '']",
			err:      ``,
			expected: []string{`1`, `true`},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sp, err := NewJSONStreamReader(strings.NewReader(test.js+"\n"), test.xpath)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, sp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, sp.AtLine())
				for {
					n, err := sp.Read()
					if err == io.EOF {
						assert.Equal(t, 0, len(test.expected))
						break
					}
					assert.NoError(t, err)
					assert.True(t, len(test.expected) > 0)
					assert.Equal(t, test.expected[0], JSONify2(n))
					test.expected = test.expected[1:]
				}
				assert.Equal(t, 2, sp.AtLine())
			}
		})
	}
}

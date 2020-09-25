package customfuncs

import (
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/nodes"
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
	sp, err := nodes.NewJSONStreamReader(strings.NewReader(`
		{
			"a": "one",
			"b": 2
		}`),
		".")
	assert.NoError(t, err)
	testNode, err := sp.Read()
	assert.NoError(t, err)

	for _, test := range []struct {
		name     string
		js       string
		args     []string
		err      string
		expected string
	}{
		// all success cases
		{
			name:     "no args",
			js:       "1+2+3+4",
			args:     nil,
			err:      "",
			expected: "10",
		},
		{
			name:     "with args but not using _node",
			js:       "(F - 32) * 5 / 9",
			args:     []string{"F:float", "104"},
			err:      "",
			expected: "40",
		},
		{
			name:     "with args and use _node",
			js:       "var n = JSON.parse(_node); n.a + delim + n.b",
			args:     []string{"delim:string", "-"},
			err:      "",
			expected: "one-2",
		},
		// all error cases
		{
			name:     "num of args wrong",
			js:       "",
			args:     []string{"delim:string"},
			err:      "invalid number of args to 'javascript'",
			expected: "",
		},
		{
			name:     "arg parsing wrong",
			js:       "",
			args:     []string{"delim:what", "124"},
			err:      "arg_type 'what' in '<arg_name>:<arg_type>' is not supported",
			expected: "",
		},
		{
			name:     "invalid javascript",
			js:       "var;",
			args:     nil,
			err:      "SyntaxError: (anonymous): Line 1:4 Unexpected token ; (and 1 more errors)",
			expected: "",
		},
		{
			name:     "result undefined",
			js:       "",
			args:     nil,
			err:      "result is undefined",
			expected: "",
		},
		{
			name:     "result NaN",
			js:       "0/0",
			args:     nil,
			err:      "result is NaN",
			expected: "",
		},
		{
			name:     "result null",
			js:       "null",
			args:     nil,
			err:      "result is null",
			expected: "",
		},
		{
			name:     "result infinity",
			js:       "Infinity + 3",
			args:     nil,
			err:      "result is Infinity",
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ret, err := javascript(nil, testNode, test.js, test.args...)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", ret)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, ret)
			}
		})
	}
}

// The following benchmarks compare 'ifElse' vs 'eval' vs 'javascript' and show that
// while the flexibility and power increase from left to right, the performance
// decreases, dramatically:
//
// BenchmarkIfElse-4       	 4385941	       270 ns/op	     133 B/op	       2 allocs/op
// BenchmarkEval-4         	  642870	      1843 ns/op	     576 B/op	      11 allocs/op
// BenchmarkJavascript-4   	    5566	    214070 ns/op	  136756 B/op	    1704 allocs/op
//
// So use 'javascript' only when expression/logic complexity warrants the performance tradeoff.
var (
	benchTitles  = []string{"", "Dr", "Sir"}
	benchNames   = []string{"", "Jane", "John"}
	benchResults = []string{"", "Dr Jane", "Sir John"}
)

func BenchmarkIfElse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		title := benchTitles[i%len(benchTitles)]
		name := benchNames[i%len(benchNames)]
		titleEmpty, err := isEmpty(nil, title)
		if err != nil {
			b.FailNow()
		}
		nameEmpty, err := isEmpty(nil, name)
		if err != nil {
			b.FailNow()
		}
		titleAndName, err := concat(nil, title, " ", name)
		if err != nil {
			b.FailNow()
		}
		ret, err := ifElse(nil, titleEmpty, "", nameEmpty, "", titleAndName)
		if err != nil {
			b.FailNow()
		}
		if ret != benchResults[i%len(benchResults)] {
			b.FailNow()
		}
	}
}

func BenchmarkEval(b *testing.B) {
	for i := 0; i < b.N; i++ {
		title := benchTitles[i%len(benchTitles)]
		name := benchNames[i%len(benchNames)]
		titleAndName, err := concat(nil, title, " ", name)
		if err != nil {
			b.FailNow()
		}
		ret, err := eval(nil,
			"([title]== '' || [name] == '') ? '' : [titleAndName]",
			"title:string", title,
			"name:string", name,
			"titleAndName:string", titleAndName)
		if err != nil {
			b.FailNow()
		}
		if ret != benchResults[i%len(benchResults)] {
			b.FailNow()
		}
	}
}

func BenchmarkJavascript(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ret, err := javascript(nil, &node.Node{}, `
			if (!title) {
				""
			} else if (!name) {
				""
			} else {
				title + " " + name
			}`,
			"title:string", benchTitles[i%len(benchTitles)],
			"name:string", benchNames[i%len(benchNames)])
		if err != nil {
			b.FailNow()
		}
		if ret != benchResults[i%len(benchResults)] {
			b.FailNow()
		}
	}
}

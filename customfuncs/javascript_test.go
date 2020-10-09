package customfuncs

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
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

const (
	noCache   = false
	withCache = true
)

func prepCachesForTest(cache bool) {
	disableCaching = !cache
	resetCaches()
}

func TestJavascript(t *testing.T) {
	sp, err := idr.NewJSONStreamReader(strings.NewReader(`
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
			err:      "invalid javascript: SyntaxError: (anonymous): Line 1:4 Unexpected token ; (and 1 more errors)",
			expected: "",
		},
		{
			name:     "javascript throws",
			js:       "throw 'failure';",
			args:     nil,
			err:      "failure at <eval>:1:7(1)",
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
		testFn := func(t *testing.T) {
			var ret string
			var err error
			if strings.Contains(test.js, "_node") {
				ret, err = javascriptWithContext(nil, testNode, test.js, test.args...)
			} else {
				ret, err = javascript(nil, test.js, test.args...)
			}
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", ret)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, ret)
			}
		}
		t.Run(test.name+" (without cache)", func(t *testing.T) {
			prepCachesForTest(noCache)
			testFn(t)
		})
		t.Run(test.name+" (with cache)", func(t *testing.T) {
			prepCachesForTest(withCache)
			resetCaches()
			testFn(t)
		})
	}
}

func TestJavascriptClearVarsAfterRunProgram(t *testing.T) {
	prepCachesForTest(noCache)
	r, err := javascript(nil, `v1 + v2`, "v1:int", "1", "v2:int", "2")
	assert.NoError(t, err)
	assert.Equal(t, "3", r)
	// Note v1 should be cleared before second run.
	r, err = javascript(nil, `v3 + v4 + v1`, "v3:int", "10", "v4:int", "20")
	assert.Error(t, err)
	assert.Equal(t, `ReferenceError: v1 is not defined at <eval>:1:11(3)`, err.Error())
	assert.Equal(t, "", r)
	// Run again without using v1.
	r, err = javascript(nil, `v3 + v4`, "v3:int", "10", "v4:int", "20")
	assert.NoError(t, err)
	assert.Equal(t, "30", r)
}

// go test -bench=. -benchmem -benchtime=30s
// BenchmarkIfElse-4                            	352522934	        97.9 ns/op	       8 B/op	       1 allocs/op
// BenchmarkEval-4                              	25887961	      1405 ns/op	     418 B/op	       8 allocs/op
// BenchmarkJavascriptWithNoCache-4             	  172342	    210129 ns/op	  136607 B/op	    1698 allocs/op
// BenchmarkJavascriptWithCache-4               	15996846	      2062 ns/op	     128 B/op	       8 allocs/op
// BenchmarkConcurrentJavascriptWithNoCache-4   	    1143	  33140598 ns/op	27328719 B/op	  339653 allocs/op
// BenchmarkConcurrentJavascriptWithCache-4     	   65140	    543091 ns/op	   26608 B/op	    1746 allocs/op

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

func benchmarkJavascript(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ret, err := javascript(nil, `
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

func BenchmarkJavascriptWithNoCache(b *testing.B) {
	prepCachesForTest(noCache)
	benchmarkJavascript(b)
}

func BenchmarkJavascriptWithCache(b *testing.B) {
	prepCachesForTest(withCache)
	benchmarkJavascript(b)
}

func concurrentBenchmarkJavascript(b *testing.B) {
	concurrency := 200
	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			index := i
			go func() {
				defer wg.Done()
				ret, err := javascript(nil, `
					if (!title) {
						""
					} else if (!name) {
						""
					} else {
						title + " " + name
					}`,
					"title:string", benchTitles[index%len(benchTitles)],
					"name:string", benchNames[index%len(benchNames)])
				if err != nil {
					b.FailNow()
				}
				if ret != benchResults[index%len(benchResults)] {
					b.FailNow()
				}
			}()
		}
		wg.Wait()
	}
}

func BenchmarkConcurrentJavascriptWithNoCache(b *testing.B) {
	prepCachesForTest(noCache)
	concurrentBenchmarkJavascript(b)
}

func BenchmarkConcurrentJavascriptWithCache(b *testing.B) {
	prepCachesForTest(withCache)
	concurrentBenchmarkJavascript(b)
}

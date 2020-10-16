package customfuncs

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

const (
	noCache   = false
	withCache = true
)

func prepCachesForTest(cache bool) {
	disableCaching = !cache
	resetCaches()
}

func TestJavaScript(t *testing.T) {
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
		args     []interface{}
		err      string
		expected interface{}
	}{
		// all success cases
		{
			name:     "no args",
			js:       "1+2+3+4",
			args:     nil,
			err:      "",
			expected: int64(10),
		},
		{
			name:     "with args but not using _node",
			js:       "(F - 32) * 5 / 9",
			args:     []interface{}{"F", 104},
			err:      "",
			expected: int64(40),
		},
		{
			name:     "with args and use _node",
			js:       "var n = JSON.parse(_node); n.a + delim + n.b",
			args:     []interface{}{"delim", "-"},
			err:      "",
			expected: "one-2",
		},
		// all error cases
		{
			name:     "num of args wrong",
			js:       "",
			args:     []interface{}{"delim"},
			err:      "number of args must be even, but got 1",
			expected: nil,
		},
		{
			name:     "invalid javascript",
			js:       "var;",
			args:     nil,
			err:      "invalid javascript: SyntaxError: (anonymous): Line 1:4 Unexpected token ; (and 1 more errors)",
			expected: nil,
		},
		{
			name:     "javascript throws",
			js:       "throw 'failure';",
			args:     nil,
			err:      "failure at <eval>:1:7(1)",
			expected: nil,
		},
		{
			name:     "result undefined",
			js:       "",
			args:     nil,
			err:      "result is undefined",
			expected: nil,
		},
		{
			name:     "result NaN",
			js:       "0/0",
			args:     nil,
			err:      "result is NaN",
			expected: nil,
		},
		{
			name:     "result null",
			js:       "null",
			args:     nil,
			err:      "result is null",
			expected: nil,
		},
		{
			name:     "result infinity",
			js:       "Infinity + 3",
			args:     nil,
			err:      "result is Infinity",
			expected: nil,
		},
	} {
		testFn := func(t *testing.T) {
			var ret interface{}
			var err error
			if strings.Contains(test.js, "_node") {
				ret, err = JavaScriptWithContext(nil, testNode, test.js, test.args...)
			} else {
				ret, err = JavaScript(nil, test.js, test.args...)
			}
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, ret)
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

func TestJavaScriptClearVarsAfterRunProgram(t *testing.T) {
	prepCachesForTest(noCache)
	r, err := JavaScript(nil, `v1 + v2`, "v1", 1, "v2", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), r)
	// Note v1 should be cleared before second run.
	r, err = JavaScript(nil, `v3 + v4 + v1`, "v3", 10, "v4", 20)
	assert.Error(t, err)
	assert.Equal(t, `ReferenceError: v1 is not defined at <eval>:1:11(3)`, err.Error())
	assert.Nil(t, r)
	// Run again without using v1.
	r, err = JavaScript(nil, `v3 + v4`, "v3", 10, "v4", 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(30), r)
}

// go test -bench=. -benchmem -benchtime=30s
// BenchmarkJavaScriptWithNoCache-8             	  225940	    160696 ns/op	  136620 B/op	    1698 allocs/op
// BenchmarkJavaScriptWithCache-8               	22289469	      1612 ns/op	     140 B/op	       9 allocs/op
// BenchmarkConcurrentJavaScriptWithNoCache-8   	    1479	  24486799 ns/op	27331471 B/op	  339793 allocs/op
// BenchmarkConcurrentJavaScriptWithCache-8     	   69219	    517898 ns/op	   34722 B/op	    1952 allocs/op

var (
	benchTitles  = []string{"", "Dr", "Sir"}
	benchNames   = []string{"", "Jane", "John"}
	benchResults = []string{"", "Dr Jane", "Sir John"}
)

func benchmarkJavaScript(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ret, err := JavaScript(nil, `
			if (!title) {
				""
			} else if (!name) {
				""
			} else {
				title + " " + name
			}`,
			"title", benchTitles[i%len(benchTitles)],
			"name", benchNames[i%len(benchNames)])
		if err != nil {
			b.FailNow()
		}
		if ret != benchResults[i%len(benchResults)] {
			b.FailNow()
		}
	}
}

func BenchmarkJavaScriptWithNoCache(b *testing.B) {
	prepCachesForTest(noCache)
	benchmarkJavaScript(b)
}

func BenchmarkJavaScriptWithCache(b *testing.B) {
	prepCachesForTest(withCache)
	benchmarkJavaScript(b)
}

func concurrentBenchmarkJavaScript(b *testing.B) {
	concurrency := 200
	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			index := i
			go func() {
				defer wg.Done()
				ret, err := JavaScript(nil, `
					if (!title) {
						""
					} else if (!name) {
						""
					} else {
						title + " " + name
					}`,
					"title", benchTitles[index%len(benchTitles)],
					"name", benchNames[index%len(benchNames)])
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

func BenchmarkConcurrentJavaScriptWithNoCache(b *testing.B) {
	prepCachesForTest(noCache)
	concurrentBenchmarkJavaScript(b)
}

func BenchmarkConcurrentJavaScriptWithCache(b *testing.B) {
	prepCachesForTest(withCache)
	concurrentBenchmarkJavaScript(b)
}

package nodes

import (
	"encoding/json"
	"io"
	"sort"
	"strconv"
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"
)

// Note parseJSON and parseJSONValue are borrowed from
// https://github.com/antchfx/jsonquery/blob/master/node.go as a reference
// implementation for verification purpose.
func parseJSONValue(x interface{}, parent *node.Node) {
	switch v := x.(type) {
	case []interface{}:
		for _, vv := range v {
			n := &node.Node{Type: node.ElementNode}
			node.AddChild(parent, n)
			parseJSONValue(vv, n)
		}
	case map[string]interface{}:
		var keys []string
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			n := &node.Node{Data: key, Type: node.ElementNode}
			node.AddChild(parent, n)
			parseJSONValue(v[key], n)
		}
	case string:
		node.AddChild(parent, &node.Node{Data: v, Type: node.TextNode})
	case float64:
		// The format fmt with 'f' means (-ddd.dddd, no exponent),
		// The special precision -1 uses the smallest number of digits
		s := strconv.FormatFloat(v, 'f', -1, 64)
		node.AddChild(parent, &node.Node{Data: s, Type: node.TextNode})
	case bool:
		s := strconv.FormatBool(v)
		node.AddChild(parent, &node.Node{Data: s, Type: node.TextNode})
	}
}

func parseJSON(b []byte) (*node.Node, error) {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	doc := &node.Node{Type: node.DocumentNode}
	parseJSONValue(v, doc)
	return doc, nil
}

func TestParseAgainstReference(t *testing.T) {
	for _, test := range []struct {
		name string
		js   string
	}{
		{
			name: "root value - string",
			js:   `"123"`,
		},
		{
			name: "root value - int",
			js:   `123`,
		},
		{
			name: "root value - float",
			js:   `3.14159265358`,
		},
		{
			name: "root value - bool",
			js:   `true`,
		},
		{
			name: "root value - nil",
			js:   `null`,
		},
		{
			name: "empty obj",
			js:   `{}`,
		},
		{
			name: "simple obj",
			js: `
				{
					"a": 123,
					"b": true,
					"c": "test"
				}`,
		},
		{
			name: "complex obj",
			js: `
				{
					"a": {
						"b": [
							{
								"c": "1"
							},
							2,
							"d"
						],
						"e": "3"
					},
					"d": null
				}`,
		},
		{
			name: "empty arr",
			js:   `[]`,
		},
		{
			name: "simple arr",
			js:   `[123, "123", false, true, null]`,
		},
		{
			name: "complex arr",
			js: `
				[
					{"a":null, "b":123, "c":true},
					123,
					[
						[1, 2, 3], 
						"four"
					]
				]`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, err := NewJSONStreamReader(strings.NewReader(test.js), "non-matching")
			assert.NoError(t, err)
			_, err = r.parse()
			assert.Equal(t, io.EOF, err)

			n, err := parseJSON([]byte(test.js))
			assert.NoError(t, err)

			assert.Equal(t, n.OutputXML(true), r.root.n.OutputXML(true))
		})
	}
}

func TestJNodeType(t *testing.T) {
	jn := &jnode{}
	for _, test := range []struct {
		name      string
		jnodeType jnodeType
		isTypeFn  func() bool
	}{
		{
			name:      "jnodeTypeRoot",
			jnodeType: jnodeTypeRoot,
			isTypeFn:  jn.isRoot,
		},
		{
			name:      "jnodeTypeObj",
			jnodeType: jnodeTypeObj,
			isTypeFn:  jn.isObj,
		},
		{
			name:      "jnodeTypeArr",
			jnodeType: jnodeTypeArr,
			isTypeFn:  jn.isArr,
		},
		{
			name:      "jnodeTypeProp",
			jnodeType: jnodeTypeProp,
			isTypeFn:  jn.isProp,
		},
		{
			name:      "jnodeTypeVal",
			jnodeType: jnodeTypeVal,
			isTypeFn:  jn.isVal,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			jn.jnodeType = 0
			jn.jnodeType = test.jnodeType
			assert.True(t, test.isTypeFn())
			jn.jnodeType = 0
			jn.jnodeType = jnodeType(^uint(0))
			assert.True(t, test.isTypeFn())
			jn.jnodeType = 0
			jn.jnodeType = ^test.jnodeType
			assert.False(t, test.isTypeFn())
		})
	}
}

func TestStream_ArrOfObj(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		[
			{
				"a": 123,
				"b": "john"
			},
			{
				"a": 456,
				"b": "jane"
			},
			"extra",
			{
				"a": 123,
				"b": "smith"
			}
		]`), "/*[a=123]")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `<><a>123</a><b>john</b></>`, n.OutputXML(true))
	assert.Equal(t, `<><><a>123</a><b>john</b></></>`, r.root.n.OutputXML(true))
	assert.True(t, r.AtLine() >= 6)

	n, err = r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `<><a>123</a><b>smith</b></>`, n.OutputXML(true))
	// this is to verify we've deleted the previous stream node so memory won't
	// grow unbounded in streaming mode.
	assert.Equal(t, `<><><a>123</a><b>smith</b></></>`, r.root.n.OutputXML(true))
	assert.True(t, r.AtLine() >= 15)

	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_RootObjMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		{
			"a": "123"
		}`), ".")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `<><a>123</a></>`, n.OutputXML(true))

	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_RootObjNotMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		{
			"a": "123"
		}`), ".[a!='123']")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_RootValMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`"abc"`), ".")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `<>abc</>`, n.OutputXML(true))

	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_RootValNotMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`"abc"`), ".[text()!='abc']")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_ObjPropMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		{
			"a": "123",
			"b": "456"
		}`),
		"a")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `<a>123</a>`, n.OutputXML(true))
	// Just so that one realizes if you stream on a prop, then the object itself may be incomplete
	// at the time of Read() call returns. This following assert shows.
	assert.Equal(t, `<><a>123</a></>`, n.Parent.OutputXML(true))

	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_ObjPropNotMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		{
			"a": "123",
			"b": "456"
		}`),
		"a[. != '123']")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_ArrValueMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		{
			"a": [
				"123",
				"abc"
			]
		}`),
		"a/*[.='123']")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `<>123</>`, n.OutputXML(true))

	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestStream_ArrValueNotMatch(t *testing.T) {
	r, err := NewJSONStreamReader(strings.NewReader(`
		{
			"a": [
				"123",
				"abc"
			]
		}`),
		"a/*[.='xyz']")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestNewJSONStreamParser_InvalidXPath(t *testing.T) {
	r, err := NewJSONStreamReader(nil, ">")
	assert.Error(t, err)
	assert.Equal(t, "invalid xpath '>', err: expression must evaluate to a node-set", err.Error())
	assert.Nil(t, r)
}

package nodes

import (
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"
)

func xmlToTestNode(t *testing.T, xpath, xmlStr string) *node.Node {
	p, err := node.CreateStreamParser(strings.NewReader(xmlStr), xpath)
	assert.NoError(t, err)
	n, err := p.Read()
	assert.NoError(t, err)
	return n
}

func jsonToTestNode(t *testing.T, xpath, jsonStr string) *node.Node {
	p, err := NewJSONStreamReader(strings.NewReader(jsonStr), xpath)
	assert.NoError(t, err)
	n, err := p.Read()
	assert.NoError(t, err)
	return n
}

func TestIsChildText(t *testing.T) {
	for _, test := range []struct {
		name       string
		xpath      string
		xmlStr     string
		isTextNode bool
	}{
		{
			name:       "xml: child is text node",
			xpath:      "a",
			xmlStr:     `<a>text</a>`,
			isTextNode: true,
		},
		{
			name:  "xml: child is array",
			xpath: "a",
			xmlStr: `<a>
						<b>1</b>
						<b>2</b>
					</a>`,
			isTextNode: false,
		},
		{
			name:  "xml: child is object",
			xpath: "a",
			xmlStr: `<a>
						<b>1</b>
						<c>2</c>
					</a>`,
			isTextNode: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := xmlToTestNode(t, test.xpath, test.xmlStr)
			assert.Equal(t, test.isTextNode, isChildText(n))
		})
	}

	for _, test := range []struct {
		name       string
		xpath      string
		jsonStr    string
		isTextNode bool
	}{
		{
			name:       "json: child is text node",
			xpath:      "a",
			jsonStr:    `{ "a": "text" }`,
			isTextNode: true,
		},
		{
			name:       "json: child is array",
			xpath:      "a",
			jsonStr:    `{ "a": [ "1", "2" ] }`,
			isTextNode: false,
		},
		{
			name:       "json: child is object",
			xpath:      "a",
			jsonStr:    `{ "a": { "b": "c" } }`,
			isTextNode: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := jsonToTestNode(t, test.xpath, test.jsonStr)
			assert.Equal(t, test.isTextNode, isChildText(n))
		})
	}
}

func TestIsChildArray(t *testing.T) {
	for _, test := range []struct {
		name    string
		xpath   string
		xmlStr  string
		isArray bool
	}{
		{
			name:    "xml: child is text node",
			xpath:   "a",
			xmlStr:  `<a>text</a>`,
			isArray: false,
		},
		{
			name:  "xml: child is array",
			xpath: "a",
			xmlStr: `<a>
                        <b>1</b>
                        <b>2</b>
                     </a>`,
			isArray: true,
		},
		{
			name:  "xml: child is object with multiple elements",
			xpath: "a",
			xmlStr: `<a>
                        <b>1</b>
                        <b>2</b>
                        <c>3</c>
                     </a>`,
			isArray: false,
		},
		{
			name:  "xml: child is object with single element",
			xpath: "a",
			xmlStr: `<a>
                        <b>1</b>
                     </a>`,
			isArray: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := xmlToTestNode(t, test.xpath, test.xmlStr)
			assert.Equal(t, test.isArray, isChildArray(n))
		})
	}

	for _, test := range []struct {
		name    string
		xpath   string
		jsonStr string
		isArray bool
	}{
		{
			name:    "json: child is text node",
			xpath:   "a",
			jsonStr: `{ "a": "text" }`,
			isArray: false,
		},
		{
			name:    "json: child is array",
			xpath:   "a",
			jsonStr: `{ "a": [ "1", "2" ] }`,
			isArray: true,
		},
		{
			name:    "json: child is object",
			xpath:   "a",
			jsonStr: `{ "a": { "b": "c" } }`,
			isArray: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := jsonToTestNode(t, test.xpath, test.jsonStr)
			assert.Equal(t, test.isArray, isChildArray(n))
		})
	}
}

func TestJ2NodeToInterface_NoChild(t *testing.T) {
	assert.Nil(t, J2NodeToInterface(&node.Node{Type: node.ElementNode, Data: "a"}))
}

func TestJ2NodeToInterface_ChildIsText(t *testing.T) {
	assert.Equal(t, "1", J2NodeToInterface(xmlToTestNode(t, "a/b", "<a><b>1</b></a>")))
}

func TestJ2NodeToInterface_ChildIsArray(t *testing.T) {
	// Note the test case below that `"c": 2` the numeric value in json, but when we parse into node
	// structure, every value is string. That typically isn't a problem since usually when we parseField
	// we can specify result_type = int/float to force value type conversion. But in result_type = object
	// case we can't do that anymore. So this is a known limitation.
	assert.Equal(t,
		[]interface{}{
			"1",
			map[string]interface{}{
				"c": "2",
			},
			[]interface{}{"3", "4"},
		},
		J2NodeToInterface(jsonToTestNode(t, "a/b",
			`{
                "a": {
                    "b": [
                        "1",
                        {
                            "c": 2
                        },
                        [ "3", "4" ]
                    ]
                }
            }`)))

	// Testing json array with single element to see if isChildArray will mistake it as object or not.
	assert.Equal(t,
		[]interface{}{"1"},
		J2NodeToInterface(jsonToTestNode(t, "a/b",
			`{
                "a": {
                    "b": [ "1" ]
                }
            }`)))

	// Testing xml array with single element to see if isChildArray will mistake it as array or not.
	assert.Equal(t,
		map[string]interface{}{
			"b": "1",
		},
		J2NodeToInterface(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
            </a>`)))

	// Testing xml array with multiple elements
	assert.Equal(t,
		[]interface{}{"1", "2", "3"},
		J2NodeToInterface(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
                <b>2</b>
                <b>3</b>
            </a>`)))
}

func TestJ2NodeToInterface_ChildIsObject(t *testing.T) {
	assert.Equal(t,
		map[string]interface{}{
			"b": []interface{}{"1", "2"},
			"c": "3",
			"d": map[string]interface{}{
				"e": "four",
			},
		},
		J2NodeToInterface(jsonToTestNode(t, "a",
			`{
                "a": {
                    "b": [ "1", "2" ],
                    "c": 3,
                    "d": {
                        "e": "four"
                    }
                }
            }`)))

	// Testing xml child object with conflict names getting overwritten.
	assert.Equal(t,
		map[string]interface{}{
			"b": "2",
			"c": "3",
		},
		J2NodeToInterface(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
                <b>2</b>
                <c>3</c>
            </a>`)))
}

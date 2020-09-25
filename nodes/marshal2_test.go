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
		name     string
		xpath    string
		xml      string
		expected bool
	}{
		{
			name:     "is text",
			xpath:    "x",
			xml:      `<x>t</x>`,
			expected: true,
		},
		{
			name:  "is array",
			xpath: "x",
			// important to keep it multi-lined so dummy text nodes will be created and tested.
			xml: `<x>
					<y>1</y>
					<y>2</y>
				</x>`,
			expected: false,
		},
		{
			name:     "is object",
			xpath:    "x",
			xml:      `<x><y>1</y><z>2</z></x>`,
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := xmlToTestNode(t, test.xpath, test.xml)
			assert.Equal(t, test.expected, isChildText(n))
		})
	}

	for _, test := range []struct {
		name     string
		xpath    string
		json     string
		expected bool
	}{
		{
			name:     "is text",
			xpath:    "x",
			json:     `{ "x": "123" }`,
			expected: true,
		},
		{
			name:     "is array",
			xpath:    "x",
			json:     `{ "x": [ "a", "b" ] }`,
			expected: false,
		},
		{
			name:     "is object",
			xpath:    "x",
			json:     `{ "x": { "y": "z" } }`,
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := jsonToTestNode(t, test.xpath, test.json)
			assert.Equal(t, test.expected, isChildText(n))
		})
	}
}

func TestIsChildArray(t *testing.T) {
	for _, test := range []struct {
		name     string
		xpath    string
		xml      string
		expected bool
	}{
		{
			name:     "is text node",
			xpath:    "x",
			xml:      `<x>abc</x>`,
			expected: false,
		},
		{
			name:  "is array",
			xpath: "x",
			xml: `<x>
					<y>1</y>
                    <y>2</y>
				</x>`,
			expected: true,
		},
		{
			name:  "is object with multiple elements",
			xpath: "x",
			xml: `<x>
					<y>a</y>
                    <y>b</y>
                    <z>c</z>
				</x>`,
			expected: false,
		},
		{
			name:  "is object with single element",
			xpath: "x",
			xml: `<x>
					<y>a</y>
				</x>`,
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := xmlToTestNode(t, test.xpath, test.xml)
			assert.Equal(t, test.expected, isChildArray(n))
		})
	}

	for _, test := range []struct {
		name     string
		xpath    string
		json     string
		expected bool
	}{
		{
			name:     "is text node",
			xpath:    "x",
			json:     `{ "x": "abc" }`,
			expected: false,
		},
		{
			name:     "is array",
			xpath:    "x",
			json:     `{ "x": [ "a", "b" ] }`,
			expected: true,
		},
		{
			name:     "is object",
			xpath:    "x",
			json:     `{ "x": { "y": "z" } }`,
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := jsonToTestNode(t, test.xpath, test.json)
			assert.Equal(t, test.expected, isChildArray(n))
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
	// Pay attention that the original value of "c" field in JSON is numeric 2, but after
	// the conversion, it becomes "2" string. This is an unfortunately minor side effect
	// due to everything we parse in and stored in node.Node is string. Lost in translation.
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

	assert.Equal(t,
		[]interface{}{"1"},
		J2NodeToInterface(jsonToTestNode(t, "a/b",
			`{
                "a": {
                    "b": [ "1" ]
                }
            }`)))

	assert.Equal(t,
		map[string]interface{}{
			"b": "1",
		},
		J2NodeToInterface(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
            </a>`)))

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

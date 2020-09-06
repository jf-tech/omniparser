package transform

import (
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"
)

// TODO: we don't have json stream parser ready yet. Leave out all the json testing.

func xmlToTestNode(t *testing.T, xpath, xmlStr string) *node.Node {
	p, err := node.CreateStreamParser(strings.NewReader(xmlStr), xpath)
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
}

func TestNodeToObject_NoChild(t *testing.T) {
	assert.Nil(t, nodeToObject(&node.Node{Type: node.ElementNode, Data: "a"}))
}

func TestNodeToObject_ChildIsText(t *testing.T) {
	assert.Equal(t, "1", nodeToObject(xmlToTestNode(t, "a/b", "<a><b>1</b></a>")))
}

func TestNodeToObject_ChildIsArray(t *testing.T) {
	// Testing xml array with single element to see if isChildArray will mistake it as array or not.
	assert.Equal(t,
		map[string]interface{}{
			"b": "1",
		},
		nodeToObject(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
            </a>`)))

	// Testing xml array with multiple elements
	assert.Equal(t,
		[]interface{}{"1", "2", "3"},
		nodeToObject(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
                <b>2</b>
                <b>3</b>
            </a>`)))
}

func TestNodeToObject_ChildIsObject(t *testing.T) {
	// Testing xml child object with conflict names getting overwritten.
	assert.Equal(t,
		map[string]interface{}{
			"b": "2",
			"c": "3",
		},
		nodeToObject(xmlToTestNode(t, "a",
			`<a>
                <b>1</b>
                <b>2</b>
                <c>3</c>
            </a>`)))
}

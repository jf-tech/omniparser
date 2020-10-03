package idr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJ1NodePtrName(t *testing.T) {
	for _, test := range []struct {
		name     string
		n        *Node
		expected string
	}{
		{name: "nil", n: nil, expected: ""},
		{name: "root", n: CreateNode(DocumentNode, "test"), expected: "(DocumentNode)"},
		{name: "elem w/o ns", n: CreateNode(ElementNode, "A"), expected: "(ElementNode A)"},
		{name: "elem w/ ns", n: CreateXMLNode(ElementNode, "A", XMLSpecific{"ns", "uri://"}), expected: "(ElementNode ns:A)"},
		{name: "text", n: CreateNode(TextNode, "data"), expected: "(TextNode 'data')"},
		{name: "attr", n: CreateNode(AttributeNode, "attr"), expected: "(AttributeNode attr)"},
		{name: "unknown", n: CreateNode(NodeType(99999), "what"), expected: "(unknown 'what')"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.expected == "" {
				assert.Nil(t, j1NodePtrName(test.n))
			} else {
				assert.Equal(t, test.expected, *j1NodePtrName(test.n))
			}
		})
	}
}

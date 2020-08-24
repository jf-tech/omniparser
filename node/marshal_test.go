package node

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestNodeNameForJSONMarshaler(t *testing.T) {
	for _, test := range []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name:     "nil",
			node:     nil,
			expected: "(nil)",
		},
		{
			name:     "root",
			node:     &Node{Type: DocumentNode},
			expected: "(ROOT)",
		},
		{
			name:     "element with no namespace",
			node:     &Node{Type: ElementNode, Data: "abc"},
			expected: "(ELEM abc)",
		},
		{
			name:     "element with namespace",
			node:     &Node{Type: ElementNode, Data: "abc", Prefix: "ns"},
			expected: "(ELEM ns:abc)",
		},
		{
			name:     "text",
			node:     &Node{Type: TextNode, Data: "abc"},
			expected: "(TEXT 'abc')",
		},
		{
			name:     "attr",
			node:     &Node{Type: AttributeNode, Data: "abc"},
			expected: "(ATTR abc)",
		},
		{
			name:     "attr",
			node:     &Node{Type: AttributeNode, Data: "abc"},
			expected: "(ATTR abc)",
		},
		{
			name:     "unknown",
			node:     &Node{Type: NodeType(99999), Data: "abc"},
			expected: "(UNKNOWN type:(unknown:99999) data:abc)",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, nodeNameForJSONMarshaler(test.node))
		})
	}
}

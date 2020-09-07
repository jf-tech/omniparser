package nodes

import (
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"
)

func TestNodeName(t *testing.T) {
	for _, test := range []struct {
		name     string
		n        *node.Node
		expected string
	}{
		{name: "nil", n: nil, expected: "(nil)"},
		{name: "root", n: &node.Node{Type: node.DocumentNode}, expected: "(ROOT)"},
		{name: "decl", n: &node.Node{Type: node.DeclarationNode, Data: "xml"}, expected: "(DECL xml)"},
		{name: "elem w/o ns", n: &node.Node{Type: node.ElementNode, Data: "A"}, expected: "(ELEM A)"},
		{name: "elem w/ ns", n: &node.Node{Type: node.ElementNode, Data: "A", Prefix: "ns"}, expected: "(ELEM ns:A)"},
		{name: "text", n: &node.Node{Type: node.TextNode, Data: "data"}, expected: "(TEXT 'data')"},
		{name: "cdata", n: &node.Node{Type: node.CharDataNode, Data: "data"}, expected: "(CDATA 'data')"},
		{name: "comment", n: &node.Node{Type: node.CommentNode, Data: "huh"}, expected: "(COMMENT 'huh')"},
		{name: "attr", n: &node.Node{Type: node.AttributeNode, Data: "attr"}, expected: "(ATTR attr)"},
		{name: "unknown", n: &node.Node{Type: node.NodeType(99999), Data: "what"}, expected: "(unknown 'what')"},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, nodeName(test.n))
		})
	}
}

func TestNodeTypeStr(t *testing.T) {
	assert.Equal(t, "DocumentNode", nodeTypeStr(node.DocumentNode))
	assert.Equal(t, "DeclarationNode", nodeTypeStr(node.DeclarationNode))
	assert.Equal(t, "ElementNode", nodeTypeStr(node.ElementNode))
	assert.Equal(t, "TextNode", nodeTypeStr(node.TextNode))
	assert.Equal(t, "CharDataNode", nodeTypeStr(node.CharDataNode))
	assert.Equal(t, "CommentNode", nodeTypeStr(node.CommentNode))
	assert.Equal(t, "AttributeNode", nodeTypeStr(node.AttributeNode))
	assert.Equal(t, "(unknown:99999)", nodeTypeStr(node.NodeType(99999)))
}

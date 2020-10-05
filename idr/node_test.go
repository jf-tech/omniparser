package idr

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestNodeType_String(t *testing.T) {
	assert.Equal(t, "DocumentNode", DocumentNode.String())
	assert.Equal(t, "ElementNode", ElementNode.String())
	assert.Equal(t, "TextNode", TextNode.String())
	assert.Equal(t, "AttributeNode", AttributeNode.String())
	assert.Equal(t, "(unknown NodeType: 99)", NodeType(99).String())
}

func rootOf(n *Node) *Node {
	for ; n != nil && n.Parent != nil; n = n.Parent {
	}
	return n
}

func checkPointersInTree(t *testing.T, n *Node) {
	if n == nil {
		return
	}
	if n.FirstChild != nil {
		assert.True(t, n == n.FirstChild.Parent)
	}
	if n.LastChild != nil {
		assert.True(t, n == n.LastChild.Parent)
	}
	checkPointersInTree(t, n.FirstChild)
	// There is no need to call checkPointersInTree(t, n.LastChild)
	// because checkPointersInTree(t, n.FirstChild) will traverse all its
	// siblings to the end, and if the last one isn't n.LastChild then it will fail.
	parent := n.Parent // could be nil if n is the root of a tree.
	// Verify the PrevSibling chain
	cur, prev := n, n.PrevSibling
	for ; prev != nil; cur, prev = prev, prev.PrevSibling {
		assert.True(t, prev.Parent == parent)
		assert.True(t, prev.NextSibling == cur)
	}
	assert.True(t, cur.PrevSibling == nil)
	assert.True(t, parent == nil || parent.FirstChild == cur)
	// Verify the NextSibling chain
	cur, next := n, n.NextSibling
	for ; next != nil; cur, next = next, next.NextSibling {
		assert.True(t, next.Parent == parent)
		assert.True(t, next.PrevSibling == cur)
	}
	assert.True(t, cur.NextSibling == nil)
	assert.True(t, parent == nil || parent.LastChild == cur)
}

type testTree struct {
	//
	// root
	//   elemA                elemB         elemC
	//     elemA1,elemA2        elemB1        attrC1,attrC2,elemC3,elemC4
	//       textA1,textA2        textB1        textC1,textC2,textC3,textC4
	root *Node

	elemA          *Node
	elemA1, textA1 *Node
	elemA2, textA2 *Node

	elemB          *Node
	elemB1, textB1 *Node

	elemC          *Node
	attrC1, textC1 *Node
	attrC2, textC2 *Node
	elemC3, textC3 *Node
	elemC4, textC4 *Node
}

const (
	testTreeXML    = true
	testTreeNotXML = false
)

func newTestTree(t *testing.T, xmlNode bool) *testTree {
	mkNode := func(parent *Node, ntype NodeType, name string) *Node {
		var node *Node
		if xmlNode {
			node = CreateXMLNode(ntype, name, XMLSpecific{})
		} else {
			node = CreateNode(ntype, name)
		}
		if parent != nil {
			AddChild(parent, node)
		}
		return node
	}
	mkPair := func(parent *Node, ntype NodeType, name, text string) (*Node, *Node) {
		node := mkNode(parent, ntype, name)
		child := mkNode(node, TextNode, text)
		return node, child
	}

	root := mkNode(nil, DocumentNode, "root")

	elemA := mkNode(root, ElementNode, "elemA")
	elemA1, textA1 := mkPair(elemA, ElementNode, "elemA1", "textA1")
	elemA2, textA2 := mkPair(elemA, ElementNode, "elemA2", "textA2")

	elemB := mkNode(root, ElementNode, "elemB")
	elemB1, textB1 := mkPair(elemB, ElementNode, "elemB1", "textB1")

	elemC := mkNode(root, ElementNode, "elemC")
	attrC1, textC1 := mkPair(elemC, AttributeNode, "attrC1", "textC1")
	attrC2, textC2 := mkPair(elemC, AttributeNode, "attrC2", "textC2")
	elemC3, textC3 := mkPair(elemC, ElementNode, "elemC3", "textC3")
	elemC4, textC4 := mkPair(elemC, ElementNode, "elemC4", "textC4")

	checkPointersInTree(t, root)

	return &testTree{
		root: root,

		elemA:  elemA,
		elemA1: elemA1, textA1: textA1,
		elemA2: elemA2, textA2: textA2,

		elemB:  elemB,
		elemB1: elemB1, textB1: textB1,

		elemC:  elemC,
		attrC1: attrC1, textC1: textC1,
		attrC2: attrC2, textC2: textC2,
		elemC3: elemC3, textC3: textC3,
		elemC4: elemC4, textC4: textC4,
	}
}

func TestDumpTestTree(t *testing.T) {
	cupaloy.SnapshotT(t, JSONify1(newTestTree(t, testTreeXML).root))
}

func TestInnerText(t *testing.T) {
	tt := newTestTree(t, testTreeNotXML)
	assert.Equal(t, tt.textA1.Data+tt.textA2.Data, tt.elemA.InnerText())
	// Note attribute's texts are skipped in InnerText(), by design.
	assert.Equal(t, tt.textC3.Data+tt.textC4.Data, tt.elemC.InnerText())
}

func TestRemoveNodeAndSubTree(t *testing.T) {
	t.Run("remove a node who is its parents only child", func(t *testing.T) {
		tt := newTestTree(t, testTreeXML)
		RemoveFromTree(tt.elemB1)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a node who is its parents first child but not the last", func(t *testing.T) {
		tt := newTestTree(t, testTreeXML)
		RemoveFromTree(tt.elemA)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a node who is its parents middle child not the first not the last", func(t *testing.T) {
		tt := newTestTree(t, testTreeXML)
		RemoveFromTree(tt.elemB)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a node who is its parents last child but not the first", func(t *testing.T) {
		tt := newTestTree(t, testTreeXML)
		RemoveFromTree(tt.elemC)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a root does nothing", func(t *testing.T) {
		tt := newTestTree(t, testTreeXML)
		RemoveFromTree(tt.root)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})
}

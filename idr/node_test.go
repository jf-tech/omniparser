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

func findRoot(n *Node) *Node {
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
	//   child1                              child2                 child3
	//     grandChild11E, grandchild12E        grandChild21E          grandChild31A
	//       grandChild11T, grandchild12T        grandChild21T          grandChild31T
	root                         *Node
	child1, child2, child3       *Node
	grandChild11E, grandChild11T *Node
	grandChild12E, grandChild12T *Node
	grandChild21E, grandChild21T *Node
	grandChild31A, grandChild31T *Node
}

func newTestTree(t *testing.T) *testTree {
	root := CreateNode(DocumentNode, "root")
	child1 := CreateNode(ElementNode, "child1")
	child2 := CreateNode(ElementNode, "child2")
	child3 := CreateNode(ElementNode, "child3")
	grandChild11E := CreateNode(ElementNode, "grandChild11")
	grandChild11T := CreateNode(TextNode, "data 11")
	grandChild12E := CreateNode(ElementNode, "grandChild12")
	grandChild12T := CreateNode(TextNode, "data 12")
	grandChild21E := CreateNode(ElementNode, "grandChild21")
	grandChild21T := CreateNode(TextNode, "data 21")
	grandChild31A := CreateNode(AttributeNode, "grandChild31")
	grandChild31T := CreateNode(TextNode, "attr 31")

	AddChild(root, child1)
	AddChild(root, child2)
	AddChild(root, child3)
	AddChild(child1, grandChild11E)
	AddChild(child1, grandChild12E)
	AddChild(child2, grandChild21E)
	AddChild(child3, grandChild31A)
	AddChild(grandChild11E, grandChild11T)
	AddChild(grandChild12E, grandChild12T)
	AddChild(grandChild21E, grandChild21T)
	AddChild(grandChild31A, grandChild31T)

	checkPointersInTree(t, root)
	checkPointersInTree(t, child1)
	checkPointersInTree(t, child2)
	checkPointersInTree(t, child3)
	checkPointersInTree(t, grandChild11E)
	checkPointersInTree(t, grandChild12E)
	checkPointersInTree(t, grandChild21E)
	checkPointersInTree(t, grandChild31A)
	checkPointersInTree(t, grandChild11T)
	checkPointersInTree(t, grandChild12T)
	checkPointersInTree(t, grandChild21T)
	checkPointersInTree(t, grandChild31T)

	return &testTree{
		root:          root,
		child1:        child1,
		child2:        child2,
		child3:        child3,
		grandChild11E: grandChild11E,
		grandChild12E: grandChild12E,
		grandChild21E: grandChild21E,
		grandChild31A: grandChild31A,
		grandChild11T: grandChild11T,
		grandChild12T: grandChild12T,
		grandChild21T: grandChild21T,
		grandChild31T: grandChild31T,
	}
}

func TestReferenceTestTreeWithJSONify1(t *testing.T) {
	cupaloy.SnapshotT(t, JSONify1(newTestTree(t).root))
}

func TestInnerText(t *testing.T) {
	tt := newTestTree(t)
	assert.Equal(t, tt.grandChild11T.Data+tt.grandChild12T.Data, tt.child1.InnerText())
	assert.Equal(t, tt.grandChild11T.Data+tt.grandChild12T.Data+tt.grandChild21T.Data, tt.root.InnerText())
}

func TestRemoveNodeAndSubTree(t *testing.T) {
	t.Run("remove a node who is its parents only child", func(t *testing.T) {
		tt := newTestTree(t)
		RemoveFromTree(tt.grandChild21E)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a node who is its parents first child but not the last", func(t *testing.T) {
		tt := newTestTree(t)
		RemoveFromTree(tt.child1)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a node who is its parents middle child not the first not the last", func(t *testing.T) {
		tt := newTestTree(t)
		RemoveFromTree(tt.child2)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a node who is its parents last child but not the first", func(t *testing.T) {
		tt := newTestTree(t)
		RemoveFromTree(tt.child3)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})

	t.Run("remove a root does nothing", func(t *testing.T) {
		tt := newTestTree(t)
		RemoveFromTree(tt.root)
		checkPointersInTree(t, tt.root)
		cupaloy.SnapshotT(t, JSONify1(tt.root))
	})
}

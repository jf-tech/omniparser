package node

import (
	"encoding/xml"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/omniparser/jsonutil"
	"github.com/stretchr/testify/assert"
)

func TestNodeTypeString(t *testing.T) {
	var names []string
	for nt := 0; nt < int(allNodeType); nt++ {
		names = append(names, NodeType(nt).String())
	}
	cupaloy.SnapshotT(t, names)
	assert.Equal(t, "(unknown:123)", NodeType(123).String())
}

func TestNodeTypeString_Unknown(t *testing.T) {
	assert.Equal(t, "(unknown:123)", NodeType(123).String())
}

func TestNodeTypeJSONMarshal(t *testing.T) {
	assert.Equal(t, "\"DocumentNode\"\n", jsonutil.BestEffortPrettyMarshal(DocumentNode))
}

func root(n *Node) *Node {
	if n == nil {
		return nil
	}
	for ; n.Parent != nil; n = n.Parent {
	}
	return n
}

// Given a root node in a tree, verify all the pointers are valid (parent, first child, etc.)
// If you supply a node that isn't the root node, this function only verifies the subtree rooted
// by the given's node's parent, including the given node's siblings.
func verifyPointerIntegrityInTree(t *testing.T, n *Node) {
	if n == nil {
		return
	}

	if n.FirstChild != nil {
		assert.Equal(t, n, n.FirstChild.Parent)
	}

	if n.LastChild != nil {
		assert.Equal(t, n, n.LastChild.Parent)
	}

	verifyPointerIntegrityInTree(t, n.FirstChild)
	// There is no need to call verifyPointerIntegrityInTree(t, n.LastChild)
	// because verifyPointerIntegrityInTree(t, n.FirstChild) will traverse all its
	// siblings to the end, and if the last one isn't n.LastChild then it will fail.

	parent := n.Parent // could be nil if n is the root of a tree.

	// Verify the PrevSibling chain
	cur, prev := n, n.PrevSibling
	for ; prev != nil; cur, prev = prev, prev.PrevSibling {
		assert.Equal(t, prev.Parent, parent)
		assert.Equal(t, prev.NextSibling, cur)
	}
	assert.Nil(t, cur.PrevSibling)
	assert.True(t, parent == nil || parent.FirstChild == cur)

	// Verify the NextSibling chain
	cur, next := n, n.NextSibling
	for ; next != nil; cur, next = next, next.NextSibling {
		assert.Equal(t, next.Parent, parent)
		assert.Equal(t, next.PrevSibling, cur)
	}
	assert.Nil(t, cur.NextSibling)
	assert.True(t, parent == nil || parent.LastChild == cur)
}

type testTree struct {
	// root
	//   child1                              child2                    child3
	//     grandChild11, grandchild12          grandChild21
	root                                     *Node
	child1, child2, child3                   *Node
	grandChild11, grandChild12, grandChild21 *Node
}

func createTestTree(t *testing.T) *testTree {
	root := &Node{
		Type:         DocumentNode,
		Data:         "root",
		Prefix:       "prefix_r",
		NamespaceURI: "namespace_r",
		Attrs: []xml.Attr{
			{Name: xml.Name{Space: "ns", Local: "r"}, Value: "r"},
		},
		level: 0,
	}
	child1 := &Node{
		Type:         ElementNode,
		Data:         "child1",
		Prefix:       "prefix_c1",
		NamespaceURI: "namespace_c1",
		Attrs: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "c1"}, Value: "c1"},
		},
		level: 1,
	}
	child2 := &Node{
		Type:         ElementNode,
		Data:         "child2",
		Prefix:       "prefix_c2",
		NamespaceURI: "namespace_c2",
		Attrs: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "c2"}, Value: "c2"},
		},
		level: 1,
	}
	child3 := &Node{
		Type:         ElementNode,
		Data:         "child3",
		Prefix:       "prefix_c3",
		NamespaceURI: "namespace_c3",
		Attrs: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "c3"}, Value: "c3"},
		},
		level: 1,
	}
	grandChild11 := &Node{
		Type:  TextNode,
		Data:  "grandChild11",
		level: 2,
	}
	grandChild12 := &Node{
		Type:  TextNode,
		Data:  "grandChild12",
		level: 2,
	}
	grandChild21 := &Node{
		Type:  TextNode,
		Data:  "grandChild21",
		level: 2,
	}

	// root -> child1
	AddChild(root, child1)
	// root -> child1, child2
	AddChild(root, child2)
	// root -> child1, child2, child3. Note child3 will be added at the end despite we call
	// AddSibling(child1, child3)
	AddSibling(child1, child3)

	AddChild(child1, grandChild11)
	AddSibling(grandChild11, grandChild12)

	AddChild(child2, grandChild21)

	verifyPointerIntegrityInTree(t, root)
	verifyPointerIntegrityInTree(t, child1)
	verifyPointerIntegrityInTree(t, child2)
	verifyPointerIntegrityInTree(t, child3)
	verifyPointerIntegrityInTree(t, grandChild11)
	verifyPointerIntegrityInTree(t, grandChild12)
	verifyPointerIntegrityInTree(t, grandChild21)

	return &testTree{
		root:         root,
		child1:       child1,
		child2:       child2,
		child3:       child3,
		grandChild11: grandChild11,
		grandChild12: grandChild12,
		grandChild21: grandChild21,
	}
}

func TestReferenceTestTree(t *testing.T) {
	cupaloy.SnapshotT(t, jsonutil.BestEffortPrettyMarshal(createTestTree(t).root))
}

func TestCopyNode(t *testing.T) {
	test := createTestTree(t)
	child2Copy := copyNode(test.child2)

	assert.Equal(t, test.child2.Type, child2Copy.Type)
	assert.Equal(t, test.child2.Data, child2Copy.Data)
	assert.Equal(t, test.child2.Prefix, child2Copy.Prefix)
	assert.Equal(t, test.child2.NamespaceURI, child2Copy.NamespaceURI)
	assert.Equal(t, test.child2.Attrs, child2Copy.Attrs)
	// this is to verify the slice Attrs is copied, not just referenced.
	child2Copy.Attrs[0].Value = "whatever"
	assert.Equal(t, "c2", test.child2.Attrs[0].Value)
	assert.Equal(t, test.child2.level, child2Copy.level)

	assert.Nil(t, child2Copy.Parent)
	assert.Nil(t, child2Copy.FirstChild)
	assert.Nil(t, child2Copy.LastChild)
	assert.Nil(t, child2Copy.PrevSibling)
	assert.Nil(t, child2Copy.NextSibling)
}

func TestInnerText(t *testing.T) {
	test := createTestTree(t)
	assert.Equal(t, test.grandChild11.Data+test.grandChild12.Data, test.child1.InnerText())
	assert.Equal(t, test.grandChild11.Data+test.grandChild12.Data+test.grandChild21.Data, test.root.InnerText())
}

func TestCopyTree(t *testing.T) {
	test := createTestTree(t)
	child2Copy := CopyTree(test.child2)
	assert.True(t, test.child2 != child2Copy)
	verifyPointerIntegrityInTree(t, child2Copy)
	assert.Equal(t, jsonutil.BestEffortPrettyMarshal(test.child2), jsonutil.BestEffortPrettyMarshal(child2Copy))

	copyRoot := root(child2Copy)
	assert.True(t, test.root != copyRoot)
	verifyPointerIntegrityInTree(t, copyRoot)
	assert.Equal(t, jsonutil.BestEffortPrettyMarshal(test.root), jsonutil.BestEffortPrettyMarshal(copyRoot))
}

func TestRemoveFromTree(t *testing.T) {
	t.Run("remove a node who is its parents only child", func(t *testing.T) {
		test := createTestTree(t)
		RemoveFromTree(test.grandChild21)
		verifyPointerIntegrityInTree(t, test.root)
		cupaloy.SnapshotT(t, jsonutil.BestEffortPrettyMarshal(test.root))
	})

	t.Run("remove a node who is its parents first child but not the last", func(t *testing.T) {
		test := createTestTree(t)
		RemoveFromTree(test.child1)
		verifyPointerIntegrityInTree(t, test.root)
		cupaloy.SnapshotT(t, jsonutil.BestEffortPrettyMarshal(test.root))
	})

	t.Run("remove a node who is its parents middle child not the first not the last", func(t *testing.T) {
		test := createTestTree(t)
		RemoveFromTree(test.child2)
		verifyPointerIntegrityInTree(t, test.root)
		cupaloy.SnapshotT(t, jsonutil.BestEffortPrettyMarshal(test.root))
	})

	t.Run("remove a node who is its parents last child but not the first", func(t *testing.T) {
		test := createTestTree(t)
		RemoveFromTree(test.child3)
		verifyPointerIntegrityInTree(t, test.root)
		cupaloy.SnapshotT(t, jsonutil.BestEffortPrettyMarshal(test.root))
	})

	t.Run("remove a root does nothing", func(t *testing.T) {
		test := createTestTree(t)
		RemoveFromTree(test.root)
		verifyPointerIntegrityInTree(t, test.root)
		cupaloy.SnapshotT(t, jsonutil.BestEffortPrettyMarshal(test.root))
	})
}

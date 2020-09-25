package nodes

import (
	"encoding/xml"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"
)

func findRoot(n *node.Node) *node.Node {
	if n == nil {
		return nil
	}
	for ; n.Parent != nil; n = n.Parent {
	}
	return n
}

func verifyPointerIntegrityInTree(t *testing.T, n *node.Node) {
	if n == nil {
		return
	}
	if n.FirstChild != nil {
		// shouldn't not use assert.Equal which actually might do DeepEqual instead of pointer value comparison
		assert.True(t, n == n.FirstChild.Parent)
	}
	if n.LastChild != nil {
		assert.True(t, n == n.LastChild.Parent)
	}
	verifyPointerIntegrityInTree(t, n.FirstChild)
	// There is no need to call verifyPointerIntegrityInTree(t, n.LastChild)
	// because verifyPointerIntegrityInTree(t, n.FirstChild) will traverse all its
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
	//   child1                              child2                    child3
	//     grandChild11E, grandchild12E          grandChild21E
	//     grandChild11T, grandchild12T          grandChild21T
	root                                        *node.Node
	child1, child2, child3                      *node.Node
	grandChild11E, grandChild12E, grandChild21E *node.Node
	grandChild11T, grandChild12T, grandChild21T *node.Node
}

func createTestTree(t *testing.T) *testTree {
	root := &node.Node{
		Type:         node.DocumentNode,
		Data:         "root",
		Prefix:       "prefix_r",
		NamespaceURI: "namespace_r",
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "ns", Local: "r"}, Value: "r"},
		},
	}
	child1 := &node.Node{
		Type:         node.ElementNode,
		Data:         "child1",
		Prefix:       "prefix_c1",
		NamespaceURI: "namespace_c1",
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "c1"}, Value: "c1"},
		},
	}
	child2 := &node.Node{
		Type:         node.ElementNode,
		Data:         "child2",
		Prefix:       "prefix_c2",
		NamespaceURI: "namespace_c2",
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "c2"}, Value: "c2"},
		},
	}
	child3 := &node.Node{
		Type:         node.ElementNode,
		Data:         "child3",
		Prefix:       "prefix_c3",
		NamespaceURI: "namespace_c3",
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "c3"}, Value: "c3"},
		},
	}
	grandChild11E := &node.Node{
		Type: node.ElementNode,
		Data: "grandChild11",
	}
	grandChild11T := &node.Node{
		Type: node.CharDataNode,
		Data: "data 11",
	}
	grandChild12E := &node.Node{
		Type: node.ElementNode,
		Data: "grandChild12",
	}
	grandChild12T := &node.Node{
		Type: node.CharDataNode,
		Data: "data 12",
	}
	grandChild21E := &node.Node{
		Type: node.ElementNode,
		Data: "grandChild21",
	}
	grandChild21T := &node.Node{
		Type: node.CharDataNode,
		Data: "data 21",
	}

	node.AddChild(root, child1)
	node.AddSibling(child1, child2)
	node.AddSibling(child1, child3)
	node.AddChild(child1, grandChild11E)
	node.AddChild(child1, grandChild12E)
	node.AddChild(child2, grandChild21E)
	node.AddChild(grandChild11E, grandChild11T)
	node.AddChild(grandChild12E, grandChild12T)
	node.AddChild(grandChild21E, grandChild21T)

	verifyPointerIntegrityInTree(t, root)
	verifyPointerIntegrityInTree(t, child1)
	verifyPointerIntegrityInTree(t, child2)
	verifyPointerIntegrityInTree(t, child3)
	verifyPointerIntegrityInTree(t, grandChild11E)
	verifyPointerIntegrityInTree(t, grandChild12E)
	verifyPointerIntegrityInTree(t, grandChild21E)
	verifyPointerIntegrityInTree(t, grandChild11T)
	verifyPointerIntegrityInTree(t, grandChild12T)
	verifyPointerIntegrityInTree(t, grandChild21T)

	return &testTree{
		root:          root,
		child1:        child1,
		child2:        child2,
		child3:        child3,
		grandChild11E: grandChild11E,
		grandChild12E: grandChild12E,
		grandChild21E: grandChild21E,
		grandChild11T: grandChild11T,
		grandChild12T: grandChild12T,
		grandChild21T: grandChild21T,
	}
}

func TestReferenceTestTreeWithJSONify1(t *testing.T) {
	cupaloy.SnapshotT(t, JSONify1(createTestTree(t).root))
}

func TestReferenceTestTreeWithJSONify2(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(JSONify2(createTestTree(t).root)))
}

func TestCopyNode(t *testing.T) {
	test := createTestTree(t)
	child2Copy := copyNode(test.child2)

	assert.Equal(t, test.child2.Type, child2Copy.Type)
	assert.Equal(t, test.child2.Data, child2Copy.Data)
	assert.Equal(t, test.child2.Prefix, child2Copy.Prefix)
	assert.Equal(t, test.child2.NamespaceURI, child2Copy.NamespaceURI)
	assert.Equal(t, test.child2.Attr, child2Copy.Attr)
	// this is to verify the slice Attrs is copied.
	child2Copy.Attr[0].Value = "whatever"
	assert.Equal(t, "c2", test.child2.Attr[0].Value)

	assert.True(t, child2Copy.Parent == nil)
	assert.True(t, child2Copy.FirstChild == nil)
	assert.True(t, child2Copy.LastChild == nil)
	assert.True(t, child2Copy.PrevSibling == nil)
	assert.True(t, child2Copy.NextSibling == nil)
}

func TestCopyTree(t *testing.T) {
	test := createTestTree(t)
	child2Copy := CopyTree(test.child2)
	assert.True(t, test.child2 != child2Copy)
	verifyPointerIntegrityInTree(t, child2Copy)
	assert.Equal(t, JSONify1(test.child2), JSONify1(child2Copy))

	rootCopy := findRoot(child2Copy)
	assert.True(t, test.root != rootCopy)
	verifyPointerIntegrityInTree(t, rootCopy)
	assert.Equal(t, JSONify1(test.root), JSONify1(rootCopy))
}

package idr

import (
	"testing"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/stretchr/testify/assert"
)

func navTestSetup(t *testing.T) (*testTree, *navigator, func(*Node)) {
	setupTestNodeCaching(testNodeCachingOff)
	tt := newTestTree(t, testTreeXML)
	nav := createNavigator(tt.root)
	moveTo := func(n *Node) { nav.cur = n }
	return tt, nav, moveTo
}

func TestCurrent(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.root)
	assert.True(t, tt.root == nav.Current())
	moveTo(tt.elemB)
	assert.True(t, tt.elemB == nav.Current())
	moveTo(tt.textC3)
	assert.True(t, tt.textC3 == nav.Current())
}

func TestNodeType(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.root)
	assert.Equal(t, xpath.RootNode, nav.NodeType())
	moveTo(tt.elemA)
	assert.Equal(t, xpath.ElementNode, nav.NodeType())
	moveTo(tt.textB1)
	assert.Equal(t, xpath.TextNode, nav.NodeType())
	moveTo(tt.attrC1)
	assert.Equal(t, xpath.AttributeNode, nav.NodeType())
	nav.cur.Type = NodeType(123)
	assert.PanicsWithValue(t, "(unknown NodeType: 123)", func() {
		nav.NodeType()
	})
}

func TestLocalName(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.root)
	assert.Equal(t, "root", nav.LocalName())
	moveTo(tt.elemA)
	assert.Equal(t, "elemA", nav.LocalName())
	moveTo(tt.attrC1)
	assert.Equal(t, "attrC1", nav.LocalName())
}

func TestPrefix(t *testing.T) {
	n := CreateNode(ElementNode, "n1")
	nav := createNavigator(n)
	assert.Equal(t, "", nav.Prefix())
	n = CreateXMLNode(ElementNode, "n2", XMLSpecific{NamespacePrefix: "ns"})
	nav = createNavigator(n)
	assert.Equal(t, "ns", nav.Prefix())
}

func TestNamespaceURL(t *testing.T) {
	n := CreateNode(ElementNode, "n1")
	nav := createNavigator(n)
	assert.Equal(t, "", nav.NamespaceURL())
	n = CreateXMLNode(ElementNode, "n2", XMLSpecific{NamespaceURI: "uri"})
	nav = createNavigator(n)
	assert.Equal(t, "uri", nav.NamespaceURL())
}

func TestValue(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.elemA)
	assert.Equal(t, tt.textA1.Data+tt.textA2.Data, nav.Value())
}

func TestCopy(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.elemC4)
	assert.True(t, tt.elemC4 == nav.cur)
	assert.True(t, tt.root == nav.root)
	navCopy := nav.Copy()
	assert.True(t, nav != navCopy.(*navigator))
	assert.True(t, nav.cur == navCopy.(*navigator).cur)
	assert.True(t, nav.root == navCopy.(*navigator).root)
}

func TestMoveToRoot(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.elemB1)
	assert.True(t, tt.elemB1 == nav.Current())
	nav.MoveToRoot()
	assert.True(t, tt.root == nav.Current())
}

func TestMoveToParent(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	assert.True(t, tt.root == nav.cur)
	assert.False(t, nav.MoveToParent())
	moveTo(tt.attrC2)
	assert.True(t, nav.MoveToParent())
	assert.True(t, tt.elemC == nav.Current())
}

func TestMoveToNextAttribute(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.elemB)
	assert.False(t, nav.MoveToNextAttribute())
	moveTo(tt.elemC)
	assert.True(t, nav.MoveToNextAttribute())
	assert.True(t, nav.Current() == tt.attrC1)
	assert.True(t, nav.MoveToNextAttribute())
	assert.True(t, nav.Current() == tt.attrC2)
	assert.False(t, nav.MoveToNextAttribute())
}

func TestMoveToChild(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.attrC1)
	assert.False(t, nav.MoveToChild())
	// remove elemC3/elemC4, so elemC now only has attrC1 and attrC2
	RemoveAndReleaseTree(tt.elemC3)
	RemoveAndReleaseTree(tt.elemC4)
	moveTo(tt.elemC)
	assert.False(t, nav.MoveToChild())
	moveTo(tt.elemB)
	assert.True(t, nav.MoveToChild())
	assert.True(t, tt.elemB1 == nav.Current())
}

func TestMoveToFirst(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.attrC2)
	assert.False(t, nav.MoveToFirst())
	// test the situation where cur is the first element, no prev sibling
	moveTo(tt.elemA)
	assert.False(t, nav.MoveToFirst())
	// test the situation where cur is the first element, and all prev siblings are attributes.
	moveTo(tt.elemC3)
	assert.False(t, nav.MoveToFirst())
	// now test the success case
	moveTo(tt.elemC)
	assert.True(t, nav.MoveToFirst())
	assert.True(t, tt.elemA == nav.Current())
}

func TestString(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.elemC)
	assert.Equal(t, tt.textC3.Data+tt.textC4.Data, nav.String())
}

func TestMoveToNext(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.attrC2)
	assert.False(t, nav.MoveToNext())
	moveTo(tt.elemB1)
	assert.False(t, nav.MoveToNext())
	moveTo(tt.elemA1)
	assert.True(t, nav.MoveToNext())
	assert.True(t, tt.elemA2 == nav.Current())
}

func TestMoveToPrevious(t *testing.T) {
	tt, nav, moveTo := navTestSetup(t)
	moveTo(tt.attrC2)
	assert.False(t, nav.MoveToPrevious())
	moveTo(tt.elemA)
	assert.False(t, nav.MoveToPrevious())
	moveTo(tt.elemC3)
	assert.False(t, nav.MoveToPrevious())
	moveTo(tt.elemC4)
	assert.True(t, nav.MoveToPrevious())
	assert.True(t, tt.elemC3 == nav.Current())
}

func TestMoveTo(t *testing.T) {
	// First test case: other nav isn't ours.
	other := &xmlquery.NodeNavigator{}
	tt, nav, _ := navTestSetup(t)
	assert.True(t, tt.root == nav.Current())
	assert.False(t, nav.MoveTo(other))
	_, nav2, _ := navTestSetup(t)
	assert.False(t, nav.MoveTo(nav2))
	nav2.root = nav.root
	nav2.cur = tt.elemB1
	assert.True(t, nav.MoveTo(nav2))
	assert.True(t, tt.elemB1 == nav.Current())
}

func TestNodeFromIter(t *testing.T) {
	tt, nav, _ := navTestSetup(t)
	iter := xpath.Select(nav, "elemC/@attrC2")
	assert.True(t, iter.MoveNext())
	assert.True(t, tt.attrC2 == nodeFromIter(iter))
}

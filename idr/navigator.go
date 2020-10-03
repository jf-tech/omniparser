package idr

import (
	"github.com/antchfx/xpath"
)

type navigator struct {
	root, cur *Node
}

func (nav *navigator) Current() *Node {
	return nav.cur
}

func (nav *navigator) NodeType() xpath.NodeType {
	switch nav.cur.Type {
	case DocumentNode:
		return xpath.RootNode
	case ElementNode:
		return xpath.ElementNode
	case TextNode:
		return xpath.TextNode
	case AttributeNode:
		return xpath.AttributeNode
	}
	panic(nav.cur.Type.String())
}

func (nav *navigator) LocalName() string {
	return nav.cur.Data
}

func (nav *navigator) Prefix() string {
	if !IsXML(nav.cur) {
		return ""
	}
	return XMLSpecificOf(nav.cur).NamespacePrefix
}

func (nav *navigator) NamespaceURL() string {
	if !IsXML(nav.cur) {
		return ""
	}
	return XMLSpecificOf(nav.cur).NamespaceURI
}

func (nav *navigator) Value() string {
	return nav.cur.InnerText()
}

func (nav *navigator) Copy() xpath.NodeNavigator {
	n := *nav
	return &n
}

func (nav *navigator) MoveToRoot() {
	nav.cur = nav.root
}

func (nav *navigator) MoveToParent() bool {
	if nav.cur.Parent == nil {
		return false
	}
	nav.cur = nav.cur.Parent
	return true
}

func (nav *navigator) MoveToNextAttribute() bool {
	if nav.cur.Type == AttributeNode {
		// if we are currently on an AttributeNode, move to the next AttributeNode if there is one.
		if nav.cur.NextSibling == nil || nav.cur.NextSibling.Type != AttributeNode {
			return false
		}
		nav.cur = nav.cur.NextSibling
		return true
	}
	// If we're not on an AttributeNode, then check its first child - because if there are
	// AttributeNodes, they will always be packed first.
	if nav.cur.FirstChild == nil || nav.cur.FirstChild.Type != AttributeNode {
		return false
	}
	nav.cur = nav.cur.FirstChild
	return true
}

func (nav *navigator) MoveToChild() bool {
	if nav.cur.Type == AttributeNode {
		// In xpath navigation, if we're on attribute, we will/should never move down to
		// child of attribute (because there is none).
		return false
	}
	n := nav.cur.FirstChild
	for ; n != nil && n.Type == AttributeNode; n = n.NextSibling {
		// this loop basically skips all the attribute nodes if any.
	}
	if n == nil {
		return false
	}
	nav.cur = n
	return true
}

func (nav *navigator) MoveToFirst() bool {
	if nav.cur.Type == AttributeNode {
		// once we're on attribute node, we're never asked to move around.
		return false
	}
	n := nav.cur
	for ; n.PrevSibling != nil && n.PrevSibling.Type != AttributeNode; n = n.PrevSibling {
		// move to the first node that is not attribute node.
	}
	if n == nav.cur {
		return false
	}
	nav.cur = n
	return true
}

func (nav *navigator) String() string {
	return nav.Value()
}

func (nav *navigator) MoveToNext() bool {
	if nav.cur.Type == AttributeNode || nav.cur.NextSibling == nil {
		// If on attribute, we never move.
		return false
	}
	nav.cur = nav.cur.NextSibling
	return true
}

func (nav *navigator) MoveToPrevious() bool {
	if nav.cur.Type == AttributeNode {
		// if on attribute, we never move.
		return false
	}
	if nav.cur.PrevSibling == nil || nav.cur.PrevSibling.Type == AttributeNode {
		return false
	}
	nav.cur = nav.cur.PrevSibling
	return true
}

func (nav *navigator) MoveTo(other xpath.NodeNavigator) bool {
	node, ok := other.(*navigator)
	if !ok || node.root != nav.root {
		return false
	}
	nav.cur = node.cur
	return true
}

// Ensure *navigator implements xpath.NodeNavigator interface
var _ xpath.NodeNavigator = &navigator{}

func createNavigator(n *Node) *navigator {
	return &navigator{root: n, cur: n}
}

func nodeFromIter(iter *xpath.NodeIterator) *Node {
	return iter.Current().(*navigator).cur
}

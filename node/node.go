package node

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

/**
Credit: much of the code in this package is copied and adapted from https://github.com/antchfx/xmlquery.
The current `xmlquery.Node` implementation largely fits our needs, however, given this `Node` and the
DOM tree are the essential intermediate in-memory representation of omniparser, we'd like to ensure the
flexibility, stability and patchability of it:
- flexibility: since this `Node` and its DOM tree will represent all formats we support: CSV/txt/EDI/XML/etc
we want it to be flexible, and we have full control of what unexposed fields we can add to it.
- stability: we don't want upstream dependency changes and/or bug fixes risk the stability of the parser, such
as new NodeType introduction or existing NodeType deprecation during parsing.
- patchability: we can patch it anytime we'd like to without dragging in other unwanted changes in upstream.
Therefore we've decided to make this "duplicate".
*/

// A NodeType is the type of a Node.
type NodeType uint

const (
	// DocumentNode is the root of the document tree, providing access to the entire document.
	DocumentNode NodeType = iota
	// ElementNode represents a node in the document tree. It represents different things for different
	// input format: it represents an element node in XML; it represents a field or an object in JSON;
	// it represents a segment or a segment group in EDI, etc.
	ElementNode
	// TextNode contains the actual data for an ElementNode or AttributeNode
	TextNode
	// AttributeNode represents an attribute in XML. Not used in other formats.
	AttributeNode

	allNodeType
)

var nodeTypeStrings = map[NodeType]string{
	DocumentNode:  "DocumentNode",
	ElementNode:   "ElementNode",
	TextNode:      "TextNode",
	AttributeNode: "AttributeNode",
}

func (nt NodeType) String() string {
	if s, found := nodeTypeStrings[nt]; found {
		return s
	}
	return fmt.Sprintf("(unknown:%d)", nt)
}

func (nt NodeType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + nt.String() + `"`), nil
}

// Node is the basic building block of the document tree. It can represent different
// part of the input data such as element, attribute, data, etc.
type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type         NodeType
	Data         string
	Prefix       string     // only valid in XML where namespaces and namespace prefixes exist.
	NamespaceURI string     // only valid in XML where namespaces and namespace prefixes exist.
	Attrs        []xml.Attr // only valid in XML

	level int // node level in the tree, only used during document tree loading phase, thus unexported.
}

// Duplicate a node without its copying pointers.
func copyNode(n *Node) *Node {
	newNode := *n
	newNode.Parent = nil
	newNode.FirstChild = nil
	newNode.LastChild = nil
	newNode.PrevSibling = nil
	newNode.NextSibling = nil
	// slice field copy is by-ref when copying a struct, so must do
	// a manual copy to ensure the separation from original node.
	if n.Attrs != nil {
		newNode.Attrs = make([]xml.Attr, len(n.Attrs))
		// xml.Attr is struct, which is copy by-value. So that's fine.
		copy(newNode.Attrs, n.Attrs)
	}
	return &newNode
}

// InnerText() returns all the raw text data within the node, concatenated recursively.
// Example:
// <book id="bk101">
//   <author>Gambardella, Matthew</author>
//   <title>XML Developer's Guide</title>
//   <genre>Computer</genre>
//   <price>44.95</price>
//   <publish_date>2000-10-01</publish_date>
//   <description>An in-depth look at creating applications
//   with XML.</description>
// </book>
// InnerText() on book ElementNode returns:
//
// Gambardella, Matthew
// XML Developer's Guide
// Computer
// 44.95
// 2000-10-01
// An in-depth look at creating applications
// with XML.
//
// InnerText() on book.author ElementNode returns:
// Gambardella, Matthew
// InnerText() on book.author TextNode (child of book.author ElementNode) returns:
// Gambardella, Matthew
func (n *Node) InnerText() string {
	var output func(*bytes.Buffer, *Node)
	output = func(buf *bytes.Buffer, n *Node) {
		if n.Type == TextNode {
			buf.WriteString(n.Data)
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(buf, child)
		}
	}
	var buf bytes.Buffer
	output(&buf, n)
	return buf.String()
}

// Add child node into parent node.
func AddChild(parent, n *Node) {
	n.Parent = parent
	n.NextSibling = nil
	switch parent.FirstChild {
	case nil:
		parent.FirstChild = n
		n.PrevSibling = nil
	default:
		parent.LastChild.NextSibling = n
		n.PrevSibling = parent.LastChild
	}
	parent.LastChild = n
}

// Add sibling node into one specific node.
func AddSibling(sibling, n *Node) {
	for t := sibling.NextSibling; t != nil; t = t.NextSibling {
		sibling = t
	}
	n.Parent = sibling.Parent
	sibling.NextSibling = n
	n.PrevSibling = sibling
	n.NextSibling = nil
	if sibling.Parent != nil {
		sibling.Parent.LastChild = n
	}
}

type copyTreeCtx struct {
	targetNode    *Node
	newTargetNode *Node
}

func (ctx *copyTreeCtx) copyTree(n *Node) *Node {
	newNode := copyNode(n)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		newChild := ctx.copyTree(child)
		newChild.PrevSibling = nil
		newChild.NextSibling = nil
		newChild.Parent = newNode
		if child == n.FirstChild {
			newNode.FirstChild = newChild
			newNode.LastChild = newChild
		} else {
			newNode.LastChild.NextSibling = newChild
			newChild.PrevSibling = newNode.LastChild
			newNode.LastChild = newChild
		}
	}
	if n == ctx.targetNode {
		ctx.newTargetNode = newNode
	}
	return newNode
}

// Make a copy of the tree the given node is in, and return the corresponding node in the new tree back.
func CopyTree(n *Node) *Node {
	ctx := &copyTreeCtx{targetNode: n}
	for ; n.Parent != nil; n = n.Parent {
	}
	ctx.copyTree(n)
	return ctx.newTargetNode
}

// Remove a node and its subtree from the tree it is in. If the node is the root of the tree, then it's no-op.
func RemoveFromTree(n *Node) {
	if n.Parent == nil {
		return
	}
	if n.Parent.FirstChild == n {
		if n.Parent.LastChild == n {
			n.Parent.FirstChild = nil
			n.Parent.LastChild = nil
		} else {
			n.Parent.FirstChild = n.NextSibling
			n.NextSibling.PrevSibling = nil
		}
	} else {
		if n.Parent.LastChild == n {
			n.Parent.LastChild = n.PrevSibling
			n.PrevSibling.NextSibling = nil
		} else {
			n.PrevSibling.NextSibling = n.NextSibling
			n.NextSibling.PrevSibling = n.PrevSibling
		}
	}
	n.Parent = nil
	n.PrevSibling = nil
	n.NextSibling = nil
}

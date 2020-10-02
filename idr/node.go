package idr

import (
	"fmt"
	"strings"
)

// NodeType is the type of a Node in an IDR.
type NodeType uint

const (
	// DocumentNode is the type of the root Node in an IDR tree.
	DocumentNode NodeType = iota
	// ElementNode is the type of an element Node in an IDR tree.
	ElementNode
	// TextNode is the type of an text/data Node in an IDR tree.
	TextNode
	// AttributeNode is the type of an attribute Node in an IDR tree.
	AttributeNode
)

// String converts NodeType to a string.
func (nt NodeType) String() string {
	switch nt {
	case DocumentNode:
		return "DocumentNode"
	case ElementNode:
		return "ElementNode"
	case TextNode:
		return "TextNode"
	case AttributeNode:
		return "AttributeNode"
	default:
		return fmt.Sprintf("(unknown NodeType: %d)", nt)
	}
}

// Node represents a node of element/data in an IDR (intermediate data representation) ingested and created
// by the omniparser.
// Credit: this is by and large a copy and some adaptation from
// https://github.com/antchfx/xmlquery/blob/master/node.go. The reasons we want to have our own struct:
// - more stability
// - one struct to represent XML/JSON/EDI/CSV/txt/etc. Vs antchfx's work have one struct (in each repo)
//   for each format.
// - Node allocation recycling.
type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type NodeType
	Data string

	FormatSpecific interface{}
}

// CreateNode creates a generic *Node.
func CreateNode(ntype NodeType, data string) *Node {
	return &Node{
		Type: ntype,
		Data: data,
	}
}

// InnerText returns a Node's children's texts concatenated.
// Note (in an XML IDR tree) none of the AttributeNode's text will be included.
func (n *Node) InnerText() string {
	var s strings.Builder
	var captureText func(*Node)
	captureText = func(a *Node) {
		switch a.Type {
		case TextNode:
			s.WriteString(a.Data)
		default:
			for child := a.FirstChild; child != nil; child = child.NextSibling {
				if child.Type != AttributeNode {
					captureText(child)
				}
			}
		}
	}
	captureText(n)
	return s.String()
}

// AddChild adds 'n' as the new last child to 'parent'.
func AddChild(parent, n *Node) {
	n.Parent = parent
	n.NextSibling = nil
	if parent.FirstChild == nil {
		parent.FirstChild = n
		n.PrevSibling = nil
	} else {
		parent.LastChild.NextSibling = n
		n.PrevSibling = parent.LastChild
	}
	parent.LastChild = n
}

// RemoveFromTree removes a node and its subtree from an IDR
// tree it is in. If the node is the root of the tree, it's a no-op.
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

package idr

import (
	"fmt"
	"strings"
)

// NodeType is the type of a Node.
type NodeType uint

const (
	// DocumentNode is the root of Node tree.
	DocumentNode NodeType = iota
	// ElementNode is an element.
	ElementNode
	// TextNode is the text content of a node.
	TextNode
	// AttributeNode is an attribute of element.
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

// Node represents a node of element/data in an IDR ingested and created by the parser.
// Credit: this is by and large a copy and some adaptation from
// https://github.com/antchfx/xmlquery/blob/master/node.go. The reasons we want to have our own struct:
// - more stability
// - one struct to represent XML/JSON/EDI/CSV/txt/etc. Vs antchfx's work have one struct for each format.
type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type NodeType
	Data string

	FormatSpecific interface{}
}

// CreateNode creates a generic *Node
func CreateNode(ntype NodeType, data string) *Node {
	return &Node{
		Type: ntype,
		Data: data,
	}
}

// InnerText returns a Node's children's texts all concatenated.
// Note in an XML Node tree, any AttributeNode's text will be ignored.
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

// RemoveFromTree removes a node and its subtree from the IDR
// tree it is in. If the node is the root of the tree, then it's no-op.
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

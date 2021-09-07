package idr

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// NodeType is the type of Node in an IDR.
type NodeType uint

const (
	// DocumentNode is the type of the root Node in an IDR tree.
	DocumentNode NodeType = iota
	// ElementNode is the type of element Node in an IDR tree.
	ElementNode
	// TextNode is the type of text/data Node in an IDR tree.
	TextNode
	// AttributeNode is the type of attribute Node in an IDR tree.
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
// - one struct to represent XML/JSON/EDI/CSV/txt/etc. Vs antchfx's work have one struct (in each repo)
//   for each format.
// - Node allocation recycling.
// - more stability
type Node struct {
	// ID uniquely identifies a Node, whether it's newly created or recycled and reused from
	// the node allocation cache. Previously we sometimes used a *Node's pointer address as a
	// unique ID which isn't sufficiently unique any more given the introduction of using
	// sync.Pool for node allocation caching.
	ID int64

	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type NodeType
	Data string

	FormatSpecific interface{}
}

// Give test a chance to turn node caching on/off. Not exported; always caching in production code.
var nodeCaching = true
var nodePool sync.Pool

func allocNode() *Node {
	n := &Node{}
	n.reset()
	return n
}

func resetNodePool() {
	nodePool = sync.Pool{
		New: func() interface{} {
			return allocNode()
		},
	}
}

func init() {
	resetNodePool()
}

// CreateNode creates a generic *Node.
func CreateNode(ntype NodeType, data string) *Node {
	if nodeCaching {
		// Node out of pool has already been reset.
		n := nodePool.Get().(*Node)
		n.Type = ntype
		n.Data = data
		return n
	}
	n := allocNode()
	n.Type = ntype
	n.Data = data
	return n
}

var nodeID = int64(0)

func newNodeID() int64 {
	return atomic.AddInt64(&nodeID, 1)
}

func (n *Node) reset() {
	n.ID = newNodeID()
	n.Parent, n.FirstChild, n.LastChild, n.PrevSibling, n.NextSibling = nil, nil, nil, nil, nil
	n.Type = 0
	n.Data = ""
	n.FormatSpecific = nil
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

// RemoveAndReleaseTree removes a node and its subtree from an IDR tree it is in and
// release the resources (Node allocation) associated with the node and its subtree.
func RemoveAndReleaseTree(n *Node) {
	if n.Parent == nil {
		goto recycle
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
recycle:
	recycle(n)
}

func recycle(n *Node) {
	if !nodeCaching {
		return
	}
	for c := n.FirstChild; c != nil; {
		// Have to save c.NextSibling before recycle(c) call or
		// c.NextSibling would be wiped out during the call.
		next := c.NextSibling
		recycle(c)
		c = next
	}
	n.reset()
	nodePool.Put(n)
}

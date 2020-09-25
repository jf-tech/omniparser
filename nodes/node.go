package nodes

import (
	"encoding/xml"

	node "github.com/antchfx/xmlquery"
)

func copyNode(n *node.Node) *node.Node {
	n2 := *n
	n2.Parent = nil
	n2.FirstChild = nil
	n2.LastChild = nil
	n2.PrevSibling = nil
	n2.NextSibling = nil
	if n.Attr != nil {
		n2.Attr = make([]xml.Attr, len(n.Attr))
		copy(n2.Attr, n.Attr)
	}
	return &n2
}

type copyTreeCtx struct {
	target    *node.Node
	newTarget *node.Node
}

func (ctx *copyTreeCtx) copyTree(n *node.Node) *node.Node {
	n2 := copyNode(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		c2 := ctx.copyTree(c)
		c2.PrevSibling = nil
		c2.NextSibling = nil
		c2.Parent = n2
		if c == n.FirstChild {
			n2.FirstChild = c2
			n2.LastChild = c2
		} else {
			n2.LastChild.NextSibling = c2
			c2.PrevSibling = n2.LastChild
			n2.LastChild = c2
		}
	}
	if n == ctx.target {
		ctx.newTarget = n2
	}
	return n2
}

// CopyTree copies the entire *Node tree the input node resides in, and returns the input node's
// corresponding node in the new tree back.
func CopyTree(n *node.Node) *node.Node {
	ctx := &copyTreeCtx{target: n}
	for ; n.Parent != nil; n = n.Parent {
	}
	ctx.copyTree(n)
	return ctx.newTarget
}

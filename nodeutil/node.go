package nodeutil

import (
	"encoding/xml"

	node "github.com/antchfx/xmlquery"
)

func copyNode(n *node.Node) *node.Node {
	// copy the content
	newNode := *n
	// reset all the pointers
	newNode.Parent = nil
	newNode.FirstChild = nil
	newNode.LastChild = nil
	newNode.PrevSibling = nil
	newNode.NextSibling = nil
	// slice field copy is by-ref when copying a struct, so must do
	// a manual copy to ensure the separation from original node.
	if n.Attr != nil {
		newNode.Attr = make([]xml.Attr, len(n.Attr))
		// xml.Attr is struct, which is copy by-value. So that's fine.
		copy(newNode.Attr, n.Attr)
	}
	return &newNode
}

type copyTreeCtx struct {
	targetNode    *node.Node
	newTargetNode *node.Node
}

func (ctx *copyTreeCtx) copyTree(n *node.Node) *node.Node {
	newNode := copyNode(n)
	for childNode := n.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		newChildNode := ctx.copyTree(childNode)
		newChildNode.PrevSibling = nil
		newChildNode.NextSibling = nil
		newChildNode.Parent = newNode
		if childNode == n.FirstChild {
			newNode.FirstChild = newChildNode
			newNode.LastChild = newChildNode
		} else {
			newNode.LastChild.NextSibling = newChildNode
			newChildNode.PrevSibling = newNode.LastChild
			newNode.LastChild = newChildNode
		}
	}
	if n == ctx.targetNode {
		ctx.newTargetNode = newNode
	}
	return newNode
}

// CopyTree copies the entire *Node tree the input node resides in, and returns the input node's
// corresponding node in the new tree back.
func CopyTree(n *node.Node) *node.Node {
	ctx := &copyTreeCtx{targetNode: n}
	for ; n.Parent != nil; n = n.Parent {
	}
	ctx.copyTree(n)
	return ctx.newTargetNode
}

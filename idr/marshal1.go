package idr

import (
	"fmt"

	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
)

// j1NodePtrName returns a categorized name for a *Node pointer used in JSONify1.
func j1NodePtrName(n *Node) *string {
	if n == nil {
		return nil
	}
	name := func(n *Node) string {
		if IsXML(n) && XMLSpecificOf(n).NamespacePrefix != "" {
			return XMLSpecificOf(n).NamespacePrefix + ":" + n.Data
		}
		return n.Data
	}
	switch n.Type {
	case DocumentNode:
		return strs.StrPtr(fmt.Sprintf("(%s)", n.Type))
	case ElementNode, AttributeNode:
		return strs.StrPtr(fmt.Sprintf("(%s %s)", n.Type, name(n)))
	case TextNode:
		return strs.StrPtr(fmt.Sprintf("(%s '%s')", n.Type, n.Data))
	default:
		return strs.StrPtr(fmt.Sprintf("(unknown '%s')", n.Data))
	}
}

// j1NodeToInterface converts *Node into an interface{} suitable for json marshaling used in JSONify1.
func j1NodeToInterface(n *Node) interface{} {
	m := make(map[string]interface{})
	m["Parent"] = j1NodePtrName(n.Parent)
	m["FirstChild"] = j1NodePtrName(n.FirstChild)
	m["LastChild"] = j1NodePtrName(n.LastChild)
	m["PrevSibling"] = j1NodePtrName(n.PrevSibling)
	m["NextSibling"] = j1NodePtrName(n.NextSibling)
	m["Type"] = n.Type.String()
	m["Data"] = n.Data
	m["FormatSpecific"] = n.FormatSpecific
	var children []interface{}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, j1NodeToInterface(child))
	}
	m["Children"] = children
	return m
}

// JSONify1 json marshals a *Node verbatim. Mostly used in test for snapshotting.
func JSONify1(n *Node) string {
	return jsons.BPM(j1NodeToInterface(n))
}

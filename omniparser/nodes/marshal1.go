package nodes

import (
	"encoding/json"
	"fmt"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/jsons"
)

// j1NodePtrName returns a categorized name for a *node.Node pointer used in JSONify1
func j1NodePtrName(n *node.Node) string {
	if n == nil {
		return "(nil)"
	}
	switch n.Type {
	case node.DocumentNode:
		return "(DocumentNode)"
	case node.DeclarationNode:
		return fmt.Sprintf("(DeclarationNode %s)", n.Data)
	case node.ElementNode:
		name := fmt.Sprintf("(ElementNode %s)", n.Data)
		if n.Prefix != "" {
			name = fmt.Sprintf("(ElementNode %s:%s)", n.Prefix, n.Data)
		}
		return name
	case node.TextNode:
		return fmt.Sprintf("(TextNode '%s')", n.Data)
	case node.CharDataNode:
		return fmt.Sprintf("(CharDataNode '%s')", n.Data)
	case node.CommentNode:
		return fmt.Sprintf("(CommentNode '%s')", n.Data)
	case node.AttributeNode:
		return fmt.Sprintf("(AttributeNode %s)", n.Data)
	default:
		return fmt.Sprintf("(unknown '%s')", n.Data)
	}
}

// j1NodeTypeStr converts uint type node.NodeType into a string for JSONify1.
func j1NodeTypeStr(nt node.NodeType) string {
	switch nt {
	case node.DocumentNode:
		return "DocumentNode"
	case node.DeclarationNode:
		return "DeclarationNode"
	case node.ElementNode:
		return "ElementNode"
	case node.TextNode:
		return "TextNode"
	case node.CharDataNode:
		return "CharDataNode"
	case node.CommentNode:
		return "CommentNode"
	case node.AttributeNode:
		return "AttributeNode"
	default:
		return fmt.Sprintf("(unknown:%d)", nt)
	}
}

// j1NodeToInterface converts *node.Node into an interface{} suitable for json marshaling used in JSONify1.
func j1NodeToInterface(n *node.Node) interface{} {
	m := make(map[string]interface{})
	m["Parent"] = j1NodePtrName(n.Parent)
	m["FirstChild"] = j1NodePtrName(n.FirstChild)
	m["LastChild"] = j1NodePtrName(n.LastChild)
	m["PrevSibling"] = j1NodePtrName(n.PrevSibling)
	m["NextSibling"] = j1NodePtrName(n.NextSibling)
	m["Type"] = j1NodeTypeStr(n.Type)
	m["Data"] = n.Data
	m["Prefix"] = n.Prefix
	m["NamespaceURI"] = n.NamespaceURI
	attrb, _ := json.Marshal(n.Attr)
	var attrv interface{}
	_ = json.Unmarshal(attrb, &attrv)
	m["Attr"] = attrv
	var children []interface{}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, j1NodeToInterface(child))
	}
	m["Children"] = children
	return m
}

// JSONify1 json marshals a *node.Node quite verbatim. Mostly used in test for snapshotting.
func JSONify1(n *node.Node) string {
	return jsons.BPM(j1NodeToInterface(n))
}

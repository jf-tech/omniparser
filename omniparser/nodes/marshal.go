package nodes

import (
	"encoding/json"
	"fmt"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/jsons"
)

func nodePtrName(n *node.Node) string {
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

func nodeTypeStr(nt node.NodeType) string {
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

func nodeToInterface(n *node.Node, omitPtr bool) interface{} {
	m := make(map[string]interface{})
	if !omitPtr {
		m["Parent"] = nodePtrName(n.Parent)
		m["FirstChild"] = nodePtrName(n.FirstChild)
		m["LastChild"] = nodePtrName(n.LastChild)
		m["PrevSibling"] = nodePtrName(n.PrevSibling)
		m["NextSibling"] = nodePtrName(n.NextSibling)
	}
	m["Type"] = nodeTypeStr(n.Type)
	m["Data"] = n.Data
	m["Prefix"] = n.Prefix
	m["NamespaceURI"] = n.NamespaceURI
	attrb, _ := json.Marshal(n.Attr)
	var attrv interface{}
	_ = json.Unmarshal(attrb, &attrv)
	m["Attr"] = attrv
	var children []interface{}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, nodeToInterface(child, omitPtr))
	}
	m["Children"] = children
	return m
}

// JSONify json marshals out a *Node.
func JSONify(n *node.Node) string {
	return JSONify2(n, false)
}

func JSONify2(n *node.Node, omitPtr bool) string {
	return jsons.BPM(nodeToInterface(n, omitPtr))
}

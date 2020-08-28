package nodeutil

import (
	"encoding/json"
	"fmt"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/jsonutil"
)

func nodeName(n *node.Node) string {
	if n == nil {
		return "(nil)"
	}
	switch n.Type {
	case node.DocumentNode:
		return "(ROOT)"
	case node.DeclarationNode:
		return fmt.Sprintf("(DECL %s)", n.Data)
	case node.ElementNode:
		name := fmt.Sprintf("(ELEM %s)", n.Data)
		if n.Prefix != "" {
			name = fmt.Sprintf("(ELEM %s:%s)", n.Prefix, n.Data)
		}
		return name
	case node.TextNode:
		return fmt.Sprintf("(TEXT '%s')", n.Data)
	case node.CharDataNode:
		return fmt.Sprintf("(CDATA '%s')", n.Data)
	case node.CommentNode:
		return fmt.Sprintf("(COMMENT '%s')", n.Data)
	case node.AttributeNode:
		return fmt.Sprintf("(ATTR '%s')", n.Data)
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

func nodeToInterface(n *node.Node) interface{} {
	m := make(map[string]interface{})
	m["Parent"] = nodeName(n.Parent)
	m["FirstChild"] = nodeName(n.FirstChild)
	m["LastChild"] = nodeName(n.LastChild)
	m["PrevSibling"] = nodeName(n.PrevSibling)
	m["NextSibling"] = nodeName(n.NextSibling)
	m["Parent"] = nodeName(n.Parent)
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
		children = append(children, nodeToInterface(child))
	}
	m["#children"] = children
	return m
}

func jsonify(n *node.Node) string {
	return jsonutil.BPM(nodeToInterface(n))
}

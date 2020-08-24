package node

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// Produces a marshaling friendly name for a given *Node.
func nodeNameForJSONMarshaler(n *Node) string {
	if n == nil {
		return "(nil)"
	}
	switch n.Type {
	case DocumentNode:
		return "(ROOT)"
	case ElementNode:
		name := fmt.Sprintf("(ELEM %s)", n.Data)
		if n.Prefix != "" {
			name = fmt.Sprintf("(ELEM %s:%s)", n.Prefix, n.Data)
		}
		return name
	case TextNode:
		return fmt.Sprintf("(TEXT '%s')", n.Data)
	case AttributeNode:
		return fmt.Sprintf("(ATTR %s)", n.Data)
	default:
		return fmt.Sprintf("(UNKNOWN type:%s data:%s)", n.Type, n.Data)
	}
}

// A custom json marshaler for Node since it contains tree link pointers, which would otherwise
// cause normal json marshaler infinite recursion and stack overflow.
func (n Node) MarshalJSON() ([]byte, error) {
	var children []*Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}
	return json.Marshal(&struct {
		Parent, FirstChild, LastChild, PrevSibling, NextSibling string
		Type                                                    NodeType
		Data                                                    string
		Prefix                                                  string
		NamespaceURI                                            string
		Attrs                                                   []xml.Attr
		Level                                                   int
		Children                                                []*Node `json:"#children"`
	}{
		Parent:       nodeNameForJSONMarshaler(n.Parent),
		FirstChild:   nodeNameForJSONMarshaler(n.FirstChild),
		LastChild:    nodeNameForJSONMarshaler(n.LastChild),
		PrevSibling:  nodeNameForJSONMarshaler(n.PrevSibling),
		NextSibling:  nodeNameForJSONMarshaler(n.NextSibling),
		Type:         n.Type,
		Data:         n.Data,
		Prefix:       n.Prefix,
		NamespaceURI: n.NamespaceURI,
		Attrs:        n.Attrs,
		Level:        n.level,
		Children:     children,
	})
}

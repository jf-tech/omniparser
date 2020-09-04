package transform

import (
	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/strs"
)

func isText(n *node.Node) bool {
	return n.Type == node.TextNode || n.Type == node.CharDataNode
}

func isChildText(n *node.Node) bool {
	// For all file formats except XML, if a node's first child is a text node, then the text node is
	// guaranteed to be the only child.
	//
	// However, for XML, there are two cases:
	// 1) an element node only contains text, e.g.:
	//   <xyz>blah</xyz>
	// In this case, the node (xyz) first child is a text node and it is the only child node of xyz.
	// 2) an element node contains sub element nodes. Due to the way XML is constructed, there might
	// be a dummy text node at the beginning in this case, e.g.:
	//   <xyz>
	//     <abc>blah</abc>
	//   </xyz>
	// The node (xyz) first child is actually a text node of "\n...." (. == space) then followed
	// by element node <abc>. In this case, we cannot say this node <xyz>'s child is of text node type.
	// (Note there is in fact another text node in the example above, that's a text node "\n" right
	// after <abc> before closing tag </xyz>)
	return n.FirstChild != nil && isText(n.FirstChild) && n.FirstChild.NextSibling == nil
}

func isChildArray(n *node.Node) bool {
	if isChildText(n) {
		return false
	}
	// Delimited, fixed-length don't have array cases.
	//
	// For json, all array children are element nodes with .Data == "".
	//
	// For xml, it's more complicated, because there is no native array notation in xml, but it can
	// be simulated:
	//   <xyz>
	//     <abc>blah1</abc>
	//     <abc>blah2</abc>
	//   </xyz>
	//
	// Note that due to the way nodes are constructed for XML, there are "dummy" text nodes (with
	// "\n" and spaces in .Data) sprinkled in between the <abc> element nodes. So to deal with it,
	// we'll go through all the child nodes, ignore all the text nodes, and if all element nodes
	// have the same .Data, then we assume it's array. For json array items are element nodes with
	// .Data == "", so that logic works too.
	//
	// Only one exception in xml:
	//   <xyz>
	//      <abc>blah</abc>
	//   </xyz>
	// Using that logic above, we'll consider this node <xyz> has array as child, arguably true but
	// counter common-sense. So we'll special case here: if there is only 1 element node, and its
	// .Data isn't "" then it's **not** considered as array. Hope this common-sense classification
	// work for most cases.
	//
	// EDI is very similar to xml case (except without those dummy text nodes)
	elemCount := 0
	elemName := (*string)(nil)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != node.ElementNode {
			continue
		}
		elemCount++
		if elemName == nil {
			elemName = strs.StrPtr(child.Data)
		} else if child.Data != *elemName {
			return false
		}
	}
	return elemCount > 1 || (elemCount == 1 && *elemName == "")
}

func nodeToObject(n *node.Node) interface{} {
	if n.FirstChild == nil {
		return nil
	}

	if isChildText(n) {
		return n.FirstChild.Data
	}

	if isChildArray(n) {
		arr := []interface{}{}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != node.ElementNode {
				continue
			}
			arr = append(arr, nodeToObject(child))
		}
		return arr
	}

	obj := map[string]interface{}{}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != node.ElementNode {
			continue
		}
		// Note: overwrite is a possible in xml input, e.g.:
		//   <xyz>
		//     <abc>blah1</abc>
		//     <abc>blah2</abc>
		//     <efg>blah3</efg>
		//   </xyz>
		// we'll end up returning map[string]interface{}{ "abc": "blah2", "efg": "blah3" }
		obj[child.Data] = nodeToObject(child)
	}
	return obj
}

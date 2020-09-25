package nodes

import (
	"encoding/json"

	node "github.com/antchfx/xmlquery"
	"github.com/jf-tech/go-corelib/strs"
)

func j2NodeName(n *node.Node) string {
	name := n.Data
	if n.Prefix != "" {
		name = n.Prefix + ":" + name
	}
	return name
}

func isText(n *node.Node) bool {
	return n.Type == node.TextNode || n.Type == node.CharDataNode
}

func isChildText(n *node.Node) bool {
	// Usually if a node's first child is a text node, then the node itself contains no other
	// children, except in XML case. Due to the way XML -> node.Node is constructed, we might
	// have unwanted text child node hanging around:
	//  <abc>
	//    <child>...</child>
	//  </abc>
	// So <abc> is clearly not a node with only text child (it has a complex sub-structure
	// <child>). But the node.Node tree constructed by XML reader looks like this:
	//
	//  ElementNode("abc")
	//    TextNode("\n")
	//    ElementNode("child")
	//    TextNode("\n")
	//
	// So in this case checking a node's first child is a text node isn't sufficient to declare
	// this node is of pure text child - we also need to check if the text child is the only
	// child or not.
	return n.FirstChild != nil && isText(n.FirstChild) && n.FirstChild.NextSibling == nil
}

func isChildArray(n *node.Node) bool {
	if isChildText(n) {
		return false
	}
	// CSV for fixed-length reader won't produce array.
	//
	// JSONStreamReader, when encountered actual array in JSON input, produces
	// a collection of child element nodes with their Data field equal empty string.
	//
	// XML/EDI are a bit more complicated:
	// There is not concept of array in XML, neither in EDI. XML, however, can simulate
	// array by having a number of child elements with the same element name. EDI has
	// segment loop. So the primary idea is to check child element nodes' names and use
	// their similarity to determine whether the children represent an array or not.
	// Throwing a wrench here is that fact in XML, there are dummy text nodes sprinkled
	// around, in between XML elements, so we need to ignore them.
	//
	// One ambiguous case in XML (sigh...):
	//   <abc>
	//      <child>something</child>
	//   </abc>
	// We can consider node 'abc' contains a field 'child', or we can consider node 'abc'
	// contains an array of one element which is 'child'. There is really no other signs/flags
	// help us to disambiguate the situation. So use common sense here, treat this situation
	// like a single child field, instead of a child array with one element.
	numElemNodes := 0
	nameElem := (*string)(nil)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != node.ElementNode {
			continue
		}
		numElemNodes++
		if nameElem == nil {
			nameElem = strs.StrPtr(j2NodeName(c))
		} else if j2NodeName(c) != *nameElem {
			// We have differently named child element nodes, clearly
			// this isn't an array case.
			return false
		}
	}
	// If the number of identically named child element nodes is > 1 (including JSON's "" named child
	// element nodes) then it's an array case;
	// If there is only one child element node, then it is array case only when its name is "" - that's
	// in JSON array case. Otherwise, just treat it as a single child field case.
	return numElemNodes > 1 || (numElemNodes == 1 && *nameElem == "")
}

// J2NodeToInterface converts *node.NodeType into an interface{} that, once marshaled, is JSON friendly.
func J2NodeToInterface(n *node.Node) interface{} {
	if n.FirstChild == nil {
		return nil
	}

	if isChildText(n) {
		return n.FirstChild.Data
	}

	if isChildArray(n) {
		arr := []interface{}{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != node.ElementNode {
				continue
			}
			arr = append(arr, J2NodeToInterface(c))
		}
		return arr
	}

	obj := map[string]interface{}{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != node.ElementNode {
			continue
		}
		// Unfortunately this node back to JSON transform can't be 100% fidelity.
		// Again, XML is the special case:
		//   <abc>
		//     <efg>1</abc>
		//     <efg>2</abc>
		//     <xyz>3</efg>
		//   </abc>
		// We'll overwrite 'efg' here so the return value will be:
		//   map[string]interface{}{ "efg": "2", "xyz": "3" }
		// It is what it is.
		obj[j2NodeName(c)] = J2NodeToInterface(c)
	}
	return obj
}

// JSONify2 JSON marshals a *node.Node into a minified JSON string that's user friendly.
func JSONify2(n *node.Node) string {
	b, _ := json.Marshal(J2NodeToInterface(n))
	return string(b)
}

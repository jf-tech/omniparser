package idr

import (
	"encoding/json"
	"strconv"

	"github.com/jf-tech/go-corelib/strs"
)

type ctx struct {
	useJSONType bool
}

func (ctx *ctx) j2NodeName(n *Node) string {
	if IsXML(n) && XMLSpecificOf(n).NamespacePrefix != "" {
		return XMLSpecificOf(n).NamespacePrefix + ":" + n.Data
	}
	return n.Data
}

func (ctx *ctx) isChildText(n *Node) bool {
	// For non XML node, situations are quite simple: if there is a text child node, then it will
	// be the one and only one child.
	// But for XML, some complexities:
	// 1) dummy text nodes hanging around:
	//  <abc>
	//    <efg/>
	//  </abc>
	// If isChildText is called upon node <abc>, it should ideally return false due to the sub
	// element node <efg> presence. However due to the way XML nodes are constructed, there will
	// be "dummy" text nodes hanging around:
	//  ElementNode ("abc")
	//    TextNode ("\n")
	//    ElementNode ("efg")
	//    TextNode ("\n")
	// So in this case we should check for the presence of non TextNode to determine isChildText
	// returns true or false, except in yet another XML only situation:
	// 2) attribute nodes hanging around:
	//  <abc attr1=".." attr2="..">
	//    only text
	//  </abc>
	// Now this time, the node tree will be built like this:
	//  ElementNode ("abc")
	//    AttributeNode ("attr1")
	//    AttributeNode ("attr2")
	//    TextNode ("only text")
	// So in this case, we need to check for the present of non TextNode, except for AttributeNode.
	// All in all, the logic for isChildText returning true on an XML node should be:
	// - it contains at least one text child
	// - it doesn't contain any element child.
	textNodeFound := false
	elemNodeFound := false
	for n = n.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == TextNode {
			textNodeFound = true
			// We've found text child node, but we need to continue probing.
		} else if n.Type == ElementNode {
			elemNodeFound = true
			// We've found an unwanted element child node, don't probing.
			break
		}
		// attribute child node is ignored.
	}
	return textNodeFound && !elemNodeFound
}

func (ctx *ctx) isChildArray(n *Node) bool {
	if ctx.useJSONType && IsJSONArr(n) {
		return true
	}
	// CSV or fixed-length formats won't produce array.
	//
	// JSONStreamReader, when encountered actual array in JSON input, produces
	// a collection of child element nodes with their Data fields all equal to
	// empty string.
	//
	// XML/EDI are a bit more complicated:
	// There is not concept of array in XML, neither in EDI. XML, however, can simulate
	// array by having a number of child elements with the same element name. EDI has
	// segment loop. So the primary idea is to check child element nodes' names and use
	// their similarity to determine whether the children represent an array or not.
	// Throwing a wrench here is that fact in XML, there are dummy text nodes sprinkled
	// around, in between XML elements, so we need to ignore them. Also there could be
	// attribute nodes in XML, need to ignore them too.
	//
	// One ambiguous case in XML (and similarly in EDI):
	//   <abc>
	//      <child>something</child>
	//   </abc>
	// We can consider node 'abc' contains a field 'child', or we can consider node 'abc'
	// contains an array of one element which is 'child'. There is really no other signs/flags
	// help us to disambiguate the situation. So use common sense here, treat this situation
	// like a single child field, instead of a child array with one element. EDI is similar.
	// EDI doesn't have some indicator (i.e. a segment can have multiple instances), but it's
	// too hard to pass the indicator here. So hope this promise/limitation is acceptable.
	elemNum := 0
	elemName := (*string)(nil)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != ElementNode {
			continue
		}
		elemNum++
		if elemName == nil {
			elemName = strs.StrPtr(ctx.j2NodeName(c))
		} else if ctx.j2NodeName(c) != *elemName {
			// We have differently named child element nodes, clearly
			// this isn't an array case.
			return false
		}
	}
	// If the number of identically named child element nodes is > 1 (including JSON's "" named child
	// element nodes) then it's an array case;
	// If there is only one child element node, then it is array case only when its name is "" - that's
	// in JSON array case. Otherwise, just treat it as a single child field case.
	return elemNum > 1 || (elemNum == 1 && *elemName == "")
}

func (ctx *ctx) getChildData(n *Node) interface{} {
	if !IsJSON(n) || !ctx.useJSONType {
		return n.InnerText()
	}
	n = n.FirstChild
	switch {
	case IsJSONValueNum(n):
		f, _ := strconv.ParseFloat(n.Data, 64)
		return f
	case IsJSONValueBool(n):
		b, _ := strconv.ParseBool(n.Data)
		return b
	case IsJSONValueNull(n):
		return nil
	default:
		return n.Data
	}
}

// J2NodeToInterface converts *Node into an interface{} that, once marshaled, is JSON friendly.
func (ctx *ctx) nodeToInterface(n *Node) interface{} {
	switch {
	case ctx.isChildText(n):
		return ctx.getChildData(n)
	case ctx.isChildArray(n):
		arr := make([]interface{}, 0)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// unfortunately there doesn't seem to be a way to include AttributeNode elegantly
			// in an XML array situation:
			//  <abc attr1="1" attr2="2">
			//    <e>one</e>
			//    <e>two</e>
			//  </abc>
			// We want to return, for the node of <abc>, this:
			//  [ "one", "two" ]
			// It's hard to jam in the attributes collection somewhere without causing confusion.
			// Don't think it's a prevailing or even relevant case. So be it.
			if c.Type != ElementNode {
				continue
			}
			arr = append(arr, ctx.nodeToInterface(c))
		}
		return arr
	default:
		obj := make(map[string]interface{})
		attrs := make(map[string]interface{})
		fieldIsArr := make(map[string]bool)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			name := ctx.j2NodeName(c)
			value := ctx.nodeToInterface(c)
			if c.Type == ElementNode {
				// Both in XML and EDI, we can potentially have something like the following:
				//   <abc>
				//     <efg>1</efg>
				//     <efg>2</efg>
				//     <xyz>3</xyz>
				//   </abc>
				// We don't want to overwrite first <efg> with second <efg>. So instead we
				// will return something like this:
				//  map[string]interface{} {
				//    "efg": [ "1", "2" ],
				//    "xyz": "3"
				//  }
				if _, found := obj[name]; found {
					if fieldIsArr[name] {
						obj[name] = append(obj[name].([]interface{}), value)
					} else {
						obj[name] = []interface{}{obj[name], value}
						fieldIsArr[name] = true
					}
				} else {
					obj[name] = value
				}
			} else if c.Type == AttributeNode {
				attrs[name] = ctx.nodeToInterface(c)
			}
		}
		if len(attrs) > 0 {
			// Given AttributeNode is only possible/present in XML case, and in
			// XML, legal element names cannot contain '#', so we use '#' prefix
			// here to indicate this is a special field.
			obj["#attributes"] = attrs
		}
		return obj
	}
}

func J2NodeToInterface(n *Node, useJSONType bool) interface{} {
	return (&ctx{useJSONType: useJSONType}).nodeToInterface(n)
}

// JSONify2 JSON marshals a *Node into a minified JSON string.
func JSONify2(n *Node) string {
	b, _ := json.Marshal(J2NodeToInterface(n, true))
	return string(b)
}

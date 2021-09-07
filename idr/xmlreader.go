package idr

import (
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"golang.org/x/net/html/charset"
)

// XMLStreamReader is a streaming XML to *Node reader.
type XMLStreamReader struct {
	d                          *xml.Decoder
	space2prefix               map[string]string
	xpathExpr, xpathFilterExpr *xpath.Expr
	root, cur, stream          *Node
	err                        error
}

// streamCandidateCheck checks if sp.cur is a potential stream candidate.
// See more details/explanation in JSONStreamReader.streamCandidateCheck.
func (sp *XMLStreamReader) streamCandidateCheck() {
	if sp.xpathExpr != nil && sp.stream == nil && MatchAny(sp.root, sp.xpathExpr) {
		sp.stream = sp.cur
	}
}

// wrapUpCurAndTargetCheck wraps sp.cur node processing and also checks if the sp.cur is the stream
// candidate and if it is, then does a final check: a stream candidate is the target if:
// - If it has finished processing (sp.cur == sp.stream)
// - Either we don't have a stream filter xpath or the stream filter xpath matches.
func (sp *XMLStreamReader) wrapUpCurAndTargetCheck() *Node {
	cur := sp.cur
	// No matter what outcome the wrapUpCurAndTargetCheck() is, the current node is done, and
	// we need to adjust sp.cur to its parent.
	sp.cur = sp.cur.Parent
	// Only do stream target check if the finished cur node is the stream candidate
	if cur != sp.stream {
		return nil
	}
	if sp.xpathFilterExpr == nil || MatchAny(sp.root, sp.xpathFilterExpr) {
		return sp.stream
	}
	// This means while the sp.stream was marked as stream candidate by the initial
	// sp.xpathExpr matching, now we've completed the construction of this node fully and
	// discovered sp.xpathFilterExpr can't be satisfied, so this sp.stream isn't a
	// stream target. To prevent future mismatch for other stream candidate, we need to
	// remove it from Node tree completely. And reset sp.stream.
	RemoveAndReleaseTree(sp.stream)
	sp.stream = nil
	return nil
}

func (sp *XMLStreamReader) updateNamespaces(attrs []xml.Attr) {
	// https://www.w3.org/TR/xml-names/#scoping-defaulting
	for _, attr := range attrs {
		if attr.Name.Local == "xmlns" {
			sp.space2prefix[attr.Value] = ""
		} else if attr.Name.Space == "xmlns" {
			sp.space2prefix[attr.Value] = attr.Name.Local
		}
	}
}

// addNonTextChild creates an XML node of a given type and put it as a child of sp.cur.
// If the call succeeds, sp.cur will be advanced to the newly created child node.
func (sp *XMLStreamReader) addNonTextChild(ntype NodeType, tok interface{}) error {
	var xmlSpecific XMLSpecific
	name := tok.(xml.Name)
	data := name.Local
	// If default namespace is declared at root such as `<root xmlns="uri://blah">`, then
	// for nodes such as "<abc>", its xml.Name returned by xml.Decoder would look like this:
	//   xml.Name.Space = "uri://blah"
	//   xml.Name.Local = "abc"
	// We need to do a look-up to find out what the corresponding namespace prefix is.
	// If no namespace whatsoever is declared, then xml.Name.Space returned by xml.Decoder
	// is empty. No namespace prefix lookup is needed.
	if name.Space != "" {
		namespaceURI := name.Space
		namespacePrefix, found := sp.space2prefix[namespaceURI]
		if !found {
			if ntype == AttributeNode && name.Space == "xmlns" {
				// When xml.Decoder returns xml.Name for attributes, there are two cases:
				// 1. normal attributes: it follows the same namespace prefix/URI rules that
				// the returned xml.Name.Space contains the full URI of an attribute namespace,
				// if any; and xml.Name.Local contains the actual attribute name.
				// 2. namespace declaration attributes: these attributes are usually at the root
				// level and look like this:
				//  <lb0:library xmlns:lb0="uri://xyz">
				//    ...
				//  </lb0:library>
				// Here we have a namespace declaration attribute, specifying for the rest of the XML
				// doc we expect to see namespace prefix 'lb0' mapped to URI 'uri://xyz'. When this
				// attribute 'xmlns:lb0' is returned by xml.Decoder, the xml.Name would be:
				//   xml.Name.Space = "xmlns"
				//   xml.Name.Local = "lb0"
				// This is the exception case where you won't be able to find a prefix for namespace
				// "xmlns". As such, we'll keep "xmlns" as prefix and URI as "".
				namespaceURI = ""
				namespacePrefix = name.Space
			} else {
				// For others (element nodes), a namespace URI to prefix lookup failure means we have
				// an undeclared namespace prefix. It's a fatal error.
				return fmt.Errorf("unknown namespace '%s' on %s '%s'", namespaceURI, ntype, data)
			}
		}
		xmlSpecific.NamespaceURI = namespaceURI
		xmlSpecific.NamespacePrefix = namespacePrefix
	}
	child := CreateXMLNode(ntype, data, xmlSpecific)
	AddChild(sp.cur, child)
	sp.cur = child
	return nil
}

// addTextChild creates an XML node of TextNode type and put it as a child of sp.cur.
// Note given we never adds anything below a TextNode, addTextChild does NOT advance
// sp.cur to the newly created child.
func (sp *XMLStreamReader) addTextChild(text string) {
	child := CreateXMLNode(TextNode, text, XMLSpecific{})
	AddChild(sp.cur, child)
}

func (sp *XMLStreamReader) parse() (*Node, error) {
	for {
		tok, err := sp.d.Token()
		if err != nil {
			// including io.EOF
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			sp.updateNamespaces(tok.Attr)
			err = sp.addNonTextChild(ElementNode, tok.Name)
			if err != nil {
				return nil, err
			}
			for _, attr := range tok.Attr {
				err = sp.addNonTextChild(AttributeNode, attr.Name)
				if err != nil {
					return nil, err
				}
				sp.addTextChild(attr.Value)
				// Remember sp.addNonTextChild auto advances sp.cur to the newly added child node
				// and sp.addTextChild doesn't. In this case, we're done with attr node and its
				// text node creation and there will be nothing more to be added below it, so back off.
				sp.cur = sp.cur.Parent
			}
			sp.streamCandidateCheck()
		case xml.EndElement:
			ret := sp.wrapUpCurAndTargetCheck()
			if ret != nil {
				return ret, nil
			}
		case xml.CharData:
			sp.addTextChild(string(tok))
		}
	}
}

// Read returns a *Node that matches the xpath streaming criteria.
func (sp *XMLStreamReader) Read() (n *Node, err error) {
	if sp.err != nil {
		return nil, sp.err
	}
	// Because this is a streaming read, we need to remove the last
	// stream node from the node tree to free up memory. If Release()
	// is called after Read() call, then sp.stream is already cleaned up;
	// adding this piece of code here just in case Release() isn't called.
	if sp.stream != nil {
		RemoveAndReleaseTree(sp.stream)
		sp.stream = nil
	}
	n, sp.err = sp.parse()
	return n, sp.err
}

// Release releases the *Node (and its subtree) that Read() has previously
// returned. Note even if Release is not explicitly called, next Read() call
// will still release the current streaming target node.
func (sp *XMLStreamReader) Release(n *Node) {
	if n == sp.stream {
		sp.stream = nil
	}
	RemoveAndReleaseTree(n)
}

// AtLine returns the **rough** line number of the current XML decoder.
func (sp *XMLStreamReader) AtLine() int {
	// Given all the libraries are of fixed versions in go modules, we're fine.
	// If in the future, something changes and breaks due to library upgrade,
	// we'll have test failures to remind us to fix.
	return int(reflect.ValueOf(sp.d).Elem().FieldByName("line").Int())
}

// NewXMLStreamReader creates a new instance of XML streaming reader.
func NewXMLStreamReader(r io.Reader, xpathStr string) (*XMLStreamReader, error) {
	xpathStr = strings.TrimSpace(xpathStr)
	xpathNoFilterStr := removeLastFilterInXPath(xpathStr)
	xpathExpr, err := caches.GetXPathExpr(xpathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid xpath '%s', err: %s", xpathStr, err.Error())
	}
	// If the original xpath is valid, then this xpath with last filter removed gotta
	// be valid as well. So no error checking.
	xpathNoFilterExpr, _ := caches.GetXPathExpr(xpathNoFilterStr)
	reader := &XMLStreamReader{
		d: xml.NewDecoder(r),
		// http://www.w3.org/XML/1998/namespace is bound by definition to the prefix xml.
		space2prefix: map[string]string{
			"http://www.w3.org/XML/1998/namespace": "xml",
		},
		xpathExpr: xpathNoFilterExpr,
		xpathFilterExpr: func() *xpath.Expr {
			if xpathStr == xpathNoFilterStr {
				return nil
			}
			return xpathExpr
		}(),
		root: CreateXMLNode(DocumentNode, "", XMLSpecific{}),
	}
	reader.d.CharsetReader = charset.NewReaderLabel
	reader.cur = reader.root
	return reader, nil
}

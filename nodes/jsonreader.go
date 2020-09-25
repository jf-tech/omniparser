package nodes

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	node "github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"
)

type jnodeType uint

const (
	jnodeTypeRoot jnodeType = 1 << iota
	jnodeTypeObj
	jnodeTypeArr
	jnodeTypeProp
	jnodeTypeVal
)

// Ultimately the parser constructs and returns a *node.Node tree, however, we need to decorate
// each node with `jnodeType` to indicate what kind of json element it is. So have to resort to
// this companion `jnode` struct during parsing.
type jnode struct {
	jnodeType jnodeType
	parent    *jnode
	n         *node.Node
}

// addChild creates a new child `jnode` and its underlying *node.Node, adds the underlying
// *node.Node to the node tree, links the child `jnode` to this parent `jnode`, and returns
// the new child `jnode`.
func (jn *jnode) addChild(jnodeType jnodeType, nodeType node.NodeType, data interface{}) *jnode {
	child := &jnode{
		jnodeType: jnodeType,
		n:         &node.Node{Type: nodeType, Data: data2str(data)},
	}
	node.AddChild(jn.n, child.n)
	child.parent = jn
	return child
}

func (jn *jnode) isRoot() bool { return jn.jnodeType&jnodeTypeRoot != 0 }
func (jn *jnode) isObj() bool  { return jn.jnodeType&jnodeTypeObj != 0 }
func (jn *jnode) isArr() bool  { return jn.jnodeType&jnodeTypeArr != 0 }
func (jn *jnode) isProp() bool { return jn.jnodeType&jnodeTypeProp != 0 }
func (jn *jnode) isVal() bool  { return jn.jnodeType&jnodeTypeVal != 0 }

// In *node.Node tree, all data is string, so we need to convert json data token (string, float64, bool, nil)
// to data string. Note we miss "nil" conversion here, because we don't really have a way to represent
// nil in *node.Node tree, thus parse won't even call this with nil token value.
func data2str(data interface{}) string {
	s := ""
	switch v := data.(type) {
	case string:
		s = v
	case float64:
		s = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		s = strconv.FormatBool(v)
	}
	return s
}

// JSONStreamParser is a streaming json to *node.Node parser.
type JSONStreamParser struct {
	r                          *ios.LineCountingReader
	d                          *json.Decoder
	xpathExpr, xpathFilterExpr *xpath.Expr
	root, curNode, streamNode  *jnode
}

// streamCandidateCheck checks if sp.curNode is a potential stream candidate.
func (sp *JSONStreamParser) streamCandidateCheck() {
	if sp.xpathExpr != nil && sp.streamNode == nil && node.QuerySelector(sp.root.n, sp.xpathExpr) != nil {
		sp.streamNode = sp.curNode
	}
}

// streamTargetCheck checks if the stream candidate is actually a stream target and ready to be
// returned to the caller. A stream candidate is the target if:
// - If it has finished processing (sp.curNode == sp.streamNode)
// - Either we don't have a stream filter xpath or the stream filter xpath matches.
func (sp *JSONStreamParser) streamTargetCheck() *node.Node {
	ret := (*node.Node)(nil)
	if sp.curNode == sp.streamNode {
		if sp.xpathFilterExpr == nil || node.QuerySelector(sp.root.n, sp.xpathFilterExpr) != nil {
			ret = sp.streamNode.n
		} else {
			// This means while the sp.streamNode was marked as stream candidate by the initial
			// sp.xpathExpr matching, now we've completed the construction of this node fully and
			// discovered sp.xpathFilterExpr can't be satisfied, so this sp.streamNode isn't a
			// stream target. To prevent future mismatch for other stream candidate, we need to
			// remove it from node.Node tree completely. And reset sp.streamNode.
			node.RemoveFromTree(sp.streamNode.n)
			sp.streamNode = nil
		}
	}
	return ret
}

func (sp *JSONStreamParser) parseDelim(tok json.Delim) *node.Node {
	switch tok {
	case '{':
		switch {
		// Note case order matters, because curNode type could be prop|arr or root|arr, in those
		// cases, we want `case isArr` to be hit first.
		case sp.curNode.isArr():
			// if we see "{" inside an "[]", we create an anonymous object element node for it.
			sp.curNode = sp.curNode.addChild(jnodeTypeObj, node.ElementNode, "")
			sp.streamCandidateCheck()
		case sp.curNode.isProp():
			// a "{" follows a property name, indicate this property is an object.
			// We don't need to streamCandidateCheck here because we've already done
			// the check when the property itself is processed.
			sp.curNode.jnodeType |= jnodeTypeObj
		case sp.curNode.isRoot():
			// if we see "{" directly on root, this root contains a single object. Mark the root
			// node's type so. Also we need to check stream candidate, in case caller wants to
			// match the entire json doc.
			sp.curNode.jnodeType |= jnodeTypeObj
			sp.streamCandidateCheck()
		}
	case '[':
		switch {
		// Note case order matters
		case sp.curNode.isArr():
			// every immediate thing created inside an array is rooted by an anonymous element node.
			sp.curNode = sp.curNode.addChild(jnodeTypeArr, node.ElementNode, "")
			sp.streamCandidateCheck()
		case sp.curNode.isProp():
			// Again, similarly we don't do streamCandidateCheck here since the check is already
			// done when the property node is created.
			sp.curNode.jnodeType |= jnodeTypeArr
		case sp.curNode.isRoot():
			// arr directly on root.
			sp.curNode.jnodeType |= jnodeTypeArr
			sp.streamCandidateCheck()
		}
	case '}', ']':
		ret := sp.streamTargetCheck()
		sp.curNode = sp.curNode.parent
		if ret != nil {
			return ret
		}
	}
	return nil
}

func (sp *JSONStreamParser) parseVal(tok json.Token) *node.Node {
	switch {
	// Note case order matters, because curNode type could be prop|obj or root|obj, in those
	// cases, we want isObj case to be hit first.
	case sp.curNode.isObj():
		sp.curNode = sp.curNode.addChild(jnodeTypeProp, node.ElementNode, tok)
		sp.streamCandidateCheck()
	// similarly we want isArr hit before isProp/isRoot
	case sp.curNode.isArr():
		// if parent is an array, so we're adding a value directly to the array.
		// by creating an anonymous element node, then the value as text node
		// underneath it.
		sp.curNode = sp.curNode.addChild(jnodeTypeProp, node.ElementNode, "")
		sp.streamCandidateCheck()
		// Limitation: we don't ever add "nil" as a text node there is no matching representation
		// in *node.Node tree.
		if tok != nil {
			sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
		}
		ret := sp.streamTargetCheck()
		sp.curNode = sp.curNode.parent
		if ret != nil {
			return ret
		}
	case sp.curNode.isProp():
		if tok != nil {
			sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
		}
		ret := sp.streamTargetCheck()
		sp.curNode = sp.curNode.parent
		if ret != nil {
			return ret
		}
	case sp.curNode.isRoot():
		// A value is directly setting on root. We need to do both stream candidate check
		// as well as target check.
		sp.streamCandidateCheck()
		if tok != nil {
			sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
		}
		ret := sp.streamTargetCheck()
		sp.curNode = sp.curNode.parent
		if ret != nil {
			return ret
		}
	}
	return nil
}

func (sp *JSONStreamParser) parse() (*node.Node, error) {
	for {
		tok, err := sp.d.Token()
		if err != nil {
			// err includes io.EOF
			return nil, err
		}
		switch tok := tok.(type) {
		case json.Delim:
			if ret := sp.parseDelim(tok); ret != nil {
				return ret, nil
			}
		case string, float64, bool, nil:
			if ret := sp.parseVal(tok); ret != nil {
				return ret, nil
			}
		}
	}
}

// Read returns a matching *node.Node
func (sp *JSONStreamParser) Read() (*node.Node, error) {
	// Because this is a streaming read, we need to release/remove last
	// stream node from the node tree to free up memory.
	if sp.streamNode != nil {
		node.RemoveFromTree(sp.streamNode.n)
		sp.streamNode = nil
	}
	return sp.parse()
}

// AtLine returns the **rough** line number of the current json decoder.
func (sp *JSONStreamParser) AtLine() int {
	return sp.r.AtLine()
}

// NewJSONStreamParser creates a new instance of json streaming parser.
func NewJSONStreamParser(r io.Reader, xpathStr string) (*JSONStreamParser, error) {
	xpathStr = strings.TrimSpace(xpathStr)
	xpathNoFilterStr := RemoveLastFilterInXPath(xpathStr)
	xpathExpr, err := caches.GetXPathExpr(xpathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid xpath '%s', err: %s", xpathStr, err.Error())
	}
	xpathNoFilterExpr, _ := caches.GetXPathExpr(xpathNoFilterStr)
	ifelse := func(cond bool, expr1, expr2 *xpath.Expr) *xpath.Expr {
		if cond {
			return expr1
		}
		return expr2
	}
	lineCountingReader := ios.NewLineCountingReader(r)
	parser := &JSONStreamParser{
		r:               lineCountingReader,
		d:               json.NewDecoder(lineCountingReader),
		xpathExpr:       xpathNoFilterExpr,
		xpathFilterExpr: ifelse(xpathStr == xpathNoFilterStr, nil, xpathExpr),
		root:            &jnode{jnodeType: jnodeTypeRoot, n: &node.Node{Type: node.DocumentNode}},
	}
	parser.curNode = parser.root
	return parser, nil
}

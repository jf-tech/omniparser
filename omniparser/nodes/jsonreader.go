package nodes

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	node "github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"

	"github.com/jf-tech/omniparser/cache"
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
	d                          *json.Decoder
	xpathExpr, xpathFilterExpr *xpath.Expr
	root, curNode, streamNode  *jnode
}

// Creates a child `jnode` (with underlying *node.Node type == element) to the curNode,
// checks if this newly created node matches the stream xpath or not, and marks it as
// stream candidate.
func (sp *JSONStreamParser) addChild(jnodeType jnodeType, data interface{}) {
	sp.curNode = sp.curNode.addChild(jnodeType, node.ElementNode, data)
	if sp.xpathExpr != nil && sp.streamNode == nil && node.QuerySelector(sp.root.n, sp.xpathExpr) != nil {
		sp.streamNode = sp.curNode
	}
}

// streamTargetReady checks if the stream candidate is ready to be returned to the caller.
// A stream candidate is ready if:
// - If it has finished processing (sp.curNode == sp.streamNode)
// - Either we don't have a stream filter xpath or the stream filter xpath matches.
func (sp *JSONStreamParser) streamTargetReady() *node.Node {
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

func (sp *JSONStreamParser) parse() (*node.Node, error) {
	// This is a special case where the stream target is directly on the root.
	// e.g. json doc with only "abc".
	if sp.curNode != nil && sp.curNode.isRoot() &&
		sp.xpathExpr != nil && sp.streamNode == nil && node.QuerySelector(sp.root.n, sp.xpathExpr) != nil {
		sp.streamNode = sp.curNode
	}
	for {
		tok, err := sp.d.Token()
		if err != nil {
			// err includes io.EOF
			return nil, err
		}
		switch tok := tok.(type) {
		case json.Delim:
			switch tok {
			case '{':
				switch {
				// Note case order matters, because curNode type could be prop|arr or root|arr, in those
				// cases, we want `case isArr` to be hit first.
				case sp.curNode.isArr():
					// if we see "{" inside an "[]", we create an anonymous object element node for it.
					sp.addChild(jnodeTypeObj, "")
				case sp.curNode.isRoot(), sp.curNode.isProp():
					// if we see "{" directly on root, or it's following a property name, then we just
					// need to change the current node (root or a prop node) attribute to be of an obj.
					sp.curNode.jnodeType |= jnodeTypeObj
				}
			case '[':
				switch {
				// Note case order matters
				case sp.curNode.isArr():
					sp.addChild(jnodeTypeArr, "") // think it as a nameless anonymous arr
				case sp.curNode.isRoot(), sp.curNode.isProp():
					sp.curNode.jnodeType |= jnodeTypeArr
				}
			case '}', ']':
				ret := sp.streamTargetReady()
				sp.curNode = sp.curNode.parent
				if ret != nil {
					return ret, nil
				}
			}
		case string, float64, bool, nil:
			switch {
			// Note case order matters, because curNode type could be prop|obj or root|obj, in those
			// cases, we want isObj case to be hit first.
			case sp.curNode.isObj():
				sp.addChild(jnodeTypeProp, tok)
			// similarly we want isArr hit before isProp/isRoot
			case sp.curNode.isArr():
				// if parent is an array, so we're adding a value directly to the array.
				// by creating an anonymous element node, then the value as text node
				// underneath it.
				sp.addChild(jnodeTypeProp, "") // think it as a nameless anonymous property
				// Limitation: we don't ever add "nil" as a text node there is no matching representation
				// in *node.Node tree.
				if tok != nil {
					sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
				}
				ret := sp.streamTargetReady()
				sp.curNode = sp.curNode.parent
				if ret != nil {
					return ret, nil
				}
			case sp.curNode.isProp(), sp.curNode.isRoot():
				if tok != nil {
					sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
				}
				ret := sp.streamTargetReady()
				sp.curNode = sp.curNode.parent
				if ret != nil {
					return ret, nil
				}
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

// NewJSONStreamParser creates a new instance of json streaming parser.
func NewJSONStreamParser(r io.Reader, xpathStr string) (*JSONStreamParser, error) {
	xpathStr = strings.TrimSpace(xpathStr)
	xpathNoFilterStr := RemoveLastFilterInXPath(xpathStr)
	xpathExpr, err := cache.GetXPathExpr(xpathStr)
	if err != nil {
		return nil, err
	}
	xpathNoFilterExpr, _ := cache.GetXPathExpr(xpathNoFilterStr)
	ifelse := func(cond bool, expr1, expr2 *xpath.Expr) *xpath.Expr {
		if cond {
			return expr1
		}
		return expr2
	}
	parser := &JSONStreamParser{
		d:               json.NewDecoder(r),
		xpathExpr:       xpathNoFilterExpr,
		xpathFilterExpr: ifelse(xpathStr == xpathNoFilterStr, nil, xpathExpr),
		root:            &jnode{jnodeType: jnodeTypeRoot, n: &node.Node{Type: node.DocumentNode}},
	}
	parser.curNode = parser.root
	return parser, nil
}

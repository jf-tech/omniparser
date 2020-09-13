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

type jnode struct {
	jnodeType jnodeType
	parent    *jnode
	n         *node.Node
}

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

func (sp *JSONStreamParser) addElemChildNodeToCurNode(jnodeType jnodeType, data interface{}) {
	sp.curNode = sp.curNode.addChild(jnodeType, node.ElementNode, data)
	if sp.xpathExpr != nil && sp.streamNode == nil && node.QuerySelector(sp.root.n, sp.xpathExpr) != nil {
		sp.streamNode = sp.curNode
	}
}

func (sp *JSONStreamParser) parse() (*node.Node, error) {
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
					sp.addElemChildNodeToCurNode(jnodeTypeObj, "") // think it as a nameless anonymous obj
				case sp.curNode.isRoot(), sp.curNode.isProp():
					sp.curNode.jnodeType |= jnodeTypeObj
				}
			case '[':
				switch {
				// Note case order matters
				case sp.curNode.isArr():
					sp.addElemChildNodeToCurNode(jnodeTypeArr, "") // think it as a nameless anonymous arr
				case sp.curNode.isRoot(), sp.curNode.isProp():
					sp.curNode.jnodeType |= jnodeTypeArr
				}
			case '}', ']':
				ret := (*node.Node)(nil)
				if sp.curNode == sp.streamNode {
					if sp.xpathFilterExpr == nil || node.QuerySelector(sp.root.n, sp.xpathFilterExpr) != nil {
						ret = sp.streamNode.n
					} else {
						node.RemoveFromTree(sp.streamNode.n)
						sp.streamNode = nil
					}
				}
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
				sp.addElemChildNodeToCurNode(jnodeTypeProp, tok)
			// similarly we want isArr hit before isProp/isRoot
			case sp.curNode.isArr():
				// if parent is an array, so we're adding a value directly to the array.
				// by creating an anonymous element node, then the value as text node
				// underneath it.
				sp.addElemChildNodeToCurNode(jnodeTypeProp, "") // think it as a nameless anonymous property
				// Limitation: we don't ever add "nil" as a text node there is no matching representation
				// in *node.Node tree.
				if tok != nil {
					sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
				}
				ret := (*node.Node)(nil)
				if sp.curNode == sp.streamNode {
					if sp.xpathFilterExpr == nil || node.QuerySelector(sp.root.n, sp.xpathFilterExpr) != nil {
						ret = sp.streamNode.n
					} else {
						node.RemoveFromTree(sp.streamNode.n)
						sp.streamNode = nil
					}
				}
				sp.curNode = sp.curNode.parent
				if ret != nil {
					return ret, nil
				}
			case sp.curNode.isProp(), sp.curNode.isRoot():
				if tok != nil {
					sp.curNode.addChild(jnodeTypeVal, node.TextNode, tok)
				}
				ret := (*node.Node)(nil)
				if sp.curNode == sp.streamNode {
					if sp.xpathFilterExpr == nil || node.QuerySelector(sp.root.n, sp.xpathFilterExpr) != nil {
						ret = sp.streamNode.n
					} else {
						node.RemoveFromTree(sp.streamNode.n)
						sp.streamNode = nil
					}
				}
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
	xpathFullExpr, err := cache.GetXPathExpr(xpathStr)
	if err != nil {
		return nil, err
	}
	xpathFull := strings.TrimSpace(xpathStr)
	xpathNoLastFilter := RemoveLastFilterInXPath(xpathFull)
	xpathNoLastFilterExpr := (*xpath.Expr)(nil)
	if xpathNoLastFilter != xpathFull {
		xpathNoLastFilterExpr, _ = cache.GetXPathExpr(xpathNoLastFilter)
	}
	parser := &JSONStreamParser{
		d: json.NewDecoder(r),
		xpathExpr: func() *xpath.Expr {
			if xpathNoLastFilterExpr != nil {
				return xpathNoLastFilterExpr
			}
			return xpathFullExpr
		}(),
		xpathFilterExpr: func() *xpath.Expr {
			if xpathNoLastFilterExpr != nil {
				return xpathFullExpr
			}
			return nil
		}(),
		root: &jnode{jnodeType: jnodeTypeRoot, n: &node.Node{Type: node.DocumentNode}},
	}
	parser.curNode = parser.root
	return parser, nil
}

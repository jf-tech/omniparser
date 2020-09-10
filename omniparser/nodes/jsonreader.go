package nodes

import (
	"encoding/json"
	"io"
	"strings"

	node "github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
)

type JSONStreamParser struct {
	d                          *json.Decoder
	xpathExpr, xpathFilterExpr *xpath.Expr
	root, streamNode           *jnode
}

func NewJSONStreamParser(r io.Reader, xpathStr string) (*JSONStreamParser, error) {
	xpathExpr, err := xpath.Compile(xpathStr)
	if err != nil {
		return nil, err
	}
	xpathFilterExpr := (*xpath.Expr)(nil)
	xpathStr = strings.TrimSpace(xpathStr)
	xpathFilterStr := RemoveLastFilterInXPath(xpathStr)
	if xpathStr != xpathFilterStr {
		xpathFilterExpr, err = xpath.Compile(xpathFilterStr)
		if err != nil {
			return nil, err
		}
	}
	parser := &JSONStreamParser{
		d:               json.NewDecoder(r),
		xpathExpr:       xpathExpr,
		xpathFilterExpr: xpathFilterExpr,
		root:            &jnode{n: &node.Node{Type: node.DocumentNode}},
	}
	return parser, nil
}

type jnode struct {
	n, parent, prev *node.Node
	level           int
}

func (sp *JSONStreamParser) Read() (*node.Node, error) {
	if sp.streamNode != nil {
		node.RemoveFromTree(sp.streamNode.n)
		sp.prev = sp.streamNodePrev
		sp.streamNode = nil
		sp.streamNodePrev = nil
	}
	return sp.parse()
}

func (sp *JSONStreamParser) parse() (*node.Node, error) {
	for {
		tok, err := sp.d.Token()
		if err != nil {
			// err includes io.EOF
			return nil, err
		}
		switch tok.(type) {
		case json.Delim:
		case string:
		case float64:
		}
	}
}

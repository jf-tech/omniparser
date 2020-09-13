package nodes

import (
	"errors"
	"fmt"

	node "github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"

	"github.com/jf-tech/omniparser/cache"
)

var (
	// ErrNoMatch is returned when not a single matched node can be found.
	ErrNoMatch = errors.New("no match")
	// ErrMoreThanExpected is returned when more than expected matched nodes are found.
	ErrMoreThanExpected = errors.New("more than expected matched")
)

const (
	// DisableXPathCache disables caching xpath compilation when MatchAll/MatchSingle
	// are called. Useful when caller knows the xpath string isn't cache-able (such as
	// containing unique IDs, timestamps, etc) which would otherwise cause the xpath
	// compilation cache grow unbounded.
	DisableXPathCache = uint(1) << iota
)

func loadXPathExpr(expr string, flags []uint) (*xpath.Expr, error) {
	var flagsActual uint
	switch len(flags) {
	case 0:
		flagsActual = 0
	case 1:
		flagsActual = flags[0]
	default:
		return nil, fmt.Errorf("only one flag is allowed, instead got: %v", flags)
	}
	var exp interface{}
	var err error
	if flagsActual&DisableXPathCache != 0 {
		exp, err = xpath.Compile(expr)
	} else {
		exp, err = cache.GetXPathExpr(expr)
	}
	if err != nil {
		return nil, fmt.Errorf("xpath '%s' compilation failed: %s", expr, err.Error())
	}
	return exp.(*xpath.Expr), nil
}

// MatchAll uses the given xpath expression 'expr' to find all the matching nodes
// contained in the tree rooted at 'top'.
func MatchAll(top *node.Node, expr string, flags ...uint) ([]*node.Node, error) {
	// We have quite a few places a simple "." xpath query can be issued, a simple
	// optimization to reduce workload in that situation.
	if expr == "." {
		return []*node.Node{top}, nil
	}
	exp, err := loadXPathExpr(expr, flags)
	if err != nil {
		return nil, err
	}
	return node.QuerySelectorAll(top, exp), nil
}

// MatchSingle uses the given xpath expression 'expr' to find one and exactly one matching node
// contained in the tree rooted at 'top'. If no matching node is found, ErrNoMatch is returned;
// if more than one matching nodes are found, ErrMoreThanExpected is returned.
func MatchSingle(top *node.Node, expr string, flags ...uint) (*node.Node, error) {
	nodes, err := MatchAll(top, expr, flags...)
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 0:
		return nil, ErrNoMatch
	case 1:
		return nodes[0], nil
	default:
		return nil, ErrMoreThanExpected
	}
}

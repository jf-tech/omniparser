package idr

import (
	"errors"
	"fmt"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
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

func loadXPathExpr(exprStr string, flags []uint) (*xpath.Expr, error) {
	var flagsActual uint
	switch len(flags) {
	case 0:
		flagsActual = 0
	case 1:
		flagsActual = flags[0]
	default:
		return nil, fmt.Errorf("only one flag is allowed, instead got: %d", len(flags))
	}
	var expr interface{}
	var err error
	if flagsActual&DisableXPathCache != 0 {
		expr, err = xpath.Compile(exprStr)
	} else {
		expr, err = caches.GetXPathExpr(exprStr)
	}
	if err != nil {
		return nil, fmt.Errorf("xpath '%s' compilation failed: %s", exprStr, err.Error())
	}
	return expr.(*xpath.Expr), nil
}

// QueryIter initiates an xpath query specified by 'expr' against an IDR tree rooted at 'n'.
func QueryIter(n *Node, expr *xpath.Expr) *xpath.NodeIterator {
	return expr.Select(createNavigator(n))
}

// MatchAny returns true if the xpath query 'expr' against an IDR tree rooted at 'n' yields any result.
func MatchAny(n *Node, expr *xpath.Expr) bool {
	return QueryIter(n, expr).MoveNext()
}

// MatchAll returns all the matched nodes by an xpath query 'exprStr' against an IDR tree rooted at 'n'.
func MatchAll(n *Node, exprStr string, flags ...uint) ([]*Node, error) {
	if exprStr == "." {
		return []*Node{n}, nil
	}
	exp, err := loadXPathExpr(exprStr, flags)
	if err != nil {
		return nil, err
	}
	iter := QueryIter(n, exp)
	var ret []*Node
	for iter.MoveNext() {
		ret = append(ret, nodeFromIter(iter))
	}
	return ret, nil
}

// MatchSingle returns one and only one matched node by an xpath query 'exprStr' against an IDR tree rooted
// at 'n'. If no matching node is found, ErrNoMatch is returned; if more than one matching nodes are found,
// ErrMoreThanExpected is returned.
func MatchSingle(n *Node, exprStr string, flags ...uint) (*Node, error) {
	if exprStr == "." {
		return n, nil
	}
	expr, err := loadXPathExpr(exprStr, flags)
	if err != nil {
		return nil, err
	}
	iter := QueryIter(n, expr)
	if !iter.MoveNext() {
		return nil, ErrNoMatch
	}
	ret := nodeFromIter(iter)
	if iter.MoveNext() {
		return nil, ErrMoreThanExpected
	}
	return ret, nil
}

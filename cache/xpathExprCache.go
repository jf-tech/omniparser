package cache

import (
	"github.com/antchfx/xpath"
)

// XPathExprCache is the default loading cache used for caching the compiled
// xpath expression. If the default size is too big/small and/or a cache limit isn't
// desired at all, caller can simply replace the cache during global initialization.
// But be aware it's global so any packages uses this package inside your process will
// be affected.
var XPathExprCache = NewLoadingCache()

// GetXPathExpr compiles a given xpath expression string and returns a compile xpath.Expr
// or error.
func GetXPathExpr(expr string) (*xpath.Expr, error) {
	exp, err := XPathExprCache.Get(expr, func(key interface{}) (interface{}, error) {
		return xpath.Compile(key.(string))
	})
	if err != nil {
		return nil, err
	}
	return exp.(*xpath.Expr), nil
}

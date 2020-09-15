package customfuncs

import (
	"errors"
	"fmt"

	"github.com/Knetic/govaluate"

	"github.com/jf-tech/omniparser/cache"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

var evalExprCache = cache.NewLoadingCache()

// Deprecated. Kept for back-compatibility. Use 'javascript' instead.
// For supported operators, check: https://github.com/Knetic/govaluate/blob/master/MANUAL.md
func eval(_ *transformctx.Ctx, exprStr string, args ...string) (string, error) {
	if len(args)%2 != 0 {
		return "", errors.New("invalid number of args to 'eval'")
	}
	params := make(map[string]interface{}, len(args)/2)
	for i := 0; i < len(args)/2; i++ {
		n, v, err := parseArgTypeAndValue(args[i*2], args[i*2+1])
		if err != nil {
			return "", err
		}
		params[n] = v
	}
	expr, err := evalExprCache.Get(exprStr, func(key interface{}) (interface{}, error) {
		return govaluate.NewEvaluableExpression(key.(string))
	})
	if err != nil {
		return "", err
	}
	result, err := expr.(*govaluate.EvaluableExpression).Evaluate(params)
	if err != nil {
		return "", err
	}
	switch {
	case result == nil:
		return "", nil
	default:
		return fmt.Sprint(result), nil
	}
}

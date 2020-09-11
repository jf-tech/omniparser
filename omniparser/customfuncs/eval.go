package customfuncs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"

	"github.com/jf-tech/omniparser/cache"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/strs"
)

var evalExprCache = cache.NewLoadingCache()

const (
	evalArgTypeString  = "string"
	evalArgTypeInt     = "int"
	evalArgTypeFloat   = "float"
	evalArgTypeBoolean = "boolean"
)

func constructEvalParam(argDecl, argValue string) (name string, value interface{}, err error) {
	declParts := strings.Split(argDecl, ":")
	if len(declParts) != 2 {
		return "", nil, errors.New("arg decl must be in format of '<arg_name>:<arg_type>'")
	}
	name = declParts[0]
	if !strs.IsStrNonBlank(name) {
		return "", nil, errors.New("arg_name in '<arg_name>:<arg_type>' cannot be a blank string")
	}
	switch declParts[1] {
	case evalArgTypeString:
		return name, argValue, nil
	case evalArgTypeInt:
		f, err := strconv.ParseFloat(argValue, 64)
		if err != nil {
			return "", nil, err
		}
		return name, int64(f), nil
	case evalArgTypeFloat:
		f, err := strconv.ParseFloat(argValue, 64)
		if err != nil {
			return "", nil, err
		}
		return name, f, nil
	case evalArgTypeBoolean:
		b, err := strconv.ParseBool(argValue)
		if err != nil {
			return "", nil, err
		}
		return name, b, nil
	default:
		return "", nil, fmt.Errorf("arg_type '%s' in '<arg_name>:<arg_type>' is not supported", declParts[1])
	}
}

// For supported operators, check: https://github.com/Knetic/govaluate/blob/master/MANUAL.md
func eval(_ *transformctx.Ctx, exprStr string, args ...string) (string, error) {
	if len(args)%2 != 0 {
		return "", errors.New("invalid number of args to 'eval'")
	}
	params := make(map[string]interface{}, len(args)/2)
	for i := 0; i < len(args)/2; i++ {
		n, v, err := constructEvalParam(args[i*2], args[i*2+1])
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

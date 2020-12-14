package transform

import (
	"reflect"

	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

// CustomParseFuncType represents a custom_parse function type. Deprecated. Use customfuncs.CustomFuncType.
type CustomParseFuncType func(*transformctx.Ctx, *idr.Node) (interface{}, error)

// CustomParseFuncs is a map from custom_parse names to an actual custom parse functions. Deprecated. Use
// customfuncs.CustomFuncs.
type CustomParseFuncs = map[string]CustomParseFuncType

func (p *parseCtx) invokeCustomParse(customParseFn CustomParseFuncType, n *idr.Node) (interface{}, error) {
	result := reflect.ValueOf(customParseFn).Call(
		[]reflect.Value{
			reflect.ValueOf(p.transformCtx),
			reflect.ValueOf(n),
		})
	// result[0] - result
	// result[1] - error
	if result[1].Interface() == nil {
		return result[0].Interface(), nil
	}
	return nil, result[1].Interface().(error)
}

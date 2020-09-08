package transform

import (
	"reflect"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

// CustomParseFuncType represents a custom_parse function type.
type CustomParseFuncType func(*transformctx.Ctx, *node.Node) (interface{}, error)

// CustomParseFuncs is a map from custom_parse names to an actual custom parse functions.
type CustomParseFuncs = map[string]CustomParseFuncType

func (p *parseCtx) invokeCustomParse(customParseFn CustomParseFuncType, n *node.Node) (interface{}, error) {
	result := reflect.ValueOf(customParseFn).Call(
		[]reflect.Value{
			reflect.ValueOf(p.opCtx),
			reflect.ValueOf(n),
		})
	// result[0] - result
	// result[1] - error
	if result[1].Interface() == nil {
		return result[0].Interface(), nil
	}
	return nil, result[1].Interface().(error)
}

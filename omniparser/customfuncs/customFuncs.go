package customfuncs

import (
	"bytes"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

// CustomFuncType is the type of a custom function. Has to use interface{} given we support
// non-variadic and variadic functions.
type CustomFuncType = interface{}

// CustomFuncs is a map from custom func names to an actual custom func.
type CustomFuncs = map[string]CustomFuncType

// BuiltinCustomFuncs contains all the built-in custom functions.
var BuiltinCustomFuncs = map[string]CustomFuncType{
	// keep these custom funcs lexically sorted
	"concat": concat,
}

func concat(_ *transformctx.Ctx, strs ...string) (string, error) {
	var b bytes.Buffer
	for _, s := range strs {
		b.WriteString(s)
	}
	return b.String(), nil
}

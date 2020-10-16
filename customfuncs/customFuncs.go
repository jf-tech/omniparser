package customfuncs

import (
	"strings"

	"github.com/google/uuid"

	"github.com/jf-tech/omniparser/transformctx"
)

// CustomFuncType is the type of a custom function. Has to use interface{} given we support
// non-variadic and variadic functions.
type CustomFuncType = interface{}

// CustomFuncs is a map from custom func names to an actual custom func.
type CustomFuncs = map[string]CustomFuncType

// Merge merges multiple custom func maps into one.
func Merge(funcs ...CustomFuncs) CustomFuncs {
	merged := make(CustomFuncs)
	for _, fs := range funcs {
		for name, f := range fs {
			merged[name] = f
		}
	}
	return merged
}

var CommonCustomFuncs = map[string]CustomFuncType{
	// keep these custom funcs lexically sorted
	"concat":                  concat,
	"dateTimeLayoutToRFC3339": dateTimeLayoutToRFC3339,
	"dateTimeToRFC3339":       dateTimeToRFC3339,
	"lower":                   lower,
	"upper":                   upper,
	"uuidv3":                  uuidv3,
}

func concat(_ *transformctx.Ctx, strs ...string) (string, error) {
	var w strings.Builder
	for _, s := range strs {
		w.WriteString(s)
	}
	return w.String(), nil
}

func lower(_ *transformctx.Ctx, s string) (string, error) {
	return strings.ToLower(s), nil
}

func upper(_ *transformctx.Ctx, s string) (string, error) {
	return strings.ToUpper(s), nil
}

func uuidv3(_ *transformctx.Ctx, s string) (string, error) {
	return uuid.NewMD5(uuid.Nil, []byte(s)).String(), nil
}

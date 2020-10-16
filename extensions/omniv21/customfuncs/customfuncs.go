package customfuncs

import (
	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

// OmniV21CustomFuncs contains 'omni.2.1' specific custom funcs.
var OmniV21CustomFuncs = map[string]customfuncs.CustomFuncType{
	// keep these custom funcs lexically sorted
	"copy":                    CopyFunc,
	"javascript":              JavaScript,
	"javascript_with_context": JavaScriptWithContext,
}

// CopyFunc copies the current contextual idr.Node and returns it as a JSON marshaling friendly interface{}.
func CopyFunc(_ *transformctx.Ctx, n *idr.Node) (interface{}, error) {
	return idr.J2NodeToInterface(n, true), nil
}

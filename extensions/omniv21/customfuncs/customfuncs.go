package customfuncs

import (
	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

// omniv21 specific custom funcs.
var OmniV21CustomFuncs = map[string]customfuncs.CustomFuncType{
	// keep these custom funcs lexically sorted
	"copy":                    copyFunc,
	"javascript":              javascript,
	"javascript_with_context": javascriptWithContext,
}

func copyFunc(_ *transformctx.Ctx, n *idr.Node) (interface{}, error) {
	return idr.J2NodeToInterface(n, true), nil
}

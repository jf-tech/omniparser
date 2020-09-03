package transformctx

import (
	"github.com/jf-tech/omniparser/omniparser/errs"
)

// ExtensionCtx is a context object supplied by an extension. An extension
// of omniparser can supply its own custom funcs and/or its own schema plugin.
// This ctx object allows caller to "communicates" with its supplied extension
// custom funcs and/or schema plugin during input parsing/transform.
type ExtensionCtx = interface{}

// Ctx contains the context object used throughout the lifespan of a TransformOp action.
type Ctx struct {
	// InputName is the name of the input stream to be processed.
	InputName string
	// CtxAwareErr allows context aware error formatting such as adding input (file) name
	// and line number as a prefix to the error string.
	CtxAwareErr errs.CtxAwareErr
	// ExtCtx is extension specific context object that allows communications between
	// caller and extension's custom functions and/or schema plugin during input
	// parsing/transform.
	ExtCtx ExtensionCtx
}

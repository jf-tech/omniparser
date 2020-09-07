package transformctx

import (
	"github.com/jf-tech/omniparser/omniparser/errs"
)

// Ctx contains the context object used throughout the lifespan of a TransformOp action.
type Ctx struct {
	// InputName is the name of the input stream to be processed.
	InputName string
	// ExternalProperties contains exteranlly set string properties used in schema.
	ExternalProperties map[string]string
	// CtxAwareErr allows context aware error formatting such as adding input (file) name
	// and line number as a prefix to the error string.
	CtxAwareErr errs.CtxAwareErr
}

func (ctx *Ctx) ExternalProperty(name string) (string, bool) {
	if len(ctx.ExternalProperties) == 0 {
		return "", false
	}
	if v, found := ctx.ExternalProperties[name]; found {
		return v, true
	}
	return "", false
}

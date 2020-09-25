package transformctx

import (
	"github.com/jf-tech/omniparser/errs"
)

// Ctx is the context object used throughout a Transform operation.
type Ctx struct {
	// InputName is the name of the input stream to be ingested and transformed.
	InputName string
	// ExternalProperties contains externally set string properties used by schema in the transform.
	ExternalProperties map[string]string
	// CtxAwareErr allows context aware error formatting such as adding input (file) name
	// and line number as a prefix to the error string.
	CtxAwareErr errs.CtxAwareErr
}

// External looks up, and returns an external property value, if exists.
func (ctx *Ctx) External(name string) (string, bool) {
	v, found := ctx.ExternalProperties[name]
	return v, found
}

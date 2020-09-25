package transformctx

import (
	"github.com/jf-tech/omniparser/errs"
)

// Ctx contains the context object used throughout a Transform operation.
type Ctx struct {
	// InputName is the name of the input stream to be ingested and transformed.
	InputName string
	// ExternalProperties contains externally set string properties used by schema in the transform.
	ExternalProperties map[string]string
	// CtxAwareErr allows context aware error formatting such as adding input (file) name
	// and line number as a prefix to the error string.
	CtxAwareErr errs.CtxAwareErr
}

// ExternalProperty looks up, and returns an external property value, if exists.
func (ctx *Ctx) ExternalProperty(name string) (string, bool) {
	if len(ctx.ExternalProperties) == 0 {
		return "", false
	}
	if v, found := ctx.ExternalProperties[name]; found {
		return v, true
	}
	return "", false
}

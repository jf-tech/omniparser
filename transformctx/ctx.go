package transformctx

import (
	"github.com/jf-tech/omniparser/errs"
)

// Ctx is the context object used throughout a Transform operation.
type Ctx struct {
	// InputName is the name of the input stream to be ingested and transformed.
	// Most of the time there is no need for caller of NewTransform to set it, it will be auto-set
	// by omniparser.
	InputName string
	// ExternalProperties contains externally set string properties used by `external` transform
	// in a schema.
	ExternalProperties map[string]string
	// CtxAwareErr allows context aware error formatting such as adding input (file) name
	// and line number as a prefix to the error string. Most of the time there is no need for caller
	// of NewTransform to set it, it will be auto-set by omniparser.
	CtxAwareErr errs.CtxAwareErr
	// CustomParam lets caller of NewTransform set a custom parameter they see fit, and this custom
	// param will be passed along with the Ctx object throughout all the stages and operations of
	// a transform, including passing to all the `custom_func` and `custom_parse`.
	CustomParam interface{}
}

// External looks up, and returns an external property value, if exists.
func (ctx *Ctx) External(name string) (string, bool) {
	v, found := ctx.ExternalProperties[name]
	return v, found
}

package omniparser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jf-tech/go-corelib/ios"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/handlers"
	omniv2 "github.com/jf-tech/omniparser/handlers/omni/v2"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/transformctx"
	"github.com/jf-tech/omniparser/validation"
)

// Schema is an interface that represents an schema used by omniparser.
// One instance of Schema is associated with one and only one schema.
// The instance of Schema can be reused for ingesting and transforming
// multiple input files/streams, as long as they are all intended for the
// same schema.
// Each ingestion/transform, however, needs a separate instance of
// Transform. A Transform must not be shared and reused across different
// input files/streams.
// While the same instance of Schema can be shared across multiple threads,
// Transform is not multi-thread safe. All operations on it must be done
// within the same go routine.
type Schema interface {
	NewTransform(name string, input io.Reader, ctx *transformctx.Ctx) (Transform, error)
	Header() header.Header
	Content() []byte
}

type schema struct {
	name    string
	header  header.Header
	content []byte
	handler handlers.SchemaHandler
}

// Extension allows user of omniparser be able to add new schema handlers, and/or new custom functions
// in addition to builtin handlers and functions.
type Extension struct {
	CreateHandler handlers.CreateHandlerFunc
	HandlerParams interface{}
	CustomFuncs   customfuncs.CustomFuncs
}

var (
	defaultExtOmniV2Handler = Extension{
		CreateHandler: omniv2.CreateHandler,
	}
	defaultExtBuiltinCustomFuncs = Extension{
		CustomFuncs: customfuncs.BuiltinCustomFuncs,
	}
	defaultExts = []Extension{
		defaultExtOmniV2Handler,
		defaultExtBuiltinCustomFuncs,
	}
)

// NewSchema creates a new instance of Schema. Caller can use the optional Extensions for customization.
func NewSchema(name string, schemaReader io.Reader, exts ...Extension) (Schema, error) {
	content, err := ioutil.ReadAll(schemaReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read schema '%s': %s", name, err.Error())
	}
	// validate the universal parser_settings header schema.
	err = validation.SchemaValidate(name, content, validation.JSONSchemaParserSettings)
	if err != nil {
		// The err from validation.SchemaValidate is already context formatted.
		return nil, err
	}
	var h header.Header
	// parser_settings has just been json schema validated. so unmarshaling will not go wrong.
	_ = json.Unmarshal(content, &h)

	allExts := append([]Extension(nil), exts...)
	allExts = append(allExts, defaultExts...)
	allCustomFuncs := collectCustomFuncs(allExts)
	for _, ext := range allExts {
		if ext.CreateHandler == nil {
			continue
		}
		handler, err := ext.CreateHandler(&handlers.HandlerCtx{
			Name:          name,
			Header:        h,
			Content:       content,
			CustomFuncs:   allCustomFuncs,
			HandlerParams: ext.HandlerParams,
		})
		if err == errs.ErrSchemaNotSupported {
			continue
		}
		if err != nil {
			// The err from handler's CreateHandler is already ctxAwareErr formatted, so directly return.
			return nil, err
		}
		return &schema{
			name:    name,
			header:  h,
			content: content,
			handler: handler,
		}, nil
	}
	return nil, errs.ErrSchemaNotSupported
}

func collectCustomFuncs(exts []Extension) customfuncs.CustomFuncs {
	cfs := customfuncs.CustomFuncs(nil)
	for _, ext := range exts {
		cfs = customfuncs.Merge(cfs, ext.CustomFuncs)
	}
	return cfs
}

// NewTransform creates and returns an instance of Transform for a given input stream.
func (s *schema) NewTransform(name string, input io.Reader, ctx *transformctx.Ctx) (Transform, error) {
	br, err := ios.StripBOM(s.header.ParserSettings.WrapEncoding(input))
	if err != nil {
		return nil, err
	}
	if ctx.InputName != name {
		ctx.InputName = name
	}
	ingester, err := s.handler.NewIngester(ctx, br)
	if err != nil {
		return nil, err
	}
	// If caller already specified a way to do context aware error formatting, use it;
	// otherwise (vast majority cases), use the Ingester (which implements CtxAwareErr
	// interface) created by the schema handler.
	if ctx.CtxAwareErr == nil {
		ctx.CtxAwareErr = ingester
	}
	return &transform{ingester: ingester}, nil
}

// Header returns the schema header.
func (s *schema) Header() header.Header {
	return s.header
}

// Content returns the schema content.
func (s *schema) Content() []byte {
	return s.content
}

package omniparser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jf-tech/go-corelib/ios"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21"
	v21 "github.com/jf-tech/omniparser/extensions/omniv21/customfuncs"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/schemahandler"
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
	handler schemahandler.SchemaHandler
}

// Extension allows user of omniparser to add new schema handlers, and/or new custom functions
// in addition to the builtin handlers and functions.
type Extension struct {
	CreateSchemaHandler       schemahandler.CreateFunc
	CreateSchemaHandlerParams interface{}
	CustomFuncs               customfuncs.CustomFuncs
}

var (
	defaultExt = Extension{
		// 'omni.2.1' extension
		CreateSchemaHandler: omniv21.CreateSchemaHandler,
		CustomFuncs:         customfuncs.Merge(customfuncs.CommonCustomFuncs, v21.OmniV21CustomFuncs),
	}
)

// NewSchema creates a new instance of Schema. Caller can use the optional Extensions for customization.
// NewSchema will scan through exts left to right to find the first extension with a schema handler (specified
// by CreateSchemaHandler field) that supports the input schema. If no ext provided or no ext with a handler
// that supports the schema, then NewSchema will fall back to builtin extension (currently for schema version
// 'omni.2.1'. If the input schema is still not supported by builtin extension, NewSchema will fail with
// ErrSchemaNotSupported. Each extension much be fully self-contained meaning all the custom functions it
// intends to use in the schemas supported by it must be included in the same extension.
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
	allExts = append(allExts, defaultExt)
	for _, ext := range allExts {
		if ext.CreateSchemaHandler == nil {
			continue
		}
		handler, err := ext.CreateSchemaHandler(&schemahandler.CreateCtx{
			Name:         name,
			Header:       h,
			Content:      content,
			CustomFuncs:  ext.CustomFuncs,
			CreateParams: ext.CreateSchemaHandlerParams,
		})
		if err == errs.ErrSchemaNotSupported {
			continue
		}
		if err != nil {
			// The err from handler's CreateSchemaHandler is already ctxAwareErr formatted, so directly return.
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

package omniparser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

// Parser is an interface that represents an instance of omniparser.
// One instance of Parser is associated with one and only one schema.
// The instance of Parser can be reused for parsing and transforming
// multiple input files/streams, as long as they are all intended for the
// same schema.
// Each parsing/transform, however, needs a separate instance of
// TransformOp. TransformOp must not be shared and reused across different
// input files/streams.
// While the same instance of Parser can be shared across multiple threads,
// TransformOp is not multi-thread safe. All operations on it must be done
// within the same go routine.
type Parser interface {
	GetTransformOp(name string, input io.Reader, ctx *transformctx.Ctx) (TransformOp, error)
	SchemaHeader() schemaplugin.Header
	SchemaContent() []byte
}

// Extension allows client of omniparser to supply its own custom funcs and/or schema plugin.
type Extension struct {
	CustomFuncs customfuncs.CustomFuncs
	ParseSchema schemaplugin.SchemaParserFunc
}

var BuiltinExtensions = []Extension{
	{
		CustomFuncs: customfuncs.BuiltinCustomFuncs,
	},
}

type parser struct {
	schemaName    string
	schemaHeader  schemaplugin.Header
	schemaContent []byte
	exts          []Extension
	schemaPlugin  schemaplugin.Plugin
}

// NewParser creates a new instance of omniparser for a given schema. Caller can also supply
// additional extensions (on top of builtin extensions) so caller can adds new custom funcs and/or
// new schema plugins.
func NewParser(schemaName string, schemaReader io.Reader, exts ...Extension) (Parser, error) {
	schemaContent, err := ioutil.ReadAll(schemaReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read schema '%s': %s", schemaName, err.Error())
	}
	var schemaHeader schemaplugin.Header
	err = json.Unmarshal(schemaContent, &schemaHeader)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read schema '%s': corrupted header `parser_settings`: %s", schemaName, err)
	}
	var allExts []Extension
	allExts = append(allExts, exts...)
	allExts = append(allExts, BuiltinExtensions...)
	for _, ext := range allExts {
		if ext.ParseSchema == nil {
			continue
		}
		plugin, err := ext.ParseSchema(schemaName, schemaHeader, schemaContent)
		if err == errs.ErrSchemaNotSupported {
			continue
		}
		if err != nil {
			// The err from plugin's ParseSchema is already ctxAwareErr formatted, so directly return.
			return nil, err
		}
		return &parser{
			schemaName:    schemaName,
			schemaHeader:  schemaHeader,
			schemaContent: schemaContent,
			exts:          allExts,
			schemaPlugin:  plugin,
		}, nil
	}
	return nil, errs.ErrSchemaNotSupported
}

// GetTransformOp creates and returns an instance of TransformOp for a given input.
func (p *parser) GetTransformOp(name string, input io.Reader, ctx *transformctx.Ctx) (TransformOp, error) {
	panic("TBD")
}

// SchemaHeader returns the associated schema plugin's schema header.
func (p *parser) SchemaHeader() schemaplugin.Header {
	return p.schemaHeader
}

// SchemaContent returns the associated schema plugin's schema content.
func (p *parser) SchemaContent() []byte {
	return p.schemaContent
}

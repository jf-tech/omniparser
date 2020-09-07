package omniparser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jf-tech/iohelper"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	omniv2 "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2"
	"github.com/jf-tech/omniparser/omniparser/schemavalidate"
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

type parser struct {
	schemaName    string
	schemaHeader  schemaplugin.Header
	schemaContent []byte
	schemaPlugin  schemaplugin.Plugin
}

type SchemaPluginConfig struct {
	CustomFuncs  customfuncs.CustomFuncs
	ParseSchema  schemaplugin.ParseSchemaFunc
	PluginParams interface{}
}

// NewParser creates a new instance of omniparser for a given schema. Caller can use the optional
// pluginConfigs to supply additional custom funcs, and/or additional schema plugins and their
// corresponding initialization params.
func NewParser(schemaName string, schemaReader io.Reader, pluginConfigs ...SchemaPluginConfig) (Parser, error) {
	schemaContent, err := ioutil.ReadAll(schemaReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read schema '%s': %s", schemaName, err.Error())
	}
	// validate the universal parser_settings header schema.
	err = schemavalidate.SchemaValidate(schemaName, schemaContent, schemavalidate.JSONSchemaParserSettings)
	if err != nil {
		// The err from schemavalidate.SchemaValidate is already context formatted.
		return nil, err
	}
	var schemaHeader schemaplugin.Header
	// parser_settings has just been json schema validated. so unmarshaling will not go wrong.
	_ = json.Unmarshal(schemaContent, &schemaHeader)

	var allPluginConfigs []SchemaPluginConfig
	allPluginConfigs = append(allPluginConfigs, pluginConfigs...)
	allPluginConfigs = append(allPluginConfigs, SchemaPluginConfig{
		ParseSchema: omniv2.ParseSchema,
	})
	for _, plugin := range allPluginConfigs {
		if plugin.ParseSchema == nil {
			continue
		}
		plugin, err := plugin.ParseSchema(&schemaplugin.ParseSchemaCtx{
			Name:    schemaName,
			Header:  schemaHeader,
			Content: schemaContent,
			// keep builtin custom funcs at tail, so that if we have name collision between the caller
			// supplied custom funcs and built-in ones, the built-in ones win.
			CustomFuncs:  customfuncs.Merge(plugin.CustomFuncs, customfuncs.BuiltinCustomFuncs),
			PluginParams: plugin.PluginParams,
		})
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
			schemaPlugin:  plugin,
		}, nil
	}
	return nil, errs.ErrSchemaNotSupported
}

// GetTransformOp creates and returns an instance of TransformOp for a given input.
func (p *parser) GetTransformOp(name string, input io.Reader, ctx *transformctx.Ctx) (TransformOp, error) {
	br, err := iohelper.StripBOM(p.schemaHeader.ParserSettings.WrapEncoding(input))
	if err != nil {
		return nil, err
	}
	inputProcessor, err := p.schemaPlugin.GetInputProcessor(ctx, br)
	if err != nil {
		return nil, err
	}
	if ctx.InputName != name {
		ctx.InputName = name
	}
	// If caller already specified a way to do context aware error formatting, use it;
	// otherwise (vast majority cases), use the InputProcessor (which implements CtxAwareErr
	// interface) created by the schema plugin.
	if ctx.CtxAwareErr == nil {
		ctx.CtxAwareErr = inputProcessor
	}
	return &transformOp{inputProcessor: inputProcessor}, nil
}

// SchemaHeader returns the associated schema plugin's schema header.
func (p *parser) SchemaHeader() schemaplugin.Header {
	return p.schemaHeader
}

// SchemaContent returns the associated schema plugin's schema content.
func (p *parser) SchemaContent() []byte {
	return p.schemaContent
}

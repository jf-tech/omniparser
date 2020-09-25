package omniv2

import (
	"fmt"
	"io"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/schemaplugin"
	omniv2fileformat "github.com/jf-tech/omniparser/schemaplugin/omni/v2/fileformat"
	omniv2json "github.com/jf-tech/omniparser/schemaplugin/omni/v2/fileformat/json"
	omniv2xml "github.com/jf-tech/omniparser/schemaplugin/omni/v2/fileformat/xml"
	"github.com/jf-tech/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/schemavalidate"
	"github.com/jf-tech/omniparser/transformctx"
)

const (
	// PluginVersion is the version of omniv2 schema plugin.
	PluginVersion = "omni.2.0"
)

// PluginParams allows user of omniparser to provide omniv2 schema plugin customization.
type PluginParams struct {
	CustomFileFormat omniv2fileformat.FileFormat
	CustomParseFuncs transform.CustomParseFuncs
}

// ParseSchema parses, validates and creates an omni-schema based schema plugin.
func ParseSchema(ctx *schemaplugin.ParseSchemaCtx) (schemaplugin.Plugin, error) {
	if ctx.Header.ParserSettings.Version != PluginVersion {
		return nil, errs.ErrSchemaNotSupported
	}
	// First do a `transform_declarations` json schema validation
	err := schemavalidate.SchemaValidate(ctx.Name, ctx.Content, schemavalidate.JSONSchemaTransformDeclarations)
	if err != nil {
		// err is already context formatted.
		return nil, err
	}
	finalOutputDecl, err := transform.ValidateTransformDeclarations(ctx.Content, ctx.CustomFuncs, customParseFuncs(ctx))
	if err != nil {
		return nil, fmt.Errorf(
			"schema '%s' 'transform_declarations' validation failed: %s",
			ctx.Name, err.Error())
	}
	for _, fileFormat := range fileFormats(ctx) {
		formatRuntime, err := fileFormat.ValidateSchema(
			ctx.Header.ParserSettings.FileFormatType,
			ctx.Content,
			finalOutputDecl)
		if err == errs.ErrSchemaNotSupported {
			continue
		}
		if err != nil {
			// error from FileFormat is already context formatted.
			return nil, err
		}
		return &schemaPlugin{
			ctx:             ctx,
			fileFormat:      fileFormat,
			formatRuntime:   formatRuntime,
			finalOutputDecl: finalOutputDecl,
		}, nil
	}
	return nil, errs.ErrSchemaNotSupported
}

func customParseFuncs(ctx *schemaplugin.ParseSchemaCtx) transform.CustomParseFuncs {
	if ctx.PluginParams == nil {
		return nil
	}
	params := ctx.PluginParams.(*PluginParams)
	if len(params.CustomParseFuncs) == 0 {
		return nil
	}
	return params.CustomParseFuncs
}

func fileFormats(ctx *schemaplugin.ParseSchemaCtx) []omniv2fileformat.FileFormat {
	// If caller specifies a custom FileFormat, we'll use it (and it only);
	// otherwise we'll use the builtin ones.
	formats := []omniv2fileformat.FileFormat{
		omniv2json.NewJSONFileFormat(ctx.Name),
		omniv2xml.NewXMLFileFormat(ctx.Name),
		// TODO more bulit-in omniv2 file formats to come.
	}
	if ctx.PluginParams != nil && ctx.PluginParams.(*PluginParams).CustomFileFormat != nil {
		formats = []omniv2fileformat.FileFormat{ctx.PluginParams.(*PluginParams).CustomFileFormat}
	}
	return formats
}

type schemaPlugin struct {
	ctx             *schemaplugin.ParseSchemaCtx
	fileFormat      omniv2fileformat.FileFormat
	formatRuntime   interface{}
	finalOutputDecl *transform.Decl
}

func (s *schemaPlugin) GetInputProcessor(ctx *transformctx.Ctx, input io.Reader) (schemaplugin.InputProcessor, error) {
	reader, err := s.fileFormat.CreateFormatReader(ctx.InputName, input, s.formatRuntime)
	if err != nil {
		return nil, err
	}
	return &inputProcessor{
		finalOutputDecl:  s.finalOutputDecl,
		customFuncs:      s.ctx.CustomFuncs,
		customParseFuncs: customParseFuncs(s.ctx),
		ctx:              ctx,
		reader:           reader,
	}, nil
}

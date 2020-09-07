package omniv2

import (
	"fmt"
	"io"

	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	omniv2fileformat "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat"
	omniv2xml "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat/xml"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/omniparser/schemavalidate"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

const (
	PluginVersion = "omni.2.0"
)

type PluginParams struct {
	CustomFileFormat omniv2fileformat.FileFormat
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
	finalOutputDecl, err := transform.ValidateTransformDeclarations(ctx.Content, ctx.CustomFuncs)
	if err != nil {
		return nil, fmt.Errorf(
			"schema '%s' 'transform_declarations' validation failed': %s",
			ctx.Name, err.Error())
	}
	// If caller specifies a custom FileFormat, we'll use it (and it only);
	// otherwise we'll use the builtin ones.
	fileFormats := []omniv2fileformat.FileFormat{
		omniv2xml.NewXMLFileFormat(ctx.Name),
	}
	if ctx.PluginParams != nil {
		params := ctx.PluginParams.(*PluginParams)
		if params != nil {
			fileFormats = []omniv2fileformat.FileFormat{params.CustomFileFormat}
		}
	}
	for _, fileFormat := range fileFormats {
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
		finalOutputDecl: s.finalOutputDecl,
		customFuncs:     s.ctx.CustomFuncs,
		ctx:             ctx,
		reader:          reader,
	}, nil
}

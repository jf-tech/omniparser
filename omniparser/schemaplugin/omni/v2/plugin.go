package omniv2

import (
	"io"

	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	omniv2fileformat "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat"
	omniv2xml "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat/xml"
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
			// TODO ('transform_declarations' parsing)
			nil)
		if err == errs.ErrSchemaNotSupported {
			continue
		}
		if err != nil {
			return nil, err
		}
		return &schemaPlugin{
			ctx:           ctx,
			fileFormat:    fileFormat,
			formatRuntime: formatRuntime,
		}, nil
	}
	return nil, errs.ErrSchemaNotSupported
}

type schemaPlugin struct {
	ctx           *schemaplugin.ParseSchemaCtx
	fileFormat    omniv2fileformat.FileFormat
	formatRuntime interface{}
}

func (s *schemaPlugin) GetInputProcessor(ctx *transformctx.Ctx, input io.Reader) (schemaplugin.InputProcessor, error) {
	reader, err := s.fileFormat.CreateFormatReader(ctx.InputName, input, s.formatRuntime)
	if err != nil {
		return nil, err
	}
	return &inputProcessor{
		finalOutputDecl: nil, // TODO ('transform_declarations' parsing)
		customFuncs:     s.ctx.CustomFuncs,
		ctx:             ctx,
		reader:          reader,
	}, nil
}

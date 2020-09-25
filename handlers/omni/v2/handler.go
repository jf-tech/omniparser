package omniv2

import (
	"fmt"
	"io"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/handlers"
	omniv2fileformat "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat"
	omniv2json "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat/json"
	omniv2xml "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat/xml"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
	"github.com/jf-tech/omniparser/validation"
	"github.com/jf-tech/omniparser/transformctx"
)

const (
	version = "omni.2.0"
)

// HandlerParams allows user of omniparser to provide omniv2 schema handler customization.
type HandlerParams struct {
	CustomFileFormat omniv2fileformat.FileFormat
	CustomParseFuncs transform.CustomParseFuncs
}

// CreateHandler parses, validates and creates an omni-schema based handler.
func CreateHandler(ctx *handlers.HandlerCtx) (handlers.SchemaHandler, error) {
	if ctx.Header.ParserSettings.Version != version {
		return nil, errs.ErrSchemaNotSupported
	}
	// First do a `transform_declarations` json schema validation
	err := validation.SchemaValidate(ctx.Name, ctx.Content, validation.JSONSchemaTransformDeclarations)
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
		return &schemaHandler{
			ctx:             ctx,
			fileFormat:      fileFormat,
			formatRuntime:   formatRuntime,
			finalOutputDecl: finalOutputDecl,
		}, nil
	}
	return nil, errs.ErrSchemaNotSupported
}

func customParseFuncs(ctx *handlers.HandlerCtx) transform.CustomParseFuncs {
	if ctx.HandlerParams == nil {
		return nil
	}
	params := ctx.HandlerParams.(*HandlerParams)
	if len(params.CustomParseFuncs) == 0 {
		return nil
	}
	return params.CustomParseFuncs
}

func fileFormats(ctx *handlers.HandlerCtx) []omniv2fileformat.FileFormat {
	// If caller specifies a custom FileFormat, we'll use it (and it only);
	// otherwise we'll use the builtin ones.
	formats := []omniv2fileformat.FileFormat{
		omniv2json.NewJSONFileFormat(ctx.Name),
		omniv2xml.NewXMLFileFormat(ctx.Name),
		// TODO more bulit-in omniv2 file formats to come.
	}
	if ctx.HandlerParams != nil && ctx.HandlerParams.(*HandlerParams).CustomFileFormat != nil {
		formats = []omniv2fileformat.FileFormat{ctx.HandlerParams.(*HandlerParams).CustomFileFormat}
	}
	return formats
}

type schemaHandler struct {
	ctx             *handlers.HandlerCtx
	fileFormat      omniv2fileformat.FileFormat
	formatRuntime   interface{}
	finalOutputDecl *transform.Decl
}

func (h *schemaHandler) NewIngester(ctx *transformctx.Ctx, input io.Reader) (handlers.Ingester, error) {
	reader, err := h.fileFormat.CreateFormatReader(ctx.InputName, input, h.formatRuntime)
	if err != nil {
		return nil, err
	}
	return &ingester{
		finalOutputDecl:  h.finalOutputDecl,
		customFuncs:      h.ctx.CustomFuncs,
		customParseFuncs: customParseFuncs(h.ctx),
		ctx:              ctx,
		reader:           reader,
	}, nil
}

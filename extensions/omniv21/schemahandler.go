package omniv21

import (
	"fmt"
	"io"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/csv"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/edi"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/fixedlength"
	csv2 "github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile/csv"
	fixedlength2 "github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile/fixedlength"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/json"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/xml"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	v21validation "github.com/jf-tech/omniparser/extensions/omniv21/validation"
	"github.com/jf-tech/omniparser/schemahandler"
	"github.com/jf-tech/omniparser/transformctx"
	"github.com/jf-tech/omniparser/validation"
)

const (
	version = "omni.2.1"
)

// CreateParams allows user of this 'omni.2.1' schema handler to provide creation customization.
type CreateParams struct {
	CustomFileFormats []fileformat.FileFormat
	// Deprecated.
	CustomParseFuncs transform.CustomParseFuncs
}

// CreateSchemaHandler parses, validates and creates an omni-schema based handler.
func CreateSchemaHandler(ctx *schemahandler.CreateCtx) (schemahandler.SchemaHandler, error) {
	if ctx.Header.ParserSettings.Version != version {
		return nil, errs.ErrSchemaNotSupported
	}
	// First do a `transform_declarations` json schema validation
	err := validation.SchemaValidate(ctx.Name, ctx.Content, v21validation.JSONSchemaTransformDeclarations)
	if err != nil {
		// err is already context formatted.
		return nil, err
	}
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		ctx.Content, ctx.CustomFuncs, customParseFuncs(ctx))
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

func customParseFuncs(ctx *schemahandler.CreateCtx) transform.CustomParseFuncs {
	if ctx.CreateParams == nil {
		return nil
	}
	params, ok := ctx.CreateParams.(*CreateParams)
	if !ok {
		return nil
	}
	if len(params.CustomParseFuncs) == 0 {
		return nil
	}
	return params.CustomParseFuncs
}

func fileFormats(ctx *schemahandler.CreateCtx) []fileformat.FileFormat {
	formats := []fileformat.FileFormat{
		csv.NewCSVFileFormat(ctx.Name),
		csv2.NewCSVFileFormat(ctx.Name),
		edi.NewEDIFileFormat(ctx.Name),
		fixedlength.NewFixedLengthFileFormat(ctx.Name),
		fixedlength2.NewFixedLengthFileFormat(ctx.Name),
		json.NewJSONFileFormat(ctx.Name),
		xml.NewXMLFileFormat(ctx.Name),
	}
	if ctx.CreateParams == nil {
		return formats
	}
	params, ok := ctx.CreateParams.(*CreateParams)
	if !ok {
		return formats
	}
	// If caller specifies a list of custom FileFormats, we'll give them priority
	// over builtin ones.
	return append(params.CustomFileFormats, formats...)
}

type schemaHandler struct {
	ctx             *schemahandler.CreateCtx
	fileFormat      fileformat.FileFormat
	formatRuntime   interface{}
	finalOutputDecl *transform.Decl
}

func (h *schemaHandler) NewIngester(ctx *transformctx.Ctx, input io.Reader) (schemahandler.Ingester, error) {
	reader, err := h.fileFormat.CreateFormatReader(h.ctx.Header, ctx.InputName, input, h.formatRuntime)
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

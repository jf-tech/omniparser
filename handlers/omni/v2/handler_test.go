package omniv2

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/handlers"
	omniv2fileformat "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat"
	omniv2json "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat/json"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/transformctx"
)

type testFileFormat struct {
	validateSchemaErr     error
	validateSchemaRuntime interface{}
	createFormatReaderErr error
}

func (f testFileFormat) ValidateSchema(_ string, _ []byte, _ *transform.Decl) (interface{}, error) {
	if f.validateSchemaErr != nil {
		return nil, f.validateSchemaErr
	}
	return f.validateSchemaRuntime, nil
}

func (f testFileFormat) CreateFormatReader(
	inputName string, input io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	if f.createFormatReaderErr != nil {
		return nil, f.createFormatReaderErr
	}
	return testFormatReader{
		inputName: inputName,
		input:     input,
		runtime:   runtime,
	}, nil
}

type testFormatReader struct {
	inputName string
	input     io.Reader
	runtime   interface{}
}

func (r testFormatReader) Read() (*node.Node, error)           { panic("implement me") }
func (r testFormatReader) IsContinuableError(error) bool       { panic("implement me") }
func (r testFormatReader) FmtErr(string, ...interface{}) error { panic("implement me") }

func TestCreateHandler_VersionNotSupported(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version: "12345",
				},
			},
		})
	assert.Error(t, err)
	assert.Equal(t, errs.ErrSchemaNotSupported, err)
	assert.Nil(t, p)
}

func TestCreateHandler_FormatNotSupported(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version:        version,
					FileFormatType: "unknown",
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
		})
	assert.Error(t, err)
	assert.Equal(t, errs.ErrSchemaNotSupported, err)
	assert.Nil(t, p)
}

func TestCreateHandler_TransformDeclarationsJSONValidationFailed(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Name: "test-schema",
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version:        version,
					FileFormatType: "xml",
				},
			},
			Content: []byte(`{"transform_declarations": {}}`),
		})
	assert.Error(t, err)
	assert.Equal(t,
		`schema 'test-schema' validation failed: transform_declarations: FINAL_OUTPUT is required`,
		err.Error())
	assert.Nil(t, p)
}

func TestCreateHandler_TransformDeclarationsInCodeValidationFailed(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Name: "test-schema",
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version:        version,
					FileFormatType: "xml",
				},
			},
			Content: []byte(
				`{
					"transform_declarations": {
						"FINAL_OUTPUT": { "template": "non-existing" }
					}
				}`),
		})
	assert.Error(t, err)
	assert.Equal(t,
		`schema 'test-schema' 'transform_declarations' validation failed: 'FINAL_OUTPUT' contains non-existing template reference 'non-existing'`,
		err.Error())
	assert.Nil(t, p)
}

func TestCreateHandler_HandlerParamsTypeNotRight_Fallback(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version:        version,
					FileFormatType: "json",
				},
			},
			Content:       []byte(`{"transform_declarations": { "FINAL_OUTPUT": { "xpath": "." }}}`),
			HandlerParams: "not nil but not the right type",
		})
	assert.NoError(t, err)
	assert.IsType(t, omniv2json.NewJSONFileFormat(""), p.(*schemaHandler).fileFormat)
	assert.Equal(t, ".", p.(*schemaHandler).formatRuntime.(string))
}

func TestCreateHandler_CustomFileFormat_FormatNotSupported_Fallback(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version:        version,
					FileFormatType: "json",
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": { "xpath": "." }}}`),
			HandlerParams: &HandlerParams{
				CustomFileFormats: []omniv2fileformat.FileFormat{
					// Having the custom FileFormat returns ErrSchemaNotSupported
					// causes it to fallback and continue probing the built-in FileFormats.
					testFileFormat{validateSchemaErr: errs.ErrSchemaNotSupported},
				},
			},
		})
	assert.NoError(t, err)
	assert.IsType(t, omniv2json.NewJSONFileFormat(""), p.(*schemaHandler).fileFormat)
	assert.Equal(t, ".", p.(*schemaHandler).formatRuntime.(string))
}

func TestCreateHandler_CustomFileFormat_ValidationFailure(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version: version,
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
			HandlerParams: &HandlerParams{
				CustomFileFormats: []omniv2fileformat.FileFormat{
					testFileFormat{validateSchemaErr: errors.New("validation failure")},
				},
			},
		})
	assert.Error(t, err)
	assert.Equal(t, "validation failure", err.Error())
	assert.Nil(t, p)
}

func TestCreateHandler_CustomFileFormat_Success(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version: version,
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
			HandlerParams: &HandlerParams{
				CustomFileFormats: []omniv2fileformat.FileFormat{
					testFileFormat{validateSchemaRuntime: "runtime data"},
				},
			},
		})
	assert.NoError(t, err)
	assert.IsType(t, &schemaHandler{}, p)
	assert.Equal(t, "runtime data", p.(*schemaHandler).formatRuntime.(string))
}

func TestCreateHandler_CustomParseFuncs_Success(t *testing.T) {
	p, err := CreateHandler(
		&handlers.HandlerCtx{
			Header: header.Header{
				ParserSettings: header.ParserSettings{
					Version:        version,
					FileFormatType: "xml",
				},
			},
			Content: []byte(`{
					"transform_declarations": {
						"FINAL_OUTPUT": { "xpath": "/A/B", "custom_parse": "test_custom_parse" }
					}
				}`),
			HandlerParams: &HandlerParams{
				CustomParseFuncs: transform.CustomParseFuncs{
					"test_custom_parse": func(_ *transformctx.Ctx, _n *node.Node) (interface{}, error) {
						return "test", nil
					},
				},
			},
		})
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestNewIngester_CustomFileFormat_Failure(t *testing.T) {
	ip, err := (&schemaHandler{
		fileFormat: testFileFormat{
			createFormatReaderErr: errors.New("failed to create reader"),
		},
	}).NewIngester(&transformctx.Ctx{InputName: "test-input"}, nil)
	assert.Error(t, err)
	assert.Equal(t, "failed to create reader", err.Error())
	assert.Nil(t, ip)
}

func TestNewIngester_CustomFileFormat_Success(t *testing.T) {
	handler := &schemaHandler{
		ctx: &handlers.HandlerCtx{
			CustomFuncs: customfuncs.BuiltinCustomFuncs,
		},
		fileFormat:    testFileFormat{},
		formatRuntime: "test runtime",
	}
	ctx := &transformctx.Ctx{InputName: "test-input"}
	ip, err := handler.NewIngester(ctx, strings.NewReader("test input"))
	assert.NoError(t, err)
	g := ip.(*ingester)
	assert.Equal(t, ctx, g.ctx)
	assert.Equal(t, customfuncs.BuiltinCustomFuncs, g.customFuncs)
	r := g.reader.(testFormatReader)
	assert.Equal(t, "test-input", r.inputName)
	data, err := ioutil.ReadAll(r.input)
	assert.NoError(t, err)
	assert.Equal(t, "test input", string(data))
	assert.Equal(t, "test runtime", r.runtime.(string))
}

package omniv2

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	omniv2fileformat "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

type testFileFormat struct {
	validateSchemaErr     error
	validateSchemaRuntime interface{}
	createFormatReaderErr error
}

func (t testFileFormat) ValidateSchema(_ string, _ []byte, _ *transform.Decl) (interface{}, error) {
	if t.validateSchemaErr != nil {
		return nil, t.validateSchemaErr
	}
	return t.validateSchemaRuntime, nil
}

func (t testFileFormat) CreateFormatReader(
	inputName string, input io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	if t.createFormatReaderErr != nil {
		return nil, t.createFormatReaderErr
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

func TestParseSchema_VersionNotSupported(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version: "12345",
				},
			},
		})
	assert.Error(t, err)
	assert.Equal(t, errs.ErrSchemaNotSupported, err)
	assert.Nil(t, p)
}

func TestParseSchema_FormatNotSupported(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version:        PluginVersion,
					FileFormatType: "unknown",
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
		})
	assert.Error(t, err)
	assert.Equal(t, errs.ErrSchemaNotSupported, err)
	assert.Nil(t, p)
}

func TestParseSchema_TransformDeclarationsJSONValidationFailed(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Name: "test-schema",
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version:        PluginVersion,
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

func TestParseSchema_TransformDeclarationsInCodeValidationFailed(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Name: "test-schema",
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version:        PluginVersion,
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
		`schema 'test-schema' 'transform_declarations' validation failed': 'FINAL_OUTPUT' contains non-existing template reference 'non-existing'`,
		err.Error())
	assert.Nil(t, p)
}

func TestParseSchema_CustomFileFormat_FormatNotSupported(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version: PluginVersion,
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
			PluginParams: &PluginParams{
				CustomFileFormat: testFileFormat{
					validateSchemaErr: errs.ErrSchemaNotSupported,
				},
			},
		})
	assert.Error(t, err)
	assert.Equal(t, "schema not supported", err.Error())
	assert.Nil(t, p)
}

func TestParseSchema_CustomFileFormat_ValidationFailure(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version: PluginVersion,
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
			PluginParams: &PluginParams{
				CustomFileFormat: testFileFormat{
					validateSchemaErr: errors.New("validation failure"),
				},
			},
		})
	assert.Error(t, err)
	assert.Equal(t, "validation failure", err.Error())
	assert.Nil(t, p)
}

func TestParseSchema_CustomFileFormat_Success(t *testing.T) {
	p, err := ParseSchema(
		&schemaplugin.ParseSchemaCtx{
			Header: schemaplugin.Header{
				ParserSettings: schemaplugin.ParserSettings{
					Version: PluginVersion,
				},
			},
			Content: []byte(`{"transform_declarations": { "FINAL_OUTPUT": {} }}`),
			PluginParams: &PluginParams{
				CustomFileFormat: testFileFormat{
					validateSchemaRuntime: "runtime data",
				},
			},
		})
	assert.NoError(t, err)
	plugin := p.(*schemaPlugin)
	assert.Equal(t, "runtime data", plugin.fileFormat.(testFileFormat).validateSchemaRuntime.(string))
	assert.Equal(t, "runtime data", plugin.formatRuntime.(string))
}

func TestGetInputProcessor_CustomFileFormat_Failure(t *testing.T) {
	ip, err := (&schemaPlugin{
		fileFormat: testFileFormat{
			createFormatReaderErr: errors.New("failed to create reader"),
		},
	}).GetInputProcessor(&transformctx.Ctx{InputName: "test-input"}, nil)
	assert.Error(t, err)
	assert.Equal(t, "failed to create reader", err.Error())
	assert.Nil(t, ip)
}

func TestGetInputProcessor_CustomFileFormat_Success(t *testing.T) {
	plugin := &schemaPlugin{
		ctx: &schemaplugin.ParseSchemaCtx{
			CustomFuncs: customfuncs.BuiltinCustomFuncs,
		},
		fileFormat:    testFileFormat{},
		formatRuntime: "test runtime",
	}
	ctx := &transformctx.Ctx{InputName: "test-input"}
	ip, err := plugin.GetInputProcessor(ctx, strings.NewReader("test input"))
	assert.NoError(t, err)
	processor := ip.(*inputProcessor)
	assert.Equal(t, ctx, processor.ctx)
	assert.Equal(t, customfuncs.BuiltinCustomFuncs, processor.customFuncs)
	r := processor.reader.(testFormatReader)
	assert.Equal(t, "test-input", r.inputName)
	data, err := ioutil.ReadAll(r.input)
	assert.NoError(t, err)
	assert.Equal(t, "test input", string(data))
	assert.Equal(t, "test runtime", r.runtime.(string))
}

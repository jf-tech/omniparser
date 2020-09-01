package omniparser

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/testlib"
)

func TestNewParser(t *testing.T) {
	for _, test := range []struct {
		name        string
		schema      string
		exts        []Extension
		expectedErr string
	}{
		{
			name:        "fail to read out schema content",
			schema:      "",
			exts:        nil,
			expectedErr: "unable to read schema 'test-schema': mock reading failure",
		},
		{
			name:        "fail to unmarshal schema header",
			schema:      "[invalid",
			exts:        nil,
			expectedErr: "unable to read schema 'test-schema': corrupted header `parser_settings`:.*",
		},
		{
			name:   "no supported schema plugin",
			schema: `{"parser_settings": {"version": "9999", "file_format_type": "exe" }}`,
			exts: []Extension{
				{}, // Empty extension to test if we skip empty extension properly or not.
				{
					ParseSchema: func(string, schemaplugin.Header, []byte) (schemaplugin.Plugin, error) {
						return nil, errs.ErrSchemaNotSupported
					},
				},
			},
			expectedErr: errs.ErrSchemaNotSupported.Error(),
		},
		{
			name:   "supported schema plugin found, but schema validation fails",
			schema: `{"parser_settings": {"version": "9999", "file_format_type": "exe" }}`,
			exts: []Extension{
				{
					ParseSchema: func(string, schemaplugin.Header, []byte) (schemaplugin.Plugin, error) {
						return nil, errors.New("invalid schema")
					},
				},
			},
			expectedErr: "invalid schema",
		},
		{
			name:   "supported schema plugin found, schema parsing successful",
			schema: `{"parser_settings": {"version": "9999", "file_format_type": "exe" }}`,
			exts: []Extension{
				{
					ParseSchema: func(string, schemaplugin.Header, []byte) (schemaplugin.Plugin, error) {
						return nil, nil
					},
				},
			},
			expectedErr: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var schemaReader io.Reader
			if test.schema == "" {
				schemaReader = testlib.NewMockReadCloser("mock reading failure", nil)
			} else {
				schemaReader = strings.NewReader(test.schema)
			}
			plugin, err := NewParser("test-schema", schemaReader, test.exts...)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Regexp(t, test.expectedErr, err.Error())
				assert.Nil(t, plugin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)
				assert.Equal(t, test.schema, string(plugin.SchemaContent()))
			}
		})
	}
}

func TestParser_GetTransformOp_StripBOMFailure(t *testing.T) {
	p := &parser{
		schemaHeader: schemaplugin.Header{
			ParserSettings: schemaplugin.ParserSettings{Version: "999", FileFormatType: "exe"},
		},
	}
	op, err := p.GetTransformOp("test input", testlib.NewMockReadCloser("bom read failure", nil), nil)
	assert.Error(t, err)
	assert.Equal(t, "bom read failure", err.Error())
	assert.Nil(t, op)
}

type testSchemaPlugin struct {
	getInputProcessorErr error
}

func (t testSchemaPlugin) GetInputProcessor(_ *transformctx.Ctx, _ io.Reader) (schemaplugin.InputProcessor, error) {
	if t.getInputProcessorErr != nil {
		return nil, t.getInputProcessorErr
	}
	return &testInputProcessor{}, nil
}

func TestParser_GetTransformOp_GetInputProcessorFailure(t *testing.T) {
	p := &parser{
		schemaHeader: schemaplugin.Header{
			ParserSettings: schemaplugin.ParserSettings{Version: "999", FileFormatType: "exe"},
		},
		schemaPlugin: testSchemaPlugin{getInputProcessorErr: errors.New("test failure")},
	}
	op, err := p.GetTransformOp("test input", strings.NewReader("something"), nil)
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())
	assert.Nil(t, op)
}

func TestParser_GetTransformOp_NameAndCtxAwareErrOverwrite(t *testing.T) {
	header := schemaplugin.Header{
		ParserSettings: schemaplugin.ParserSettings{Version: "999", FileFormatType: "exe"},
	}
	p := &parser{
		schemaHeader:  header,
		schemaContent: []byte("test schema content"),
		schemaPlugin:  testSchemaPlugin{},
	}
	ctx := &transformctx.Ctx{}
	op, err := p.GetTransformOp("test input", strings.NewReader("something"), ctx)
	assert.NoError(t, err)
	assert.NotNil(t, op)
	assert.Equal(t, "test input", ctx.InputName)
	assert.NotNil(t, ctx.CtxAwareErr)

	assert.Equal(t, header, p.SchemaHeader())
	assert.Equal(t, "test schema content", string(p.SchemaContent()))
}

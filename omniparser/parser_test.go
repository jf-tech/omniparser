package omniparser

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/testlib"
)

func TestNewParser(t *testing.T) {
	for _, test := range []struct {
		name        string
		schema      string
		pluginCfgs  []SchemaPluginConfig
		expectedErr string
	}{
		{
			name:        "fail to read out schema content",
			schema:      "",
			pluginCfgs:  nil,
			expectedErr: "unable to read schema 'test-schema': mock reading failure",
		},
		{
			name:        "fail to unmarshal schema header",
			schema:      "[invalid",
			pluginCfgs:  nil,
			expectedErr: "unable to perform schema validation: invalid character 'i' looking for beginning of value",
		},
		{
			name:   "no supported schema plugin",
			schema: `{"parser_settings": {"version": "9999", "file_format_type": "exe" }}`,
			pluginCfgs: []SchemaPluginConfig{
				{}, // Empty to test if we skip empty 3rd party plugin config properly or not.
				{
					ParseSchema: func(_ *schemaplugin.ParseSchemaCtx) (schemaplugin.Plugin, error) {
						return nil, errs.ErrSchemaNotSupported
					},
				},
			},
			expectedErr: errs.ErrSchemaNotSupported.Error(),
		},
		{
			name:        "supported schema plugin found, but json schema validation for parser_settings failed",
			schema:      `{"parser_settings": {"versionx": "9999", "file_format_type": "exe" }}`,
			pluginCfgs:  nil,
			expectedErr: "schema 'test-schema' validation failed:\nparser_settings: version is required\nparser_settings: Additional property versionx is not allowed",
		},
		{
			name:   "supported schema plugin found, but schema validation failed",
			schema: `{"parser_settings": {"version": "9999", "file_format_type": "exe" }}`,
			pluginCfgs: []SchemaPluginConfig{
				{
					ParseSchema: func(_ *schemaplugin.ParseSchemaCtx) (schemaplugin.Plugin, error) {
						return nil, errors.New("invalid schema")
					},
				},
			},
			expectedErr: "invalid schema",
		},
		{
			name:   "supported schema plugin found, schema parsing successful",
			schema: `{"parser_settings": {"version": "9999", "file_format_type": "exe" }}`,
			pluginCfgs: []SchemaPluginConfig{
				{
					CustomFuncs: customfuncs.CustomFuncs{
						"upper":            nil,
						"very_very_unique": func() {},
					},
					ParseSchema: func(ctx *schemaplugin.ParseSchemaCtx) (schemaplugin.Plugin, error) {
						// since there is a naming collision for "upper", totally this plugin only
						// adds one new custom func 'very_very_unique'.
						assert.Equal(t, len(customfuncs.BuiltinCustomFuncs)+1, len(ctx.CustomFuncs))
						// make sure the name-collided 'upper' is overwritten by builtin one.
						assert.NotNil(t, ctx.CustomFuncs["upper"])
						// make sure 'very_very_unique' is added.
						assert.NotNil(t, ctx.CustomFuncs["very_very_unique"])
						// make sure plugin param is passed in correctly.
						assert.Equal(t, 13, ctx.PluginParams.(int))
						return nil, nil
					},
					PluginParams: 13,
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
			plugin, err := NewParser("test-schema", schemaReader, test.pluginCfgs...)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
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

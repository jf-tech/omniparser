package omniparser

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
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
						return nil, schemaplugin.ErrSchemaNotSupported
					},
				},
			},
			expectedErr: "unsupported schema 'test-schema':.*",
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

func TestParser(t *testing.T) {
	header := schemaplugin.Header{
		ParserSettings: schemaplugin.ParserSettings{Version: "999", FileFormatType: "exe"},
	}
	p := &parser{
		schemaHeader:  header,
		schemaContent: []byte("test schema content"),
	}
	assert.Panics(t, func() {
		_, _ = p.GetTransformOp("name", nil, nil)
	})
	assert.Equal(t, header, p.SchemaHeader())
	assert.Equal(t, []byte("test schema content"), p.SchemaContent())
}

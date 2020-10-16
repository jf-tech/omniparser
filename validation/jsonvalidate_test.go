package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaValidate(t *testing.T) {
	for _, test := range []struct {
		name          string
		jsonSchema    string
		schemaContent string
		expectedErr   string
	}{
		{
			name:       "success",
			jsonSchema: JSONSchemaParserSettings,
			schemaContent: `{
					"parser_settings": {
						"version": "test-version",
						"file_format_type": "test-format",
						"encoding": "utf-8"
					}
				}`,
			expectedErr: "",
		},
		{
			name:       "invalid json",
			jsonSchema: ">>",
			schemaContent: `{
					"parser_settings": {
						"version": "test-version",
						"file_format_type": "test-format"
					}
				}`,
			expectedErr: `unable to perform schema validation: invalid character '>' looking for beginning of value`,
		},
		{
			name:       "invalid encoding",
			jsonSchema: JSONSchemaParserSettings,
			schemaContent: `{
					"parser_settings": {
						"version": "test-version",
						"file_format_type": "test-format",
						"encoding": "invalid"
					}
				}`,
			expectedErr: `schema 'test-schema' validation failed: parser_settings.encoding: parser_settings.encoding must be one of the following: "utf-8", "iso-8859-1", "windows-1252"`,
		},
		{
			name:       "multiple errors",
			jsonSchema: JSONSchemaParserSettings,
			schemaContent: `{
					"parser_settings": {
						"version": "test-version",
						"unknown": "blah"
					}
				}`,
			expectedErr: "schema 'test-schema' validation failed:\nparser_settings: Additional property unknown is not allowed\nparser_settings: file_format_type is required",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := SchemaValidate("test-schema", []byte(test.schemaContent), test.jsonSchema)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

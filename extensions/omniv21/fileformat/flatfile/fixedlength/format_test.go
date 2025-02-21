package fixedlength

import (
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/idr"
)

func TestValidateSchema(t *testing.T) {
	for _, test := range []struct {
		name        string
		format      string
		fileDecl    string
		finalOutput *transform.Decl
		err         string
	}{
		{
			name:        "not supported format",
			format:      "exe",
			fileDecl:    "",
			finalOutput: nil,
			err:         errs.ErrSchemaNotSupported.Error(),
		},
		{
			name:        "file_declaration JSON schema validation error",
			format:      fileFormatFixedLength,
			fileDecl:    `{}`,
			finalOutput: nil,
			err:         `schema 'test' validation failed: (root): file_declaration is required`,
		},
		{
			name:   "group with rows",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "rows": 42, "type": "envelope_group" }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes.0.type: file_declaration.envelopes.0.type does not match: \"envelope\"\nfile_declaration.envelopes.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "group with header",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "header": ".", "type": "envelope_group" }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes.0.type: file_declaration.envelopes.0.type does not match: \"envelope\"\nfile_declaration.envelopes.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "group with columns",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "type": "envelope_group", "columns": [] }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes.0.type: file_declaration.envelopes.0.type does not match: \"envelope\"\nfile_declaration.envelopes.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "both rows and header/footer envelopes",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "rows": 42, "header": "^42$" }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes.0: Additional property rows is not allowed\nfile_declaration.envelopes.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "multiple target envelopes",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "name": "e1", "is_target": true },
							{ "name": "e2", "is_target": true }
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': a second envelope/envelope_group ('e2') with 'is_target' = true is not allowed`,
		},
		{
			name:   "invalid rows",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [ { "name": "e1", "rows": 0 } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes.0.rows: Must be greater than or equal to 1\nfile_declaration.envelopes.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "invalid header regex",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [ { "name": "e1", "header": "[" } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': envelope/envelope_group 'e1' has an invalid 'header' regexp '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "invalid footer regex",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [ { "name": "e1", "header": ".", "footer": "[" } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': envelope/envelope_group 'e1' has an invalid 'footer' regexp '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "invalid type",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [ { "name": "e1", "type": "test" } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes.0.type: file_declaration.envelopes.0.type does not match: \"envelope\"\nfile_declaration.envelopes.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "min > max",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [ { "name": "e1", "min": 5, "max": 4 } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': envelope/envelope_group 'e1' has 'min' value 5 > 'max' value 4",
		},
		{
			name:   "invalid line_pattern regexp",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"name": "e1",
								"columns": [{ "name": "c1", "start_pos": 1, "length": 3, "line_pattern": "[" }]
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': envelope 'e1' column 'c1' has an invalid 'line_pattern' regexp '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "FINAL_OUTPUT decl is nil",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "columns": [{ "name": "c1", "start_pos": 1, "length": 3 }] }
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': 'FINAL_OUTPUT' is missing`,
		},
		{
			name:   "FINAL_OUTPUT xpath is invalid",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : []
					}
				}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr("[")},
			err:         `schema 'test': 'FINAL_OUTPUT.xpath' (value: '[') is invalid, err: expression must evaluate to a node-set`,
		},
		{
			name:   "success",
			format: fileFormatFixedLength,
			fileDecl: `
					{
						"file_declaration": {
							"envelopes" : [
								{
									"name": "e1", "type": "envelope_group", "min": 1, "max": 1,
									"child_envelopes": [
										{ "name": "e2", "max": 5, "columns": [{ "name": "c1", "start_pos": 1, "length": 3 }] },
										{ "name": "e2", "header": "^ABC$", "columns": [{ "name": "c2", "start_pos": 2, "length": 5, "line_pattern": "^H00" }] }
									]
								}
							]
						}
					}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr(".[c1 != 'skip']")},
			err:         "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := NewFixedLengthFileFormat("test").
				ValidateSchema(test.format, []byte(test.fileDecl), test.finalOutput)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, runtime)
			} else {
				assert.NoError(t, err)
				cupaloy.SnapshotT(t, jsons.BPM(runtime))
			}
		})
	}
}

func TestCreateFormatReader(t *testing.T) {
	test := func(finalOutputXPath *string) {
		format := NewFixedLengthFileFormat("test-schema")
		runtime, err := format.ValidateSchema(
			fileFormatFixedLength,
			[]byte(`
				{
					"file_declaration": {
						"envelopes" : [
							{
								"rows": 2,
								"columns": [
									{ "name": "letters", "start_pos": 1, "length": 3, "line_pattern": "^[a-z]" },
									{ "name": "numerics", "start_pos": 1, "length": 3, "line_pattern": "^[0-9]" }
								]
							}
						]
					}
				}`),
			&transform.Decl{XPath: finalOutputXPath})
		assert.NoError(t, err)
		reader, err := format.CreateFormatReader(
			header.Header{},
			"test-input",
			strings.NewReader("abcd\n1234\n"),
			runtime)
		assert.NoError(t, err)
		n, err := reader.Read()
		assert.NoError(t, err)
		assert.Equal(t, `{"letters":"abc","numerics":"123"}`, idr.JSONify2(n))
		reader.Release(n)
		n, err = reader.Read()
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, n)
	}
	test(nil)                                // test without FINAL_OUTPUT xpath filtering.
	test(strs.StrPtr(".[letters != 'xyz']")) // test with FINAL_OUTPUT xpath filtering.

	// test CreateFormatReader called with invalid target xpath.
	reader, err := NewFixedLengthFileFormat("test-schema").CreateFormatReader(
		header.Header{},
		"test-input",
		strings.NewReader("abcd\n1234\n"),
		&fixedLengthFormatRuntime{XPath: "["})
	assert.Error(t, err)
	assert.Equal(t,
		"schema 'test-schema': xpath '[' on 'FINAL_OUTPUT' is invalid: expression must evaluate to a node-set",
		err.Error())
	assert.Nil(t, reader)
}

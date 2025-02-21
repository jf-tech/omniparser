package csv

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
			format:      fileFormatCSV,
			fileDecl:    `{}`,
			finalOutput: nil,
			err:         `schema 'test' validation failed: (root): file_declaration is required`,
		},
		{
			name:   "group with rows",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{ "rows": 42, "type": "record_group" }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.records.0.type: file_declaration.records.0.type does not match: \"record\"\nfile_declaration.records.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "group with header",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{ "header": ".", "type": "record_group" }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.records.0.type: file_declaration.records.0.type does not match: \"record\"\nfile_declaration.records.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "group with columns",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{ "type": "record_group", "columns": [] }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.records.0.type: file_declaration.records.0.type does not match: \"record\"\nfile_declaration.records.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "both rows and header/footer records",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{ "rows": 42, "header": "^42$" }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.records.0: Additional property rows is not allowed\nfile_declaration.records.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "multiple target records",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{ "name": "e1", "is_target": true },
							{ "name": "e2", "is_target": true }
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': a second record/record_group ('e2') with 'is_target' = true is not allowed`,
		},
		{
			name:   "invalid rows",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [ { "name": "e1", "rows": 0 } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.records.0.rows: Must be greater than or equal to 1\nfile_declaration.records.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "invalid header regex",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [ { "name": "e1", "header": "[" } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': record/record_group 'e1' has an invalid 'header' regexp '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "invalid footer regex",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [ { "name": "e1", "header": ".", "footer": "[" } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': record/record_group 'e1' has an invalid 'footer' regexp '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "invalid type",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [ { "name": "e1", "type": "test" } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.records.0.type: file_declaration.records.0.type does not match: \"record\"\nfile_declaration.records.0: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "min > max",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [ { "name": "e1", "min": 5, "max": 4 } ]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': record/record_group 'e1' has 'min' value 5 > 'max' value 4",
		},
		{
			name:   "invalid line_pattern regexp",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{
								"name": "e1",
								"columns": [{ "name": "c1", "line_pattern": "[" }]
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': record 'e1' column 'c1' has an invalid 'line_pattern' regexp '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "FINAL_OUTPUT decl is nil",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : [
							{ "columns": [{ "name": "c1", "index": 1 }] }
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': 'FINAL_OUTPUT' is missing`,
		},
		{
			name:   "FINAL_OUTPUT xpath is invalid",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"records" : []
					}
				}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr("[")},
			err:         `schema 'test': 'FINAL_OUTPUT.xpath' (value: '[') is invalid, err: expression must evaluate to a node-set`,
		},
		{
			name:   "success",
			format: fileFormatCSV,
			fileDecl: `
					{
						"file_declaration": {
							"delimiter": ",",
							"records" : [
								{
									"name": "e1", "type": "record_group", "min": 1, "max": 1,
									"child_records": [
										{ "name": "e2", "max": 5, "columns": [{ "name": "c1", "index": 2 }] },
										{ "name": "e2", "header": "^ABC$", "columns": [{ "name": "c2", "line_pattern": "^H00" }] }
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
			runtime, err := NewCSVFileFormat("test").
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
		format := NewCSVFileFormat("test-schema")
		runtime, err := format.ValidateSchema(
			fileFormatCSV,
			[]byte(`
				{
					"file_declaration": {
						"delimiter": "|",
						"records" : [
							{
								"rows": 2,
								"columns": [
									{ "name": "letters", "index": 2, "line_pattern": "^[a-z]" },
									{ "name": "numerics", "line_pattern": "^[0-9]" }
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
			strings.NewReader("abcd|efgh|jklm\n123|456|789\n"),
			runtime)
		assert.NoError(t, err)
		n, err := reader.Read()
		assert.NoError(t, err)
		assert.Equal(t, `{"letters":"efgh","numerics":"789"}`, idr.JSONify2(n))
		reader.Release(n)
		n, err = reader.Read()
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, n)
	}
	test(nil)                                // test without FINAL_OUTPUT xpath filtering.
	test(strs.StrPtr(".[letters != 'xyz']")) // test with FINAL_OUTPUT xpath filtering.

	// test CreateFormatReader called with invalid target xpath.
	reader, err := NewCSVFileFormat("test-schema").CreateFormatReader(
		header.Header{},
		"test-input",
		strings.NewReader("abcd\n1234\n"),
		&csvFormatRuntime{XPath: "["})
	assert.Error(t, err)
	assert.Equal(t,
		"schema 'test-schema': xpath '[' on 'FINAL_OUTPUT' is invalid: expression must evaluate to a node-set",
		err.Error())
	assert.Nil(t, reader)
}

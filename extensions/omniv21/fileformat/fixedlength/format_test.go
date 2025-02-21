package fixedlength

import (
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
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
			name:   "no envelope",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
                		"envelopes" : []
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes: Array must have at least 1 items\nfile_declaration.envelopes: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "two 'by_rows' envelopes",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
                		"envelopes" : [
							{ "columns": [{ "name": "abc", "start_pos": 1, "length": 3 }] },
		                    { "by_rows": 11, "columns": [{ "name": "efg", "start_pos": 1, "length": 3 }] }
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test' validation failed:\nfile_declaration.envelopes: Array must have at most 1 items\nfile_declaration.envelopes: Must validate one and only one schema (oneOf)",
		},
		{
			name:   "multiple target envelopes",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"by_header_footer": { "header": ".", "footer": "." },
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3 }]
							},
							{
								"by_header_footer": { "header": ".", "footer": "." },
								"columns": [{ "name": "efg", "start_pos": 1, "length": 3 }],
								"not_target": false
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': cannot have more than one target envelope`,
		},
		{
			name:   "missing target envelope",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"by_header_footer": { "header": ".", "footer": "." },
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3 }],
								"not_target": true
							},
							{
								"by_header_footer": { "header": ".", "footer": "." },
								"columns": [{ "name": "efg", "start_pos": 1, "length": 3 }],
								"not_target": true
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': missing target envelope`,
		},
		{
			name:   "duplicate envelope names",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"name": "dup",
								"by_header_footer": { "header": ".", "footer": "." },
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3 }]
							},
							{
								"name": "dup",
								"by_header_footer": { "header": ".", "footer": "." },
								"columns": [{ "name": "efg", "start_pos": 1, "length": 3 }],
								"not_target": true
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': more than one envelope has the name 'dup'`,
		},
		{
			name:   "invalid header regex",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"by_header_footer": { "header": "[", "footer": "." },
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3 }]
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': invalid 'header' regex '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "invalid footer regex",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"by_header_footer": { "header": ".", "footer": "[" },
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3 }]
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': invalid 'footer' regex '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "dup column name",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"columns": [
									{ "name": "abc", "start_pos": 1, "length": 3 },
									{ "name": "abc", "start_pos": 4, "length": 6 }
								]
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': more than one column has the name 'abc'`,
		},
		{
			name:   "invalid line_pattern regex",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3, "line_pattern": "[" }]
							}
						]
					}
				}`,
			finalOutput: nil,
			err:         "schema 'test': invalid 'line_pattern' regex '[': error parsing regexp: missing closing ]: `[`",
		},
		{
			name:   "FINAL_OUTPUT decl is nil",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "columns": [{ "name": "abc", "start_pos": 1, "length": 3 }] }
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
						"envelopes" : [
							{ "columns": [{ "name": "abc", "start_pos": 1, "length": 3 }] }
						]
					}
				}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr("[")},
			err:         `schema 'test': 'FINAL_OUTPUT.xpath' (value: '[') is invalid, err: expression must evaluate to a node-set`,
		},
		{
			name:   "success - simple by_rows=1",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{ "columns": [{ "name": "abc", "start_pos": 1, "length": 10 }] }
						]
					}
				}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr(".[abc != 'skip']")},
			err:         "",
		},
		{
			name:   "success - by_rows=3",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"by_rows": 3,
								"columns": [
									{ "name": "abc", "start_pos": 1, "length": 10, "line_pattern": "^L01.*" },
									{ "name": "efg", "start_pos": 3, "length": 5, "line_pattern": "^L03.*" }
								]
							}
						]
					}
				}`,
			finalOutput: &transform.Decl{},
			err:         "",
		},
		{
			name:   "success - by_header_footer",
			format: fileFormatFixedLength,
			fileDecl: `
				{
					"file_declaration": {
						"envelopes" : [
							{
								"by_header_footer": { "header": "^FILE-BEGIN$", "footer": "^FILE-BEGIN$" },
								"not_target": true
							},
							{
								"by_header_footer": { "header": "^DATA-BLOCK-BEGIN$", "footer": "^DATA-BLOCK-END$" },
								"columns": [{ "name": "abc", "start_pos": 1, "length": 3, "line_pattern": "^DATA:.*$" }]
							},
							{
								"by_header_footer": { "header": "^FILE-END$", "footer": "^FILE-END$" },
								"not_target": true
							}
						]
					}
				}`,
			finalOutput: &transform.Decl{},
			err:         "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := NewFixedLengthFileFormat(
				"test").ValidateSchema(test.format, []byte(test.fileDecl), test.finalOutput)
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
	r, err := NewFixedLengthFileFormat("test").CreateFormatReader(
		header.Header{},
		"test",
		strings.NewReader("abcd\n1234\n"),
		&fixedLengthFormatRuntime{
			Decl: &FileDecl{
				Envelopes: []*EnvelopeDecl{
					{
						Name:   strs.StrPtr("env1"),
						ByRows: testlib.IntPtr(2),
						Columns: []*ColumnDecl{
							{Name: "letters", StartPos: 1, Length: 3, LinePattern: strs.StrPtr("^[a-z]")},
							{Name: "numerics", StartPos: 1, Length: 3, LinePattern: strs.StrPtr("^[0-9]")},
						},
					},
				},
			},
		})
	assert.NoError(t, err)
	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `{"letters":"abc","numerics":"123"}`, idr.JSONify2(n))
	r.Release(n)
	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

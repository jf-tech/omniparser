package csv

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
			format:      fileFormatCSV,
			fileDecl:    `{}`,
			finalOutput: nil,
			err:         `schema 'test' validation failed: (root): file_declaration is required`,
		},
		{
			name:   "file_declaration.header_row_index >= file_declaration.data_row_index",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"header_row_index": 2,
						"data_row_index": 2,
						"columns": [ { "name": "col1" } ]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': file_declaration.header_row_index(2) must be smaller than file_declaration.data_row_index(2)`,
		},
		{
			name:   "file_declaration.columns has duplicate names",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"data_row_index": 2,
						"columns": [ { "name": "col1" }, { "name": "col1" } ]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': file_declaration.columns contains duplicate name 'col1'`,
		},
		{
			name:   "file_declaration.columns has duplicate aliases",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"data_row_index": 2,
						"columns": [ { "name": "col1", "alias": "a1" }, { "name": "col2", "alias": "a1" } ]
					}
				}`,
			finalOutput: nil,
			err:         `schema 'test': file_declaration.columns contains duplicate alias 'a1'`,
		},
		{
			name:   "FINAL_OUTPUT decl is nil",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"data_row_index": 2,
						"columns": [ { "name": "col1" } ]
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
						"data_row_index": 2,
						"columns": [ { "name": "col1" } ]
					}
				}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr("[")},
			err:         `schema 'test': 'FINAL_OUTPUT.xpath' (value: '[') is invalid, err: expression must evaluate to a node-set`,
		},
		{
			name:   "success 1",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"replace_double_quotes": true,
						"header_row_index": 2,
						"data_row_index": 4,
						"columns": [ { "name": "col1" }, { "name": "col  2", "alias": "col2" } ]
					}
				}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr(".[col1 != 'skip']")},
			err:         "",
		},
		{
			name:   "success 2",
			format: fileFormatCSV,
			fileDecl: `
				{
					"file_declaration": {
						"delimiter": ",",
						"data_row_index": 1,
						"columns": [ { "name": "col1" } ]
					}
				}`,
			finalOutput: &transform.Decl{},
			err:         "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := NewCSVFileFormat(
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
	r, err := NewCSVFileFormat("test").CreateFormatReader(
		header.Header{},
		"test-input",
		strings.NewReader(
			lf("A|B|C")+
				lf("1|2|3")+
				lf("x|y")+
				lf("4|5|6")),
		&csvFormatRuntime{
			Decl: &FileDecl{
				Delimiter:      "|",
				HeaderRowIndex: testlib.IntPtr(1),
				DataRowIndex:   2,
				Columns:        []Column{{Name: "A"}, {Name: "B"}, {Name: "C"}},
			},
			XPath: ".[A != 'x']",
		})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	n, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `{"A":"1","B":"2","C":"3"}`, idr.JSONify2(n))
	n, err = r.Read()
	assert.NoError(t, err)
	assert.Equal(t, `{"A":"4","B":"5","C":"6"}`, idr.JSONify2(n))
	n, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

package edi

import (
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

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
			name:        "format not supported",
			format:      "mp3",
			fileDecl:    "",
			finalOutput: nil,
			err:         "schema not supported",
		},
		{
			name:   "json schema validation fail",
			format: fileFormatEDI,
			fileDecl: `{
				"file_declaration": {
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": -2,
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}
			}`,
			finalOutput: nil,
			err:         `schema 'test' validation failed: file_declaration.segment_declarations.0.max: Must be greater than or equal to -1`,
		},
		{
			name:   "in code schema validation fail",
			format: fileFormatEDI,
			fileDecl: `{
				"file_declaration": {
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": 0,
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}
			}`,
			finalOutput: nil,
			err:         `schema 'test': segment 'ISA' has 'min' value 1 > 'max' value 0`,
		},
		{
			name:   "FINAL_OUTPUT is nil",
			format: fileFormatEDI,
			fileDecl: `{
				"file_declaration": {
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}
			}`,
			finalOutput: nil,
			err:         `schema 'test': 'FINAL_OUTPUT' is missing`,
		},
		{
			name:   "FINAL_OUTPUT xpath is invalid",
			format: fileFormatEDI,
			fileDecl: `{
				"file_declaration": {
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}
			}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr("[")},
			err:         `schema 'test': 'FINAL_OUTPUT.xpath' (value: '[') is invalid, err: expression must evaluate to a node-set`,
		},
		{
			name:   "success",
			format: fileFormatEDI,
			fileDecl: `{
				"file_declaration": {
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"elements": [
								{ "name": "e2", "index": 2 },
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}
			}`,
			finalOutput: &transform.Decl{XPath: strs.StrPtr(".")},
			err:         ``,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			rt, err := NewEDIFileFormat("test").ValidateSchema(test.format, []byte(test.fileDecl), test.finalOutput)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, rt)
			} else {
				assert.NoError(t, err)
				cupaloy.SnapshotT(t, jsons.BPM(rt))
			}
		})
	}
}

func TestCreateFormatReader(t *testing.T) {
	format := NewEDIFileFormat("test")
	fileDecl := `{
		"file_declaration": {
			"segment_delimiter": "\n",
			"element_delimiter": "*",
			"segment_declarations": [
				{
					"name": "ISA",
					"is_target": true,
					"elements": [
						{ "name": "e3", "index": 3 },
						{ "name": "e1", "index": 1 }
					]
				}
			]
		}
	}`
	rt, err := format.ValidateSchema(fileFormatEDI, []byte(fileDecl), &transform.Decl{XPath: strs.StrPtr(".")})
	assert.NoError(t, err)
	reader, err := format.CreateFormatReader(
		header.Header{}, "test", strings.NewReader("ISA*e1*e2*e3\nISA*e4*e5*e6\n"), rt)
	assert.NoError(t, err)
	n, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, `{"e1":"e1","e3":"e3"}`, idr.JSONify2(n))
	reader.Release(n)
	n, err = reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, `{"e1":"e4","e3":"e6"}`, idr.JSONify2(n))
	reader.Release(n)
	n, err = reader.Read()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

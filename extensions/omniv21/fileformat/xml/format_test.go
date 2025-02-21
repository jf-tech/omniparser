package xml

import (
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
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
		decl        *transform.Decl
		expected    interface{}
		expectedErr string
	}{
		{
			name:        "not supported format",
			format:      "exe",
			decl:        nil,
			expected:    nil,
			expectedErr: errs.ErrSchemaNotSupported.Error(),
		},
		{
			name:        "FINAL_OUTPUT decl is nil",
			format:      fileFormatXML,
			decl:        nil,
			expected:    nil,
			expectedErr: `schema 'test-schema': 'FINAL_OUTPUT' is missing`,
		},
		{
			name:        "FINAL_OUTPUT 'xpath' is invalid",
			format:      fileFormatXML,
			decl:        &transform.Decl{XPath: strs.StrPtr("[invalid")},
			expected:    nil,
			expectedErr: `schema 'test-schema': 'FINAL_OUTPUT.xpath' (value: '[invalid') is invalid, err: expression must evaluate to a node-set`,
		},
		{
			name:        "success 1",
			format:      fileFormatXML,
			decl:        &transform.Decl{XPath: strs.StrPtr("/A/B[.!='skip']")},
			expected:    "/A/B[.!='skip']",
			expectedErr: "",
		},
		{
			name:        "success 2",
			format:      fileFormatXML,
			decl:        &transform.Decl{},
			expected:    ".",
			expectedErr: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := NewXMLFileFormat("test-schema").ValidateSchema(test.format, nil, test.decl)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Nil(t, runtime)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, runtime)
			}
		})
	}
}

func TestCreateFormatReader(t *testing.T) {
	r, err := NewXMLFileFormat("test-schema").CreateFormatReader(
		header.Header{},
		"test-input",
		strings.NewReader(`<A><B>data1</B><B>skip</B><B>data2</B></A>`),
		"/A/B[.!='skip']")
	assert.NoError(t, err)
	assert.NotNil(t, r)
	t.Run("B1", func(t *testing.T) {
		n1, err := r.Read()
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, idr.JSONify1(n1))
	})
	t.Run("B2", func(t *testing.T) {
		n2, err := r.Read()
		assert.NoError(t, err)
		cupaloy.SnapshotT(t, idr.JSONify1(n2))
	})
	t.Run("EOF", func(t *testing.T) {
		n3, err := r.Read()
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, n3)
	})

	r, err = NewXMLFileFormat("test-schema").CreateFormatReader(
		header.Header{}, "test-input", strings.NewReader(""), "[invalid")
	assert.Error(t, err)
	assert.Equal(t, `invalid xpath '[invalid', err: expression must evaluate to a node-set`, err.Error())
	assert.Nil(t, r)
}

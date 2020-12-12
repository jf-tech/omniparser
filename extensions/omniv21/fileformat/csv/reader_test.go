package csv

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

func TestIsErrInvalidHeader(t *testing.T) {
	assert.True(t, IsErrInvalidHeader(ErrInvalidHeader("test")))
	assert.Equal(t, "test", ErrInvalidHeader("test").Error())
	assert.False(t, IsErrInvalidHeader(errors.New("test")))
}

func TestNewReader_InvalidXPath(t *testing.T) {
	r, err := NewReader("test", nil, nil, "[invalid")
	assert.Error(t, err)
	assert.Equal(t, `invalid xpath '[invalid', err: expression must evaluate to a node-set`, err.Error())
	assert.Nil(t, r)
}

func lf(s string) string {
	return s + "\n"
}

func TestReader(t *testing.T) {
	for _, test := range []struct {
		name     string
		decl     *FileDecl
		xpath    string
		input    io.Reader
		expected []interface{}
	}{
		{
			name: "header row; alias used; with xpath; variable data row column size; replace double quote",
			decl: &FileDecl{
				Delimiter:           "|",
				ReplaceDoubleQuotes: true,
				HeaderRowIndex:      testlib.IntPtr(2),
				DataRowIndex:        4,
				Columns: []Column{
					{Name: "a"},
					{Name: "b with space", Alias: strs.StrPtr("b_with_space")},
					{Name: "c"},
				},
			},
			xpath: ".[a != 'skip']",
			input: strings.NewReader(lf("line1") +
				lf("a|b with space|c") +
				lf("line3") +
				lf(`1|"2|3`) + // put an unescaped double quote here to see if our replacement works or not.
				lf("") + // auto skipped by csv reader.
				lf("skip") + // skipped by the xpath.
				lf("one|two")),
			expected: []interface{}{
				`{ "a": "1", "b_with_space": "'2", "c": "3" }`,
				`{ "a": "one", "b_with_space": "two" }`,
			},
		},
		{
			name: "cannot jump to header row",
			decl: &FileDecl{
				Delimiter:      "|",
				HeaderRowIndex: testlib.IntPtr(3),
			},
			input: strings.NewReader(lf("line1")),
			expected: []interface{}{
				ErrInvalidHeader("input 'test-input' line 2: unable to read header: EOF"),
			},
		},
		{
			name: "cannot read to header row",
			decl: &FileDecl{
				Delimiter:      "|",
				HeaderRowIndex: testlib.IntPtr(2),
			},
			input: strings.NewReader(lf("line1")),
			expected: []interface{}{
				ErrInvalidHeader("input 'test-input' line 2: unable to read header: EOF"),
			},
		},
		{
			name: "header columns less than the declared",
			decl: &FileDecl{
				Delimiter:      "|",
				HeaderRowIndex: testlib.IntPtr(2),
				Columns: []Column{
					{Name: "a"},
					{Name: "b"},
				},
			},
			input: strings.NewReader(lf("line1") + lf("a")),
			expected: []interface{}{
				ErrInvalidHeader("input 'test-input' line 2: actual header column size (1) is less than the size (2) declared in file_declaration.columns in schema"),
			},
		},
		{
			name: "header column not matching the declared",
			decl: &FileDecl{
				Delimiter:      "|",
				HeaderRowIndex: testlib.IntPtr(2),
				Columns: []Column{
					{Name: "a"},
					{Name: "b"},
				},
			},
			input: strings.NewReader(lf("line1") + lf("a|B")),
			expected: []interface{}{
				ErrInvalidHeader("input 'test-input' line 2: header column[2] 'B' does not match declared column name 'b' in schema"),
			},
		},
		{
			name: "header row but no data row",
			decl: &FileDecl{
				Delimiter:      ",",
				HeaderRowIndex: testlib.IntPtr(2),
				DataRowIndex:   5,
				Columns: []Column{
					{Name: "a"},
					{Name: "b"},
				},
			},
			input:    strings.NewReader(lf("line1") + lf("a,b") + lf("line3")),
			expected: nil,
		},
		{
			name: "unable to read data row",
			decl: &FileDecl{
				Delimiter:    ",",
				DataRowIndex: 1,
				Columns:      []Column{{Name: "a"}},
			},
			input: testlib.NewMockReadCloser("read failure", nil),
			expected: []interface{}{
				errors.New("input 'test-input' line 1: failed to fetch record: read failure"),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, err := NewReader("test-input", test.input, test.decl, test.xpath)
			assert.NoError(t, err)
			for {
				n, err := r.Read()
				if err == io.EOF {
					assert.Equal(t, 0, len(test.expected))
					assert.Nil(t, n)
					break
				}
				assert.True(t, len(test.expected) > 0)
				if expectedErr, ok := test.expected[0].(error); ok {
					assert.Error(t, err)
					assert.Equal(t, expectedErr, err)
					assert.Nil(t, n)
					assert.Equal(t, 1, len(test.expected)) // if there is an error, it will be the last one.
					break
				}
				expectedJSON, ok := test.expected[0].(string)
				assert.True(t, ok)
				assert.Equal(t, jsons.BPJ(expectedJSON), jsons.BPJ(idr.JSONify2(n)))
				r.Release(n)
				test.expected = test.expected[1:]
			}
		})
	}
}

func TestIsContinuableError(t *testing.T) {
	r := &reader{}
	assert.True(t, r.IsContinuableError(errors.New("some error")))
	assert.False(t, r.IsContinuableError(ErrInvalidHeader("invalid header")))
	assert.False(t, r.IsContinuableError(io.EOF))
}

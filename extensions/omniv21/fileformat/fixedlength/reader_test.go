package fixedlength

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestIsErrInvalidEnvelope(t *testing.T) {
	assert.True(t, IsErrInvalidEnvelope(ErrInvalidEnvelope("test")))
	assert.Equal(t, "test", ErrInvalidEnvelope("test").Error())
	assert.False(t, IsErrInvalidEnvelope(errors.New("test")))
}

func testReader(t *testing.T, r io.Reader, decl *fileDecl) *reader {
	return &reader{
		inputName:     "test",
		r:             bufio.NewReader(r),
		decl:          decl,
		line:          1,
		envelopeLines: make([][]byte, 0, EnvelopeLinesCap),
	}
}

func TestReadLine(t *testing.T) {
	r := testReader(t, strings.NewReader("abc\n\nefg\n   \nxyz\n"), nil)
	assert.Equal(t, 1, r.line)

	line, err := r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, []byte("abc"), line)
	assert.Equal(t, 2, r.line)

	// the second read will skip a truly empty line.
	line, err = r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, []byte("efg"), line)
	assert.Equal(t, 4, r.line)

	// next line is not truly empty, it contains just spaces, we need to read it in.
	line, err = r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, []byte("   "), line)
	assert.Equal(t, 5, r.line)

	line, err = r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, []byte("xyz"), line)
	assert.Equal(t, 6, r.line)

	// io.EOF shouldn't bump up current line number.
	line, err = r.readLine()
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 6, r.line)

	// reading again should still return io.EOF and line number stays.
	line, err = r.readLine()
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 6, r.line)

	// Another scenario that io.Reader fails
	r = testReader(t, testlib.NewMockReadCloser("read error", nil), nil)
	assert.Equal(t, 1, r.line)
	line, err = r.readLine()
	assert.Error(t, err)
	assert.Equal(t, "read error", err.Error())
	assert.Nil(t, line)
	// reading error (unless it's EOF) bumps current line number
	assert.Equal(t, 2, r.line)
}

func TestReadByRowsEnvelope_ByRowsDefault(t *testing.T) {
	// default by_rows = 1
	r := testReader(t, strings.NewReader("abc\n\nefg\n   \nxyz\n"), &fileDecl{Envelopes: []*envelopeDecl{{}}})

	lines, err := r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("abc")}, lines)

	lines, err = r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("efg")}, lines)

	lines, err = r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("   ")}, lines)

	lines, err = r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("xyz")}, lines)

	lines, err = r.readByRowsEnvelope()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, lines)
}

func TestReadByRowsEnvelope_ByRowsNonDefault(t *testing.T) {
	r := testReader(t, strings.NewReader("abc\n\nefg\n   \nxyz\n"),
		&fileDecl{Envelopes: []*envelopeDecl{{ByRows: testlib.IntPtr(3)}}})

	lines, err := r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("abc"), []byte("efg"), []byte("   ")}, lines)

	lines, err = r.readByRowsEnvelope()
	assert.Error(t, err)
	assert.Equal(t, "input 'test' line 6: envelope incomplete: EOF", err.Error())
	assert.Nil(t, lines)
}

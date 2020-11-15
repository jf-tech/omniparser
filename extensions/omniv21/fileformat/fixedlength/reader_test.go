package fixedlength

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

func TestIsErrInvalidEnvelope(t *testing.T) {
	assert.True(t, IsErrInvalidEnvelope(ErrInvalidEnvelope("test")))
	assert.Equal(t, "test", ErrInvalidEnvelope("test").Error())
	assert.False(t, IsErrInvalidEnvelope(errors.New("test")))
}

func testReader(r io.Reader, decl *fileDecl) *reader {
	return &reader{
		inputName: "test",
		r:         bufio.NewReader(r),
		decl:      decl,
		line:      1,
	}
}

func TestReadLine(t *testing.T) {
	r := testReader(strings.NewReader("abc\n\nefg\n   \nxyz\n"), nil)
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
	r = testReader(testlib.NewMockReadCloser("read error", nil), nil)
	assert.Equal(t, 1, r.line)
	line, err = r.readLine()
	assert.Error(t, err)
	assert.Equal(t, "read error", err.Error())
	assert.Nil(t, line)
	assert.Equal(t, 1, r.line)
}

func TestReadByRowsEnvelope_ByRowsDefault(t *testing.T) {
	// default by_rows = 1
	r := testReader(strings.NewReader("abc\n\nefghijklmn\n   \nxyz\n"),
		&fileDecl{Envelopes: []*envelopeDecl{{
			Name: strs.StrPtr("env1"),
			Columns: []*columnDecl{
				{
					Name:     "col1",
					StartPos: 2,
					Length:   4,
				},
			},
		}}})

	n, err := r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, `{"col1":"bc"}`, idr.JSONify2(n))
	assert.Equal(t, 2, r.line)

	n, err = r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, `{"col1":"fghi"}`, idr.JSONify2(n))
	assert.Equal(t, 4, r.line)

	n, err = r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, `{"col1":"  "}`, idr.JSONify2(n))
	assert.Equal(t, 5, r.line)

	n, err = r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, `{"col1":"yz"}`, idr.JSONify2(n))
	assert.Equal(t, 6, r.line)

	n, err = r.readByRowsEnvelope()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestReadByRowsEnvelope_ByRowsNonDefault(t *testing.T) {
	r := testReader(strings.NewReader("abcdefg\n\nhijklmn\n   \nabc012345\n"),
		&fileDecl{Envelopes: []*envelopeDecl{{
			Name:   strs.StrPtr("env1"),
			ByRows: testlib.IntPtr(3),
			Columns: []*columnDecl{
				{Name: "col1", StartPos: 2, Length: 4, LinePattern: strs.StrPtr("^abc")},
				{Name: "col2", StartPos: 2, Length: 4, LinePattern: strs.StrPtr("^hij")},
				{Name: "col3", StartPos: 3, Length: 5, LinePattern: strs.StrPtr("^abc")},
			},
		}}})

	n, err := r.readByRowsEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, `{"col1":"bcde","col2":"ijkl","col3":"cdefg"}`, idr.JSONify2(n))

	n, err = r.readByRowsEnvelope()
	assert.Error(t, err)
	assert.True(t, IsErrInvalidEnvelope(err))
	assert.Equal(t, "input 'test' line 6: incomplete envelope, missing 2 row(s)", err.Error())
	assert.Nil(t, n)
}

var (
	benchReadByRowsEnvelopeInput = strings.Repeat(
		"abcdefghijklmnopqrstuvwxyz\n  \n012345678901234567890123456789\n", 1000)
	benchReadByRowsEnvelopeDecl = &fileDecl{
		Envelopes: []*envelopeDecl{
			{
				Name:   strs.StrPtr("env1"),
				ByRows: testlib.IntPtr(3),
				Columns: []*columnDecl{
					{Name: "col1", StartPos: 2, Length: 10, LinePattern: strs.StrPtr("^abc")},
					{Name: "col2", StartPos: 2, Length: 10, LinePattern: strs.StrPtr("^0123")},
					{Name: "col3", StartPos: 12, Length: 19, LinePattern: strs.StrPtr("^abc")},
				},
			},
		},
	}
)

// BenchmarkReadByRowsEnvelope-8   	     624	   1891740 ns/op	  133140 B/op	    9005 allocs/op
func BenchmarkReadByRowsEnvelope(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := testReader(strings.NewReader(benchReadByRowsEnvelopeInput), benchReadByRowsEnvelopeDecl)
		for {
			n, err := r.readByRowsEnvelope()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.FailNow()
			}
			idr.RemoveAndReleaseTree(n)
		}
	}
}

func TestReadByHeaderFooterEnvelope_EOFBeforeStart(t *testing.T) {
	r := testReader(strings.NewReader(""), &fileDecl{Envelopes: []*envelopeDecl{{Name: strs.StrPtr("env1")}}})
	n, err := r.readByHeaderFooterEnvelope()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestReadByHeaderFooterEnvelope_ReadErrorBeforeStart(t *testing.T) {
	r := testReader(
		testlib.NewMockReadCloser("read error", nil),
		&fileDecl{Envelopes: []*envelopeDecl{{Name: strs.StrPtr("env1")}}})
	n, err := r.readByHeaderFooterEnvelope()
	assert.Error(t, err)
	assert.True(t, IsErrInvalidEnvelope(err))
	assert.Equal(t, `input 'test' line 1: incomplete envelope: read error`, err.Error())
	assert.Nil(t, n)
}

func TestReadByHeaderFooterEnvelope_NoEnvelopeMatch(t *testing.T) {
	r := testReader(
		strings.NewReader("efg"),
		&fileDecl{Envelopes: []*envelopeDecl{{
			Name:           strs.StrPtr("env1"),
			ByHeaderFooter: &byHeaderFooterDecl{Header: "abc", Footer: "abc"},
		}}})
	n, err := r.readByHeaderFooterEnvelope()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestReadByHeaderFooterEnvelope_IncompleteEnvelope(t *testing.T) {
	r := testReader(
		strings.NewReader("abc"),
		&fileDecl{Envelopes: []*envelopeDecl{{
			Name:           strs.StrPtr("env1"),
			ByHeaderFooter: &byHeaderFooterDecl{Header: "abc", Footer: "efg"},
		}}})
	n, err := r.readByHeaderFooterEnvelope()
	assert.Error(t, err)
	assert.True(t, IsErrInvalidEnvelope(err))
	assert.Equal(t, `input 'test' line 2: incomplete envelope: EOF`, err.Error())
	assert.Nil(t, n)
}

func lf(s string) string { return s + "\n" }

func TestReadByHeaderFooterEnvelope_Success1(t *testing.T) {
	r := testReader(
		strings.NewReader(
			lf("begin")+
				lf("header-01")+
				lf("a001-abc")+
				lf("a002-def")+
				lf("a003-ghi")+
				lf("footer")+
				lf("header-02")+
				lf("a001-012")+
				lf("a002-345")+
				lf("a003-678")+
				lf("footer")+
				lf("end")),
		&fileDecl{Envelopes: []*envelopeDecl{
			{
				Name:           strs.StrPtr("begin"),
				ByHeaderFooter: &byHeaderFooterDecl{Header: "^begin", Footer: "^begin"},
			},
			{
				Name:           strs.StrPtr("data"),
				ByHeaderFooter: &byHeaderFooterDecl{Header: "^header", Footer: "^footer"},
				Columns: []*columnDecl{
					{Name: "data_id", StartPos: 8, Length: 2, LinePattern: strs.StrPtr("^header")},
					{Name: "a001_first2chars", StartPos: 6, Length: 2, LinePattern: strs.StrPtr("^a001")},
					{Name: "a003_last2chars", StartPos: 7, Length: 2, LinePattern: strs.StrPtr("^a003")},
					{Name: "a001_last1char", StartPos: 8, Length: 1, LinePattern: strs.StrPtr("^a001")},
				},
			},
			{
				Name:           strs.StrPtr("end"),
				ByHeaderFooter: &byHeaderFooterDecl{Header: "^end", Footer: "^end"},
			},
		}})
	n, err := r.readByHeaderFooterEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, "begin", n.Data)

	n, err = r.readByHeaderFooterEnvelope()
	assert.NoError(t, err)
	assert.Equal(t,
		`{"a001_first2chars":"ab","a001_last1char":"c","a003_last2chars":"hi","data_id":"01"}`, idr.JSONify2(n))

	n, err = r.readByHeaderFooterEnvelope()
	assert.NoError(t, err)
	assert.Equal(t,
		`{"a001_first2chars":"01","a001_last1char":"2","a003_last2chars":"78","data_id":"02"}`, idr.JSONify2(n))

	n, err = r.readByHeaderFooterEnvelope()
	assert.NoError(t, err)
	assert.Equal(t, "end", n.Data)

	n, err = r.readByHeaderFooterEnvelope()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

var (
	benchReadByHeaderFooterEnvelopeInput = lf("begin") +
		strings.Repeat(
			lf("header")+
				lf("a001-abcdefghijklmnopqrstuvwxyz0123456789")+
				lf("a002-abcdefghijklmnopqrstuvwxyz0123456789")+
				lf("a003-abcdefghijklmnopqrstuvwxyz0123456789")+
				lf("footer"), 1000) +
		lf("end")
	benchReadByHeaderFooterEnvelopeDecl = &fileDecl{
		Envelopes: []*envelopeDecl{
			{
				Name:           strs.StrPtr("begin"),
				ByHeaderFooter: &byHeaderFooterDecl{Header: "^begin", Footer: "^begin"},
			},
			{
				Name:           strs.StrPtr("data"),
				ByHeaderFooter: &byHeaderFooterDecl{Header: "^header", Footer: "^footer"},
				Columns: []*columnDecl{
					{Name: "a001_1", StartPos: 6, Length: 12, LinePattern: strs.StrPtr("^a001")},
					{Name: "a003_1", StartPos: 7, Length: 9, LinePattern: strs.StrPtr("^a003")},
					{Name: "a001_2", StartPos: 30, Length: 20, LinePattern: strs.StrPtr("^a001")},
				},
			},
			{
				Name:           strs.StrPtr("end"),
				ByHeaderFooter: &byHeaderFooterDecl{Header: "^end", Footer: "^end"},
			},
		},
	}
)

// BenchmarkReadByHeaderFooterEnvelope-8   	     310	   3819649 ns/op	  213840 B/op	   14009 allocs/op
func BenchmarkReadByHeaderFooterEnvelope(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := testReader(strings.NewReader(benchReadByHeaderFooterEnvelopeInput), benchReadByHeaderFooterEnvelopeDecl)
		for {
			n, err := r.readByHeaderFooterEnvelope()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.FailNow()
			}
			idr.RemoveAndReleaseTree(n)
		}
	}
}

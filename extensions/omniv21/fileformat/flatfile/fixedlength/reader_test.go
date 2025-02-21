package fixedlength

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/idr"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	format := NewFixedLengthFileFormat("test-schema")
	rt, err := format.ValidateSchema(
		fileFormatFixedLength,
		[]byte(`
			{
				"file_declaration": {
					"envelopes" : [
						{
							"name": "e1", "type": "envelope_group", "is_target": true,
							"child_envelopes": [
								{
									"name": "e2", "header": "^e2h", "footer": "^e2t", "min": 1, "max": 1,
									"columns": [
										{ "name": "c2", "start_pos": 5, "length": 1, "line_pattern": "^e2b" }
									]
								},
								{
									"name": "e3", "rows": 2, "min": 1, "max": 2,
									"columns": [
										{ "name": "c31", "start_pos": 1, "length": 1, "line_index": 1 },
										{ "name": "c32", "start_pos": 1, "length": 1, "line_index": 2 }
									]
								},
								{
									"name": "e4", "header": "^e4h", "min": 1, "max": 1,
									"columns": [
										{ "name": "c4", "start_pos": 5, "length": 1 }
									]
								}
							]
						}
					]
				}
			}
		`),
		&transform.Decl{})
	assert.NoError(t, err)
	for _, test := range []struct {
		name string
		r    io.Reader
		err  string
	}{
		{
			name: "empty input",
			r:    strings.NewReader(""),
			err:  io.EOF.Error(),
		},
		{
			name: "e4 min occurs not satisfied",
			r:    strings.NewReader("e2h\ne2b:1\ne2t\n1-line\n2-line\n3-line\n"),
			err:  "input 'test-input' line 6: envelope/envelope_group 'e1/e4' needs min occur 1, but only got 0",
		},
		{
			name: "unexpected data",
			r:    strings.NewReader("1-line\n"),
			err:  "input 'test-input' line 1: unexpected data",
		},
		{
			name: "success",
			r:    strings.NewReader("e2h\ne2b:2\ne2t\n1-line\n2-line\n3-line\n4-line\ne4h:4\n"),
			err:  "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, err := format.CreateFormatReader(header.Header{}, "test-input", test.r, rt)
			assert.NoError(t, err)
			n, err := r.Read()
			if strs.IsStrNonBlank(test.err) {
				assert.Nil(t, n)
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
			} else {
				assert.NotNil(t, n)
				assert.NoError(t, err)
				cupaloy.SnapshotT(t, idr.JSONify1(n))
			}
		})
	}
}

func TestMoreUnprocessedData(t *testing.T) {
	for _, test := range []struct {
		name    string
		lines   []string
		r       io.Reader
		expMore bool
		expErr  string
	}{
		{
			name:    "linesBuf not empty",
			lines:   []string{""},
			expMore: true,
			expErr:  "",
		},
		{
			name:    "linesBuf empty, io error",
			r:       testlib.NewMockReadCloser("io error", nil),
			expMore: false,
			expErr:  "input 'test-input' line 1: io error",
		},
		{
			name:    "linesBuf empty, io.EOF",
			r:       strings.NewReader(""),
			expMore: false,
			expErr:  "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &reader{
				inputName: "test-input",
			}
			if len(test.lines) > 0 {
				r.linesRead = len(test.lines)
				r.linesBuf = make([]line, len(test.lines))
				for i := range test.lines {
					r.linesBuf[i] = line{lineNum: i + 1, b: []byte(test.lines[i])}
				}
			}
			if test.r != nil {
				r.r = bufio.NewReader(test.r)
			}
			more, err := r.MoreUnprocessedData()
			assert.Equal(t, test.expMore, more)
			if strs.IsStrNonBlank(test.expErr) {
				assert.Error(t, err)
				assert.Equal(t, test.expErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadAndMatchRowsBasedEnvelope(t *testing.T) {
	for _, test := range []struct {
		name           string
		linesBuf       []string
		r              io.Reader
		decl           *EnvelopeDecl
		createIDR      bool
		expMatch       bool
		expIDR         bool
		expErr         string
		expLinesRemain int
	}{
		{
			name:           "non-empty buf, io read err",
			linesBuf:       []string{"line 1"},
			r:              testlib.NewMockReadCloser("io error", nil),
			decl:           &EnvelopeDecl{Rows: testlib.IntPtr(2)},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "input 'test-input' line 2: io error",
			expLinesRemain: 1,
		},
		{
			name:           "empty buf, io.EOF",
			linesBuf:       nil,
			r:              strings.NewReader(""),
			decl:           &EnvelopeDecl{Rows: testlib.IntPtr(2)},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         io.EOF.Error(),
			expLinesRemain: 0,
		},
		{
			name:           "non-empty buf, io.EOF",
			linesBuf:       []string{"line 1"},
			r:              strings.NewReader(""),
			decl:           &EnvelopeDecl{Rows: testlib.IntPtr(2)},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 1,
		},
		{
			name:     "non-empty buf, no read, match, create IDR",
			linesBuf: []string{"line 1", "line 2", "line 3"},
			r:        strings.NewReader(""),
			decl: &EnvelopeDecl{
				Name: "E",
				Rows: testlib.IntPtr(2),
				Columns: []*ColumnDecl{
					{Name: "C1", StartPos: 6, Length: 1, LineIndex: testlib.IntPtr(1)},
					{Name: "C2", StartPos: 6, Length: 1, LineIndex: testlib.IntPtr(2)},
				},
			},
			createIDR:      true,
			expMatch:       true,
			expIDR:         true,
			expErr:         "",
			expLinesRemain: 1,
		},
		{
			name:     "empty buf, read, match, no IDR",
			linesBuf: nil,
			r:        strings.NewReader("line 1\n"),
			decl:     &EnvelopeDecl{
				// Rows == nil, use default value 1
			},
			createIDR:      false,
			expMatch:       true,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &reader{
				inputName: "test-input",
				linesRead: len(test.linesBuf),
				r:         bufio.NewReader(test.r),
			}
			r.linesBuf = make([]line, len(test.linesBuf))
			for i := range test.linesBuf {
				r.linesBuf[i] = line{lineNum: i + 1, b: []byte(test.linesBuf[i])}
			}
			matched, node, err := r.readAndMatchRowsBasedEnvelope(test.decl, test.createIDR)
			assert.Equal(t, test.expMatch, matched)
			if test.expIDR {
				assert.NotNil(t, node)
				cupaloy.SnapshotT(t, idr.JSONify1(node))
			} else {
				assert.Nil(t, node)
			}
			if strs.IsStrNonBlank(test.expErr) {
				assert.Error(t, err)
				assert.Equal(t, test.expErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expLinesRemain, len(r.linesBuf))
		})
	}
}

func TestReadAndMatchHeaderFooterBasedEnvelope(t *testing.T) {
	for _, test := range []struct {
		name           string
		linesBuf       []string
		r              io.Reader
		decl           *EnvelopeDecl
		createIDR      bool
		expMatch       bool
		expIDR         bool
		expErr         string
		expLinesRemain int
	}{
		{
			name:           "empty buf, io read err",
			linesBuf:       nil,
			r:              testlib.NewMockReadCloser("io error", nil),
			decl:           nil,
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "input 'test-input' line 1: io error",
			expLinesRemain: 0,
		},
		{
			name:           "non-empty buf, header not match",
			linesBuf:       []string{"line 1"},
			r:              testlib.NewMockReadCloser("io error", nil), // shouldn't be called
			decl:           &EnvelopeDecl{headerRegexp: regexp.MustCompile("^header")},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 1,
		},
		{
			name:     "empty buf, single line header/footer match",
			linesBuf: nil,
			r:        strings.NewReader("line 1"),
			decl: &EnvelopeDecl{
				Name:         "E",
				Columns:      []*ColumnDecl{{Name: "C", StartPos: 6, Length: 1}},
				headerRegexp: regexp.MustCompile("^line"),
			},
			createIDR:      true,
			expMatch:       true,
			expIDR:         true,
			expErr:         "",
			expLinesRemain: 0,
		},
		{
			name:     "1-line buf, 3-line header/footer match",
			linesBuf: []string{"ABCEFG"},
			r:        strings.NewReader("hello\n123456\n"),
			decl: &EnvelopeDecl{
				Name: "E",
				Columns: []*ColumnDecl{{
					Name:              "C",
					StartPos:          4,
					Length:            2,
					linePatternRegexp: regexp.MustCompile("^he"),
				}},
				headerRegexp: regexp.MustCompile("^ABC"),
				footerRegexp: regexp.MustCompile("^123"),
			},
			createIDR:      false,
			expMatch:       true,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 3, // it's a match, but no idr created, so the lines remain cached
		},
		{
			name:     "1-line buf, header matches, footer line read io error",
			linesBuf: []string{"ABCEFG"},
			r:        testlib.NewMockReadCloser("io error", nil),
			decl: &EnvelopeDecl{
				headerRegexp: regexp.MustCompile("^ABC"),
				footerRegexp: regexp.MustCompile("^123"),
			},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "input 'test-input' line 2: io error",
			expLinesRemain: 1,
		},
		{
			name:     "1-line buf, header matches, footer line read io.EOF",
			linesBuf: []string{"ABCEFG"},
			r:        strings.NewReader(""),
			decl: &EnvelopeDecl{
				headerRegexp: regexp.MustCompile("^ABC"),
				footerRegexp: regexp.MustCompile("^123"),
			},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &reader{
				inputName: "test-input",
				linesRead: len(test.linesBuf),
				r:         bufio.NewReader(test.r),
			}
			r.linesBuf = make([]line, len(test.linesBuf))
			for i := range test.linesBuf {
				r.linesBuf[i] = line{lineNum: i + 1, b: []byte(test.linesBuf[i])}
			}
			matched, node, err := r.readAndMatchHeaderFooterBasedEnvelope(test.decl, test.createIDR)
			assert.Equal(t, test.expMatch, matched)
			if test.expIDR {
				assert.NotNil(t, node)
				cupaloy.SnapshotT(t, idr.JSONify1(node))
			} else {
				assert.Nil(t, node)
			}
			if strs.IsStrNonBlank(test.expErr) {
				assert.Error(t, err)
				assert.Equal(t, test.expErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expLinesRemain, len(r.linesBuf))
		})
	}
}

func TestReadLine(t *testing.T) {
	r := &reader{
		inputName: "test-input",
		linesRead: 42,
		linesBuf: []line{
			{lineNum: 41, b: []byte("line 41")},
			{lineNum: 42, b: []byte("line 42")},
		},
	}
	r.r = bufio.NewReader(strings.NewReader(""))
	err := r.readLine()
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 42, r.linesRead)
	assert.Equal(t, 2, len(r.linesBuf))
	r.r = bufio.NewReader(testlib.NewMockReadCloser("test error", nil))
	err = r.readLine()
	assert.Error(t, err)
	assert.Equal(t, "input 'test-input' line 43: test error", err.Error())
	assert.Equal(t, 42, r.linesRead)
	assert.Equal(t, 2, len(r.linesBuf))
	r.r = bufio.NewReader(strings.NewReader("\n\na new line"))
	err = r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, 45, r.linesRead)
	assert.Equal(t, 3, len(r.linesBuf))
	assert.Equal(t, line{lineNum: 45, b: []byte("a new line")}, r.linesBuf[2])
}

func TestReadLineMultiCalls(t *testing.T) {
	// See more details about the bug: https://github.com/jf-tech/omniparser/issues/213
	r := &reader{inputName: "test-input"}
	// construct a two-line input, and total length of the two line input is longer than
	// the bufio.Reader's internal buffer. Note bufio.Reader has a minReadBufferSize at 16.
	r.r = bufio.NewReaderSize(strings.NewReader("0123456789\nabcdefghijklmnopq\n!@#"), 16)
	// First read should be completely normal and fine.
	err := r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, 1, r.linesRead)
	assert.Equal(t, 1, len(r.linesBuf))
	assert.Equal(t, line{lineNum: 1, b: []byte("0123456789"), copied: false}, r.linesBuf[0])
	// Now the second read, if we didn't do any copying of the first line result, the first line
	// would have become corrupted since the bufio.Reader's tiny internal buffer gets overwritten
	// by the new line data. With the fix, both lines in the linesBuf should be without corruption.
	err = r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, 2, r.linesRead)
	assert.Equal(t, 2, len(r.linesBuf))
	assert.Equal(t, line{lineNum: 1, b: []byte("0123456789"), copied: true}, r.linesBuf[0])
	assert.Equal(t, line{lineNum: 2, b: []byte("abcdefghijklmnopq"), copied: false}, r.linesBuf[1])
	// 3rd line reading
	err = r.readLine()
	assert.NoError(t, err)
	assert.Equal(t, 3, r.linesRead)
	assert.Equal(t, 3, len(r.linesBuf))
	assert.Equal(t, line{lineNum: 1, b: []byte("0123456789"), copied: true}, r.linesBuf[0])
	assert.Equal(t, line{lineNum: 2, b: []byte("abcdefghijklmnopq"), copied: true}, r.linesBuf[1])
	assert.Equal(t, line{lineNum: 3, b: []byte("!@#"), copied: false}, r.linesBuf[2])
}

func TestLinesToNode(t *testing.T) {
	for _, test := range []struct {
		name     string
		lines    [][]byte
		n        int
		cols     []*ColumnDecl
		panicStr string
	}{
		{
			name:     "n > len(lines)",
			lines:    [][]byte{{}},
			n:        3,
			panicStr: "linesBuf has 1 lines but requested 3 lines to convert",
		},
		{
			name: "line matches",
			lines: [][]byte{
				[]byte("abc"),
				[]byte("hello, world"),
				[]byte("1234"),
				[]byte("#$%^"),
			},
			n: 3,
			cols: []*ColumnDecl{
				{Name: "W", StartPos: 8, Length: 5, linePatternRegexp: regexp.MustCompile("^hello")},
				{Name: "C", StartPos: 3, Length: 1, LineIndex: testlib.IntPtr(1)},
				{Name: "0", StartPos: 1, Length: 1, linePatternRegexp: regexp.MustCompile("no-match")},
				{Name: "A", StartPos: 1, Length: 1},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &reader{}
			r.linesBuf = make([]line, len(test.lines))
			for i := range test.lines {
				r.linesBuf[i] = line{lineNum: i, b: test.lines[i]}
			}
			if strs.IsStrNonBlank(test.panicStr) {
				assert.PanicsWithValue(t, test.panicStr, func() {
					r.linesToNode(
						&EnvelopeDecl{Name: "test-envelope", Columns: test.cols},
						test.n)
				})
			} else {
				node := r.linesToNode(
					&EnvelopeDecl{Name: "test-envelope", Columns: test.cols},
					test.n)
				cupaloy.SnapshotT(t, idr.JSONify1(node))
			}
		})
	}
}

func TestPopFrontLinesBuf(t *testing.T) {
	r := &reader{}
	r.linesBuf = make([]line, 3, 10)
	r.linesBuf[0] = line{lineNum: 10, b: []byte("a")}
	r.linesBuf[1] = line{lineNum: 11, b: []byte("b")}
	r.linesBuf[2] = line{lineNum: 12, b: []byte("c")}
	assert.PanicsWithValue(t,
		"less lines (3) in r.linesBuf than requested pop front count (5)",
		func() { r.popFrontLinesBuf(5) })
	assert.Equal(t, 3, len(r.linesBuf))
	assert.Equal(t, 10, cap(r.linesBuf))
	r.popFrontLinesBuf(2)
	assert.Equal(t, 1, len(r.linesBuf))
	assert.Equal(t, 10, cap(r.linesBuf))
	assert.Equal(t, line{lineNum: 12, b: []byte("c")}, r.linesBuf[0])
}

func TestUnprocessedLineNum(t *testing.T) {
	r := &reader{linesRead: 42}
	assert.Equal(t, 42+1, r.unprocessedLineNum())
	r.linesBuf = []line{{lineNum: 13}}
	assert.Equal(t, 13, r.unprocessedLineNum())
}

func TestIsContinuableError(t *testing.T) {
	r := &reader{}
	assert.True(t, r.IsContinuableError(r.FmtErr("some error")))
	assert.False(t, r.IsContinuableError(ErrInvalidFixedLength("invalid envelope")))
	assert.False(t, r.IsContinuableError(io.EOF))
}

func TestIsErrInvalidFixedLength(t *testing.T) {
	assert.True(t, IsErrInvalidFixedLength(ErrInvalidFixedLength("test")))
	assert.Equal(t, "test", ErrInvalidFixedLength("test").Error())
	assert.False(t, IsErrInvalidFixedLength(errors.New("test")))
}

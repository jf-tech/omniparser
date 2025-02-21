package csv

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/antchfx/xpath"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/idr"
	"github.com/stretchr/testify/assert"
)

func lf(s string) string {
	return s + "\n"
}

func TestRead(t *testing.T) {
	for _, test := range []struct {
		name        string
		debug       int
		fileDecl    string
		targetXPath string
		input       io.Reader
		expErrs     []string
	}{
		{
			name: "header row min occurs not satisfied",
			fileDecl: `{
						"delimiter": "|",
						"replace_double_quotes": true,
						"records": [
							{ "name": "header", "min": 1, "max": 1 }
						]
					}`,
			input: strings.NewReader(""),
			expErrs: []string{
				"input 'test-input' line 2: record/record_group 'header' needs min occur 1, but only got 0",
			},
		},
		{
			name: "unexpected data at the end",
			fileDecl: `{
						"delimiter": ",",
						"records": [
							{ "name": "r1", "min": 1, "max": 1 }
						]
					}`,
			input: strings.NewReader("row1\nrow2"),
			expErrs: []string{
				"", // first Read() is a success.
				"input 'test-input' line 2: unexpected data",
			},
		},
		{
			name: "IO failure",
			fileDecl: `{
						"delimiter": ",",
						"records": [
							{ "name": "r1", "min": 1, "max": 1 }
						]
					}`,
			input: testlib.NewMockReadCloser("test failure", nil),
			expErrs: []string{
				"input 'test-input' line 1: test failure",
			},
		},
		{
			name:  "multiple records",
			debug: 1,
			fileDecl: `{
				"delimiter": ",",
				"records": [
					{ "name": "r0", "header": "^begin0" },
					{ "name": "g1", "type": "record_group", "min": 1, "max": 1,
						"child_records": [
							{ "name": "r1", "rows": 2, "min": 2, "max": 2, "is_target": true,
								"columns": [
									{ "name": "r1c1", "index": 2, "line_index": 1 },
									{ "name": "r1c2", "index": 3, "line_pattern": "^r1-row2" }
								]
							}
						]
					},
					{ "name": "r2", "header": "^begin2", "footer": "^end2", "min": 0 },
					{ "name": "g2", "type": "record_group", "min": 1, "max": 1,
						"child_records": [
							{ "name": "r3", "header": "^begin3", "footer": "^end3", "min": 1, "max": 1 }
						]
					},
					{ "name": "r4", "rows": 2, "min": 1, "max": 1 },
					{ "name": "r5", "header": "^begin5", "footer": "^end5", "min": 1, "max": 1 }
				]
			}`,
			input: strings.NewReader(
				lf("r1-row1,v1,v2") +
					lf("r1-row2,v3,v4") +
					lf("r2-row1") +
					lf("r2-row2") +
					lf("begin3") +
					lf("...") +
					lf("end3") +
					lf("r4-row1") +
					lf("r4-row2") +
					lf("begin5")),
			expErrs: []string{
				"",
				"",
				"input 'test-input' line 10: record/record_group 'r5' needs min occur 1, but only got 0",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var fd FileDecl
			assert.NoError(t, json.Unmarshal([]byte(test.fileDecl), &fd))
			err := (&validateCtx{}).validateFileDecl(&fd)
			assert.NoError(t, err)
			var targetXPathExpr *xpath.Expr
			if test.targetXPath != "" {
				targetXPathExpr, err = caches.GetXPathExpr(test.targetXPath)
				assert.NoError(t, err)
			}
			r := NewReader(
				header.Header{ParserSettings: header.ParserSettings{Debug: test.debug}},
				"test-input", test.input, &fd, targetXPathExpr)
			var nodes []string
			for _, expErr := range test.expErrs {
				node, err := r.Read()
				if expErr != "" {
					assert.Error(t, err)
					assert.Equal(t, expErr, err.Error())
					assert.Nil(t, node)
					break
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, node)
					nodes = append(nodes, idr.JSONify1(node))
				}
				r.Release(node)
			}
			if len(nodes) > 0 {
				cupaloy.SnapshotT(t, strings.Join(nodes, ",\n"))
			}
		})
	}
}

func TestReadAndMatchRowsBasedRecord(t *testing.T) {
	for _, test := range []struct {
		name           string
		linesBuf       []line
		records        []string
		r              io.Reader
		decl           *RecordDecl
		createIDR      bool
		expMatch       bool
		expIDR         bool
		expErr         string
		expLinesRemain int
	}{
		{
			name:           "non-empty buf, io read err",
			linesBuf:       []line{{recordStart: 0, recordNum: 1}},
			records:        []string{"line 1"},
			r:              testlib.NewMockReadCloser("io error", nil),
			decl:           &RecordDecl{Rows: testlib.IntPtr(2)},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "input 'test-input' line 1: io error",
			expLinesRemain: 1,
		},
		{
			name:           "empty buf, io.EOF",
			linesBuf:       nil,
			records:        nil,
			r:              strings.NewReader(""),
			decl:           &RecordDecl{Rows: testlib.IntPtr(2)},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         io.EOF.Error(),
			expLinesRemain: 0,
		},
		{
			name:           "non-empty buf, io.EOF",
			linesBuf:       []line{{recordStart: 0, recordNum: 1}},
			records:        []string{"line 1"},
			r:              strings.NewReader(""),
			decl:           &RecordDecl{Rows: testlib.IntPtr(2)},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 1,
		},
		{
			name: "non-empty buf, no read, match, create IDR",
			linesBuf: []line{
				{recordStart: 0, recordNum: 2},
				{recordStart: 2, recordNum: 2},
				{recordStart: 4, recordNum: 2},
			},
			records: []string{
				"line 1", "one",
				"line 2", "two",
				"line 3", "three",
			},
			r: testlib.NewMockReadCloser("io error", nil), // shouldn't be called
			decl: &RecordDecl{
				Name: "E",
				Rows: testlib.IntPtr(2),
				Columns: []*ColumnDecl{
					{Name: "C1", Index: testlib.IntPtr(2), LineIndex: testlib.IntPtr(1)},
					{Name: "C2", Index: testlib.IntPtr(2), LineIndex: testlib.IntPtr(2)},
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
			records:  nil,
			r:        strings.NewReader("line 1\n"),
			decl:     &RecordDecl{
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
				fileDecl:  &FileDecl{Delimiter: ","},
				r:         ios.NewLineNumReportingCsvReader(test.r),
				linesBuf:  test.linesBuf,
				records:   test.records,
			}
			matched, node, err := r.readAndMatchRowsBasedRecord(test.decl, test.createIDR)
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

func TestReadAndMatchHeaderFooterBasedRecord(t *testing.T) {
	for _, test := range []struct {
		name           string
		linesBuf       []line
		records        []string
		r              io.Reader
		decl           *RecordDecl
		createIDR      bool
		expMatch       bool
		expIDR         bool
		expErr         string
		expLinesRemain int
	}{
		{
			name:           "empty buf, io read err",
			linesBuf:       nil,
			records:        nil,
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
			linesBuf:       []line{{recordStart: 0, recordNum: 1}},
			records:        []string{"line 1"},
			r:              testlib.NewMockReadCloser("io error", nil), // shouldn't be called
			decl:           &RecordDecl{headerRegexp: regexp.MustCompile("^no matched")},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "",
			expLinesRemain: 1,
		},
		{
			name:     "empty buf, single line header/footer match",
			linesBuf: nil,
			records:  nil,
			r:        strings.NewReader("line 1"),
			decl: &RecordDecl{
				Name:         "E",
				Columns:      []*ColumnDecl{{Name: "C", Index: testlib.IntPtr(1)}},
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
			linesBuf: []line{{recordStart: 0, recordNum: 1}},
			records:  []string{"ABCEFG"},
			r:        strings.NewReader("hello\n123456\n"),
			decl: &RecordDecl{
				Name: "E",
				Columns: []*ColumnDecl{{
					Name:              "C",
					Index:             testlib.IntPtr(1),
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
			linesBuf: []line{{recordStart: 0, recordNum: 1}},
			records:  []string{"ABCEFG"},
			r:        testlib.NewMockReadCloser("io error", nil),
			decl: &RecordDecl{
				headerRegexp: regexp.MustCompile("^ABC"),
				footerRegexp: regexp.MustCompile("^123"),
			},
			createIDR:      false,
			expMatch:       false,
			expIDR:         false,
			expErr:         "input 'test-input' line 1: io error",
			expLinesRemain: 1,
		},
		{
			name:     "1-line buf, header matches, footer line read io.EOF",
			linesBuf: []line{{recordStart: 0, recordNum: 1}},
			records:  []string{"ABCEFG"},
			r:        strings.NewReader(""),
			decl: &RecordDecl{
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
				fileDecl:  &FileDecl{Delimiter: ","},
				r:         ios.NewLineNumReportingCsvReader(test.r),
				linesBuf:  test.linesBuf,
				records:   test.records,
			}
			matched, node, err := r.readAndMatchHeaderFooterBasedRecord(test.decl, test.createIDR)
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

func TestLinesToNode(t *testing.T) {
	r := &reader{fileDecl: &FileDecl{Delimiter: ","}}
	assert.PanicsWithValue(t,
		"linesBuf has 0 lines but requested 3 lines to convert",
		func() { r.linesToNode(&RecordDecl{}, 3) })
	r.linesBuf = []line{
		{lineNum: 1, recordStart: 0, recordNum: 3},
		{lineNum: 2, recordStart: 3, recordNum: 2},
		{lineNum: 3, recordStart: 5, recordNum: 4},
	}
	r.records = []string{
		"1", "2", "3",
		"a", "b",
		"#$%^&", "*()", ":?>", "~!@",
	}
	cupaloy.SnapshotT(t, idr.JSONify1(r.linesToNode(
		&RecordDecl{
			Name: "test",
			Columns: []*ColumnDecl{
				// c1 gets line 3 3rd record: "~!@"
				{Name: "c1", Index: testlib.IntPtr(3), LineIndex: testlib.IntPtr(3)},
				// c2 gets line 2 1st record: "a"
				{Name: "c2", Index: testlib.IntPtr(1), linePatternRegexp: regexp.MustCompile("^a,b")},
				// c3 gets line 1 (no line_index/line_pattern means always match, thus the first
				// line matches) 2nd record: "2"
				{Name: "c3", Index: testlib.IntPtr(2)},
			},
		},
		3)))
}

func TestPopFrontLinesBuf(t *testing.T) {
	r := &reader{}
	r.linesBuf = make([]line, 0, 10)
	r.linesBuf = append(r.linesBuf, []line{
		{lineNum: 10, recordStart: 0, recordNum: 3},
		{lineNum: 11, recordStart: 3, recordNum: 2},
		{lineNum: 12, recordStart: 5, recordNum: 4},
	}...)
	r.records = []string{
		"a", "b", "c",
		"1", "2",
		"#", "$", "%", "^",
	}
	assert.PanicsWithValue(t,
		"less lines (3) in r.linesBuf than requested pop front count (5)",
		func() { r.popFrontLinesBuf(5) })
	assert.Equal(t, 3, len(r.linesBuf))
	assert.Equal(t, 10, cap(r.linesBuf))
	r.popFrontLinesBuf(2)
	assert.Equal(t, 1, len(r.linesBuf))
	assert.Equal(t, 10, cap(r.linesBuf))
	assert.Equal(t, line{lineNum: 12, recordStart: 0, recordNum: 4}, r.linesBuf[0])
	assert.Equal(t, 4, len(r.records))
	assert.Equal(t, 9, cap(r.records))
	assert.Equal(t, []string{"#", "$", "%", "^"}, r.records)
}

func TestIsContinuableError(t *testing.T) {
	r := &reader{
		r: ios.NewLineNumReportingCsvReader(strings.NewReader("test")),
	}
	assert.True(t, r.IsContinuableError(r.FmtErr("some error")))
	assert.False(t, r.IsContinuableError(ErrInvalidCSV("invalid record")))
	assert.False(t, r.IsContinuableError(io.EOF))
}

func TestIsErrInvalidCSV(t *testing.T) {
	assert.True(t, IsErrInvalidCSV(ErrInvalidCSV("test")))
	assert.Equal(t, "test", ErrInvalidCSV("test").Error())
	assert.False(t, IsErrInvalidCSV(errors.New("test")))
}

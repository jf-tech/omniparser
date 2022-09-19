package csv

import (
	"regexp"
	"testing"

	"github.com/jf-tech/go-corelib/maths"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile"
	"github.com/stretchr/testify/assert"
)

func TestColumnDecl_LineMatch(t *testing.T) {
	// no line_index/line_pattern, always match
	assert.True(t, (&ColumnDecl{}).lineMatch(0, &line{}, nil, ""))
	assert.False(t, (&ColumnDecl{LineIndex: testlib.IntPtr(2)}).lineMatch(0, &line{}, nil, ""))
	assert.True(t, (&ColumnDecl{LineIndex: testlib.IntPtr(2)}).lineMatch(1, &line{}, nil, ""))
	assert.False(t, (&ColumnDecl{linePatternRegexp: regexp.MustCompile("^ABC.*$")}).
		lineMatch(0, &line{recordStart: 0, recordNum: 2}, []string{"123", "456"}, ","))
	l := &line{recordStart: 0, recordNum: 2}
	assert.True(t,
		(&ColumnDecl{linePatternRegexp: regexp.MustCompile("^ABC\\|D.*$")}).lineMatch(
			0, l, []string{"ABC", "DEF"}, "|"))
	assert.Equal(t, "ABC|DEF", l.raw)
}

func TestColumnDecl_LineToColumnValue(t *testing.T) {
	assert.Equal(t, "", (&ColumnDecl{Index: testlib.IntPtr(2)}).lineToColumnValue(
		&line{recordNum: 1}, nil)) // index out of range
	assert.Equal(t, "", (&ColumnDecl{Index: testlib.IntPtr(0)}).lineToColumnValue(
		&line{recordNum: 1}, nil)) // index out of range
	assert.Equal(t, "6", (&ColumnDecl{Index: testlib.IntPtr(5)}).lineToColumnValue(
		&line{recordStart: 1, recordNum: 7},                    // "2" .. "8"
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"})) // in range
}

func TestRecordDecl(t *testing.T) {
	// DeclName()
	r := &RecordDecl{Name: "r1"}
	assert.Equal(t, "r1", r.DeclName())
	r.fqdn = r.DeclName()

	// Target()
	assert.False(t, r.Target())
	r.IsTarget = true
	assert.True(t, r.Target())

	// Group()
	assert.False(t, r.Group())
	r.Type = strs.StrPtr(typeRecord)
	assert.False(t, r.Group())
	r.Type = strs.StrPtr(typeGroup)
	assert.True(t, r.Group())

	// MinOccurs()
	assert.Equal(t, 0, r.MinOccurs())
	r.Min = testlib.IntPtr(42)
	assert.Equal(t, 42, r.MinOccurs())

	// MaxOccurs()
	assert.Equal(t, maths.MaxIntValue, r.MaxOccurs())
	r.Max = testlib.IntPtr(-1)
	assert.Equal(t, maths.MaxIntValue, r.MaxOccurs())
	r.Max = testlib.IntPtr(42)
	assert.Equal(t, 42, r.MaxOccurs())

	// ChildDecls()
	assert.Nil(t, r.ChildDecls())
	r.childRecDecls = []flatfile.RecDecl{}
	assert.Equal(t, r.childRecDecls, r.ChildDecls())

	// rowsBased()
	assert.PanicsWithValue(t, "record_group is neither rows based nor header/footer based",
		func() { r.rowsBased() })
	r.Type = strs.StrPtr(typeRecord)
	assert.True(t, r.rowsBased())
	r.Header = strs.StrPtr("^ABC$")
	assert.False(t, r.rowsBased())

	// rows()
	assert.PanicsWithValue(t, "record 'r1' is not rows based", func() { r.rows() })
	r.Header = nil
	assert.Equal(t, 1, r.rows())
	r.Rows = testlib.IntPtr(42)
	assert.Equal(t, 42, r.rows())

	// matchHeader()
	assert.PanicsWithValue(
		t, "record 'r1' is not header/footer based", func() { r.matchHeader(&line{}, nil, "") })
	r.headerRegexp = regexp.MustCompile("^ABC,")
	assert.False(t, r.matchHeader(&line{}, nil, ","))
	line := &line{recordStart: 1, recordNum: 2}
	assert.True(t, r.matchHeader(line, []string{"123", "ABC", "EFG"}, ","))

	// matchFooter()
	assert.True(t, r.matchFooter(line, nil, ",")) // if `footer` not specified, always match.
	r.footerRegexp = regexp.MustCompile("^ABC,EF.*$")
	assert.True(t, r.matchFooter(line, []string{"123", "ABC", "EFG"}, ","))
}

func TestMatchLine(t *testing.T) {
	records := []string{"0", "1", "2", "3"}
	line := &line{recordStart: 1, recordNum: 2} // "1", "2"
	assert.Equal(t, "", line.raw)
	assert.False(t, matchLine(regexp.MustCompile("^1\\|2$"), line, records, ","))
	assert.Equal(t, "1,2", line.raw)
	assert.True(t, matchLine(regexp.MustCompile("^1,2$"), line, records, ","))
}

func TestToFlatFileRecDecls(t *testing.T) {
	assert.Nil(t, toFlatFileRecDecls(nil))
	assert.Nil(t, toFlatFileRecDecls([]*RecordDecl{}))
	rs := []*RecordDecl{
		{},
		{},
	}
	ds := toFlatFileRecDecls(rs)
	for i := range ds {
		assert.Same(t, rs[i], ds[i].(*RecordDecl))
	}
}

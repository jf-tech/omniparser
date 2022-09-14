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
	assert.True(t, (&ColumnDecl{}).lineMatch(0, line{}))
	assert.False(t, (&ColumnDecl{LineIndex: testlib.IntPtr(2)}).lineMatch(0, line{}))
	assert.True(t, (&ColumnDecl{LineIndex: testlib.IntPtr(2)}).lineMatch(1, line{}))
	assert.False(t, (&ColumnDecl{linePatternRegexp: regexp.MustCompile("^ABC.*$")}).
		lineMatch(0, line{raw: "1234567"}))
	assert.True(t, (&ColumnDecl{linePatternRegexp: regexp.MustCompile("^ABC.*$")}).
		lineMatch(0, line{raw: "ABCDEFG"}))
}

func TestColumnDecl_LineToColumnValue(t *testing.T) {
	assert.Equal(t, "", (&ColumnDecl{Index: testlib.IntPtr(2)}).lineToColumnValue(
		line{record: []string{"1"}})) // index out of range
	assert.Equal(t, "", (&ColumnDecl{Index: testlib.IntPtr(0)}).lineToColumnValue(
		line{record: []string{"1"}})) // index out of range
	assert.Equal(t, "9", (&ColumnDecl{Index: testlib.IntPtr(5)}).lineToColumnValue(
		line{record: []string{"1", "3", "5", "7", "9", "11"}})) // in range
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
		t, "record 'r1' is not header/footer based", func() { r.matchHeader(nil) })
	r.headerRegexp = regexp.MustCompile("^ABC$")
	assert.False(t, r.matchHeader([]byte("ABCD")))
	assert.True(t, r.matchHeader([]byte("ABC")))

	// matchFooter()
	assert.True(t, r.matchFooter([]byte("ABCD")))
	r.footerRegexp = regexp.MustCompile("^ABC$")
	assert.False(t, r.matchFooter([]byte("ABCD")))
	assert.True(t, r.matchFooter([]byte("ABC")))
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

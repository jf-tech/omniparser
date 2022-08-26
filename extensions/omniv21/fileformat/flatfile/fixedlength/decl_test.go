package fixedlength

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
	assert.True(t, (&ColumnDecl{}).lineMatch(0, []byte("test")))
	assert.False(t, (&ColumnDecl{LineIndex: testlib.IntPtr(2)}).lineMatch(0, []byte("test")))
	assert.True(t, (&ColumnDecl{LineIndex: testlib.IntPtr(2)}).lineMatch(1, []byte("test")))
	assert.False(t, (&ColumnDecl{linePatternRegexp: regexp.MustCompile("^ABC.*$")}).
		lineMatch(0, []byte("test")))
	assert.True(t, (&ColumnDecl{linePatternRegexp: regexp.MustCompile("^ABC.*$")}).
		lineMatch(0, []byte("ABCDEFG")))
}

func TestColumnDecl_LineToColumnValue(t *testing.T) {
	decl := func(start, length int) *ColumnDecl {
		return &ColumnDecl{StartPos: start, Length: length}
	}
	assert.Equal(t, "", decl(10, 4).lineToColumnValue([]byte("test")))   // fully out of range
	assert.Equal(t, "st", decl(3, 4).lineToColumnValue([]byte("test")))  // partially out of range
	assert.Equal(t, "tes", decl(1, 3).lineToColumnValue([]byte("test"))) // fully in range
}

func TestEnvelopeDecl(t *testing.T) {
	// DeclName()
	e := &EnvelopeDecl{Name: "e1"}
	assert.Equal(t, "e1", e.DeclName())
	e.fqdn = e.DeclName()

	// Target()
	assert.False(t, e.Target())
	e.IsTarget = true
	assert.True(t, e.Target())

	// Group()
	assert.False(t, e.Group())
	e.Type = strs.StrPtr(typeEnvelope)
	assert.False(t, e.Group())
	e.Type = strs.StrPtr(typeGroup)
	assert.True(t, e.Group())

	// MinOccurs()
	assert.Equal(t, 0, e.MinOccurs())
	e.Min = testlib.IntPtr(42)
	assert.Equal(t, 42, e.MinOccurs())

	// MaxOccurs()
	assert.Equal(t, maths.MaxIntValue, e.MaxOccurs())
	e.Max = testlib.IntPtr(-1)
	assert.Equal(t, maths.MaxIntValue, e.MaxOccurs())
	e.Max = testlib.IntPtr(42)
	assert.Equal(t, 42, e.MaxOccurs())

	// ChildDecls()
	assert.Nil(t, e.ChildDecls())
	e.childRecDecls = []flatfile.RecDecl{}
	assert.Equal(t, e.childRecDecls, e.ChildDecls())

	// rowsBased()
	assert.PanicsWithValue(t, "envelope_group is neither rows based nor header/footer based",
		func() { e.rowsBased() })
	e.Type = strs.StrPtr(typeEnvelope)
	assert.True(t, e.rowsBased())
	e.Header = strs.StrPtr("^ABC$")
	assert.False(t, e.rowsBased())

	// rows()
	assert.PanicsWithValue(t, "envelope 'e1' is not rows based", func() { e.rows() })
	e.Header = nil
	assert.Equal(t, 1, e.rows())
	e.Rows = testlib.IntPtr(42)
	assert.Equal(t, 42, e.rows())

	// matchHeader()
	assert.PanicsWithValue(
		t, "envelope 'e1' is not header/footer based", func() { e.matchHeader(nil) })
	e.headerRegexp = regexp.MustCompile("^ABC$")
	assert.False(t, e.matchHeader([]byte("ABCD")))
	assert.True(t, e.matchHeader([]byte("ABC")))

	// matchFooter()
	assert.True(t, e.matchFooter([]byte("ABCD")))
	e.footerRegexp = regexp.MustCompile("^ABC$")
	assert.False(t, e.matchFooter([]byte("ABCD")))
	assert.True(t, e.matchFooter([]byte("ABC")))
}

func TestToFlatFileRecDecls(t *testing.T) {
	assert.Nil(t, toFlatFileRecDecls(nil))
	assert.Nil(t, toFlatFileRecDecls([]*EnvelopeDecl{}))
	es := []*EnvelopeDecl{
		{},
		{},
	}
	ds := toFlatFileRecDecls(es)
	for i := range ds {
		assert.Same(t, es[i], ds[i].(*EnvelopeDecl))
	}
}

package fixedlength

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestColumnDecl_LineMatch(t *testing.T) {
	assert.True(t, (&columnDecl{}).lineMatch([]byte("test")))
	assert.False(t, (&columnDecl{LinePattern: strs.StrPtr("^ABC.*$")}).lineMatch([]byte("test")))
	assert.True(t, (&columnDecl{LinePattern: strs.StrPtr("^ABC.*$")}).lineMatch([]byte("ABCDEFG")))
}

func TestColumnDecl_LineToColumnValue(t *testing.T) {
	decl := func(start, length int) *columnDecl {
		return &columnDecl{StartPos: start, Length: length}
	}
	assert.Equal(t, "", decl(10, 4).lineToColumnValue([]byte("test")))   // fully out of range
	assert.Equal(t, "st", decl(3, 4).lineToColumnValue([]byte("test")))  // partially out of range
	assert.Equal(t, "tes", decl(1, 3).lineToColumnValue([]byte("test"))) // fully in range
}

func TestEnvelopeDecl_ByRows(t *testing.T) {
	assert.PanicsWithValue(t, `envelope 'env1' type is not 'by_rows'`, func() {
		(&envelopeDecl{Name: strs.StrPtr("env1"), ByHeaderFooter: &byHeaderFooterDecl{}}).byRows()
	})
	assert.Equal(t, 1, (&envelopeDecl{}).byRows())
	assert.Equal(t, 12, (&envelopeDecl{ByRows: testlib.IntPtr(12)}).byRows())
}

func TestFileDecl_EnvelopeType(t *testing.T) {
	assert.Equal(t, envelopeTypeByHeaderFooter,
		(&fileDecl{
			Envelopes: []*envelopeDecl{
				{ByHeaderFooter: &byHeaderFooterDecl{}},
			},
		}).envelopeType())
	assert.Equal(t, envelopeTypeByRows,
		(&fileDecl{
			Envelopes: []*envelopeDecl{
				{ByRows: testlib.IntPtr(12)},
			},
		}).envelopeType())
	assert.Equal(t, envelopeTypeByRows,
		(&fileDecl{
			Envelopes: []*envelopeDecl{
				{},
			},
		}).envelopeType())
}

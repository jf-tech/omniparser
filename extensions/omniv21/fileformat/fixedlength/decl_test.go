package fixedlength

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestColumnDecl_LineMatch(t *testing.T) {
	assert.True(t, (&columnDecl{}).lineMatch([]byte("test")))
	assert.False(t, (&columnDecl{Line: strs.StrPtr("^ABC.*$")}).lineMatch([]byte("test")))
	assert.True(t, (&columnDecl{Line: strs.StrPtr("^ABC.*$")}).lineMatch([]byte("ABCDEFG")))
}

func TestColumnDecl_LineToColumn(t *testing.T) {
	decl := func(start, length int) *columnDecl {
		return &columnDecl{StartPos: start, Length: length}
	}
	assert.Nil(t, decl(10, 4).lineToColumn([]rune("test")))                 // fully out of range
	assert.Equal(t, []rune("st"), decl(3, 4).lineToColumn([]rune("test")))  // partially out of range
	assert.Equal(t, []rune("tes"), decl(1, 3).lineToColumn([]rune("test"))) // fully in range
}

func TestEnvelopeDecl_ByRows(t *testing.T) {
	assert.PanicsWithValue(t, `envelope 'env1' type is not 'by_rows'`, func() {
		(&envelopeDecl{Name: strs.StrPtr("env1"), ByHeaderFooter: &byHeaderFooterDecl{}}).byRows()
	})
	assert.Equal(t, 1, (&envelopeDecl{}).byRows())
	assert.Equal(t, 12, (&envelopeDecl{ByRows: testlib.IntPtr(12)}).byRows())
}

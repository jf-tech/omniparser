package flatfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDecl struct {
	name     string
	target   bool
	group    bool
	min      int
	max      int
	children []testDecl
}

func (d testDecl) DeclName() string      { return d.name }
func (d testDecl) Target() bool          { return d.target }
func (d testDecl) Group() bool           { return d.group }
func (d testDecl) MinOccurs() int        { return d.min }
func (d testDecl) MaxOccurs() int        { return d.max }
func (d testDecl) ChildDecls() []RecDecl { return toDeclSlice(d.children) }

func toDeclSlice(ds []testDecl) []RecDecl {
	if len(ds) <= 0 {
		return nil
	}
	r := make([]RecDecl, len(ds))
	for i, d := range ds {
		r[i] = d
	}
	return r
}

func TestRootDecl(t *testing.T) {
	rd := rootDecl{
		children: toDeclSlice([]testDecl{
			{name: "1"},
			{name: "2"},
		}),
	}
	assert.Equal(t, rootName, rd.DeclName())
	assert.False(t, rd.Target())
	assert.True(t, rd.Group())
	assert.Equal(t, 1, rd.MinOccurs())
	assert.Equal(t, 1, rd.MaxOccurs())
	assert.Equal(t, 2, len(rd.ChildDecls()))
	assert.Equal(t, "1", rd.ChildDecls()[0].DeclName())
	assert.Equal(t, "2", rd.ChildDecls()[1].DeclName())
}

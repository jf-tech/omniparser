package edi

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestValidateFileDecl_Empty(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&FileDecl{})
	assert.Error(t, err)
	assert.Equal(t, `missing segment/segment_group with 'is_target' = true`, err.Error())
}

func TestValidateFileDecl_MinGreaterThanMax(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&FileDecl{
		SegDecls: []*SegDecl{
			{Name: "A", Children: []*SegDecl{{Name: "B", Max: testlib.IntPtr(0)}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `segment 'A/B' has 'min' value 1 > 'max' value 0`, err.Error())
}

func TestValidateFileDecl_TwoIsTarget(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&FileDecl{
		SegDecls: []*SegDecl{
			{Name: "A", IsTarget: true},
			{Name: "B", Type: strs.StrPtr(segTypeGroup), IsTarget: true},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `a second segment/group ('B') with 'is_target' = true is not allowed`, err.Error())
}

func TestValidateFileDecl_SegGroupHasNoChildren(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&FileDecl{
		SegDecls: []*SegDecl{
			{Name: "A", Type: strs.StrPtr(segTypeGroup), IsTarget: true},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `segment_group 'A' must have at least one child segment/segment_group`, err.Error())
}

func TestValidateFileDecl_Success(t *testing.T) {
	elem1 := Elem{Name: "be1", Index: 1}
	elem2 := Elem{Name: "be2c1", Index: 2, CompIndex: testlib.IntPtr(1)}
	elem3 := Elem{Name: "be2c2", Index: 2, CompIndex: testlib.IntPtr(2)}
	fd := &FileDecl{
		SegDecls: []*SegDecl{
			{Name: "A", Children: []*SegDecl{
				{Name: "B", IsTarget: true, Elems: []Elem{elem3, elem1, elem2}},
			}},
		},
	}
	err := (&ediValidateCtx{}).validateFileDecl(fd)
	assert.NoError(t, err)
	assert.Equal(t, "A", fd.SegDecls[0].fqdn)
	assert.Nil(t, fd.SegDecls[0].Elems)
	assert.Equal(t, "A/B", fd.SegDecls[0].Children[0].fqdn)
	assert.Equal(t, []Elem{elem3, elem1, elem2}, fd.SegDecls[0].Children[0].Elems)
}

package edi

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestValidateFileDecl_Empty(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&fileDecl{})
	assert.Error(t, err)
	assert.Equal(t, `missing segment/segment_group with 'is_target' = true`, err.Error())
}

func TestValidateFileDecl_MinGreaterThanMax(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&fileDecl{
		SegDecls: []*segDecl{
			{Name: "A", Children: []*segDecl{{Name: "B", Max: testlib.IntPtr(0)}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `segment 'A/B' has 'min' value 1 > 'max' value 0`, err.Error())
}

func TestValidateFileDecl_TwoIsTarget(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&fileDecl{
		SegDecls: []*segDecl{
			{Name: "A", IsTarget: true},
			{Name: "B", Type: strs.StrPtr(segTypeGroup), IsTarget: true},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `a second segment/group ('B') with 'is_target' = true is not allowed`, err.Error())
}

func TestValidateFileDecl_SegGroupHasNoChildren(t *testing.T) {
	err := (&ediValidateCtx{}).validateFileDecl(&fileDecl{
		SegDecls: []*segDecl{
			{Name: "A", Type: strs.StrPtr(segTypeGroup), IsTarget: true},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `segment_group 'A' must have at least one child segment/segment_group`, err.Error())
}

func TestValidateFileDecl_Success(t *testing.T) {
	elem1 := elem{Name: "be1", Index: 1}
	elem2 := elem{Name: "be2c1", Index: 2, CompIndex: testlib.IntPtr(1)}
	elem3 := elem{Name: "be2c2", Index: 2, CompIndex: testlib.IntPtr(2)}
	fd := &fileDecl{
		SegDecls: []*segDecl{
			{Name: "A", Children: []*segDecl{
				{Name: "B", IsTarget: true, Elems: []elem{elem3, elem1, elem2}},
			}},
		},
	}
	err := (&ediValidateCtx{}).validateFileDecl(fd)
	assert.NoError(t, err)
	assert.Equal(t, "A", fd.SegDecls[0].fqdn)
	assert.Nil(t, fd.SegDecls[0].Elems)
	assert.Equal(t, "A/B", fd.SegDecls[0].Children[0].fqdn)
	assert.Equal(t, []elem{elem3, elem1, elem2}, fd.SegDecls[0].Children[0].Elems)
}

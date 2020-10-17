package edi

import (
	"testing"

	"github.com/jf-tech/go-corelib/maths"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestSegDeclIsGroup(t *testing.T) {
	assert.False(t, (&segDecl{}).isGroup())
	assert.False(t, (&segDecl{Type: strs.StrPtr(segTypeSeg)}).isGroup())
	assert.False(t, (&segDecl{Type: strs.StrPtr("something")}).isGroup())
	assert.True(t, (&segDecl{Type: strs.StrPtr(segTypeGroup)}).isGroup())
}

func TestSegDeclMinMaxOccurs(t *testing.T) {
	for _, test := range []struct {
		name        string
		min         *int
		max         *int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "default",
			min:         nil,
			max:         nil,
			expectedMin: 1,
			expectedMax: 1,
		},
		{
			name:        "min/max=0",
			min:         testlib.IntPtr(0),
			max:         testlib.IntPtr(0),
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:        "max unlimited",
			min:         nil,
			max:         testlib.IntPtr(-1),
			expectedMin: 1,
			expectedMax: maths.MaxIntValue,
		},
		{
			name:        "min/max finite",
			min:         testlib.IntPtr(3),
			max:         testlib.IntPtr(5),
			expectedMin: 3,
			expectedMax: 5,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := &segDecl{Min: test.min, Max: test.max}
			assert.Equal(t, test.expectedMin, s.minOccurs())
			assert.Equal(t, test.expectedMax, s.maxOccurs())
		})
	}
}

func verifySegDeclRt(t *testing.T, rt *segDeclRuntime, decl *segDecl, parentRt *segDeclRuntime) {
	assert.NotNil(t, rt)
	assert.Equal(t, len(decl.Children), len(rt.children))
	assert.True(t, rt.parent == parentRt)
	assert.Equal(t, 0, rt.occurred)
	assert.Equal(t, *decl, *rt.decl)
}

func verifyRoot(t *testing.T, root *segDeclRuntime, childSegDecls ...*segDecl) {
	verifySegDeclRt(
		t,
		root,
		&segDecl{
			Name:     rootSegName,
			Type:     strs.StrPtr(segTypeGroup),
			Children: childSegDecls,
		},
		nil)
}

func TestNewSegDeclRuntimeTree_Empty(t *testing.T) {
	verifyRoot(t, newSegDeclRuntimeTree())
}

func TestNewSegDeclRuntimeTree(t *testing.T) {
	/*
	   ISA
	       GS
	           ST
	           B10
	       GE
	   IEA
	*/
	st := &segDecl{Name: "ST"}
	b10 := &segDecl{Name: "B10"}
	gs := &segDecl{Name: "GS", Children: []*segDecl{st, b10}}
	ge := &segDecl{Name: "GE"}
	isa := &segDecl{Name: "ISA", Children: []*segDecl{gs, ge}}
	iea := &segDecl{Name: "IEA"}

	r := newSegDeclRuntimeTree(isa, iea)
	verifyRoot(t, r, isa, iea)

	isaRt := r.children[0]
	verifySegDeclRt(t, isaRt, isa, r)
	assert.Equal(t, "ISA", isaRt.fqdn)

	ieaRt := r.children[1]
	verifySegDeclRt(t, ieaRt, iea, r)
	assert.Equal(t, "IEA", ieaRt.fqdn)

	gsRt := isaRt.children[0]
	verifySegDeclRt(t, gsRt, gs, isaRt)
	assert.Equal(t, "ISA/GS", gsRt.fqdn)

	geRt := isaRt.children[1]
	verifySegDeclRt(t, geRt, ge, isaRt)
	assert.Equal(t, "ISA/GE", geRt.fqdn)

	stRt := gsRt.children[0]
	verifySegDeclRt(t, stRt, st, gsRt)
	assert.Equal(t, "ISA/GS/ST", stRt.fqdn)

	b10Rt := gsRt.children[1]
	verifySegDeclRt(t, b10Rt, b10, gsRt)
	assert.Equal(t, "ISA/GS/B10", b10Rt.fqdn)
}

func TestMatchSegName(t *testing.T) {
	/*
	   ISA
	       GS
	           ST
	               B10
	       GE
	*/
	b10 := &segDecl{Name: "B10"}
	st := &segDecl{Name: "ST", Type: strs.StrPtr(segTypeGroup), Children: []*segDecl{b10}}
	gs := &segDecl{Name: "GS", Type: strs.StrPtr(segTypeGroup), Children: []*segDecl{st}}
	ge := &segDecl{Name: "GE"}
	isa := &segDecl{Name: "ISA", Children: []*segDecl{gs, ge}}

	r := newSegDeclRuntimeTree(isa)
	verifyRoot(t, r, isa)

	isaRt := r.children[0]
	assert.True(t, isaRt.matchSegName("ISA"))
	assert.False(t, isaRt.matchSegName("GS"))
	assert.False(t, isaRt.matchSegName("ST"))
	assert.False(t, isaRt.matchSegName("B10"))

	gsRt := isaRt.children[0]
	assert.False(t, gsRt.matchSegName("GS"))
	assert.False(t, gsRt.matchSegName("ST"))
	assert.True(t, gsRt.matchSegName("B10"))

	stRt := gsRt.children[0]
	assert.False(t, stRt.matchSegName("ST"))
	assert.True(t, stRt.matchSegName("B10"))
}

func TestResetChildrenOccurred(t *testing.T) {
	/*
	   ISA
	       GS
	           ST
	           B10
	       GE
	   IEA
	*/
	st := &segDecl{Name: "ST"}
	b10 := &segDecl{Name: "B10"}
	gs := &segDecl{Name: "GS", Children: []*segDecl{st, b10}}
	ge := &segDecl{Name: "GE"}
	isa := &segDecl{Name: "ISA", Children: []*segDecl{gs, ge}}
	iea := &segDecl{Name: "IEA"}

	r := newSegDeclRuntimeTree(isa, iea)
	verifyRoot(t, r, isa, iea)

	isaRt := r.children[0]
	isaRt.occurred = 1

	ieaRt := r.children[1]
	ieaRt.occurred = 2

	gsRt := isaRt.children[0]
	gsRt.occurred = 3

	geRt := isaRt.children[1]
	geRt.occurred = 4

	stRt := gsRt.children[0]
	stRt.occurred = 5

	b10Rt := gsRt.children[1]
	b10Rt.occurred = 6

	gsRt.resetChildrenOccurred()
	assert.Equal(t, 0, stRt.occurred)
	assert.Equal(t, 0, b10Rt.occurred)
	assert.Equal(t, 1, isaRt.occurred)
	assert.Equal(t, 2, ieaRt.occurred)
	assert.Equal(t, 3, gsRt.occurred)
	assert.Equal(t, 4, geRt.occurred)
}

package edi

import (
	"testing"

	"github.com/jf-tech/go-corelib/maths"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestElemCompIndex(t *testing.T) {
	assert.Equal(t, 1, elem{}.compIndex())
	assert.Equal(t, 123, elem{CompIndex: testlib.IntPtr(123)}.compIndex())
}

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

func TestSegDeclMatchSegName(t *testing.T) {
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

	assert.True(t, isa.matchSegName("ISA"))
	assert.False(t, isa.matchSegName("GS"))
	assert.False(t, isa.matchSegName("ST"))
	assert.False(t, isa.matchSegName("B10"))

	assert.False(t, gs.matchSegName("GS"))
	assert.False(t, gs.matchSegName("ST"))
	assert.True(t, gs.matchSegName("B10"))

	assert.False(t, st.matchSegName("ST"))
	assert.True(t, st.matchSegName("B10"))
}

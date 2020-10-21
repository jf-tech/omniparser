package edi

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

func TestRawSeg(t *testing.T) {
	rawSegName := "test"
	rawSegData := []byte("test data")
	r := ediReader{
		unprocessedSegData: newRawSeg(),
	}
	assert.False(t, r.unprocessedSegData.valid)
	assert.Equal(t, "", r.unprocessedSegData.name)
	assert.Nil(t, r.unprocessedSegData.raw)
	assert.Equal(t, 0, len(r.unprocessedSegData.elems))
	assert.Equal(t, defaultElemsPerSeg, cap(r.unprocessedSegData.elems))
	r.unprocessedSegData.valid = true
	r.unprocessedSegData.name = rawSegName
	r.unprocessedSegData.raw = rawSegData
	r.unprocessedSegData.elems = append(
		r.unprocessedSegData.elems, rawSegElem{1, 1, rawSegData[0:4]}, rawSegElem{2, 1, rawSegData[5:]})
	r.resetRawSeg()
	assert.False(t, r.unprocessedSegData.valid)
	assert.Equal(t, "", r.unprocessedSegData.name)
	assert.Nil(t, r.unprocessedSegData.raw)
	assert.Equal(t, 0, len(r.unprocessedSegData.elems))
	assert.Equal(t, defaultElemsPerSeg, cap(r.unprocessedSegData.elems))
}

// Adding a benchmark for rawSeg operation to ensure there is no alloc:
// BenchmarkRawSeg-8   	81410766	        13.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkRawSeg(b *testing.B) {
	rawSegName := "test"
	rawSegData := []byte("test data")
	r := ediReader{
		unprocessedSegData: newRawSeg(),
	}
	for i := 0; i < b.N; i++ {
		r.resetRawSeg()
		r.unprocessedSegData.valid = true
		r.unprocessedSegData.name = rawSegName
		r.unprocessedSegData.raw = rawSegData
		r.unprocessedSegData.elems = append(
			r.unprocessedSegData.elems, rawSegElem{1, 1, rawSegData[0:4]}, rawSegElem{2, 1, rawSegData[5:]})
	}
}

func TestStack(t *testing.T) {
	r := ediReader{
		stack: newStack(),
	}
	assert.Equal(t, 0, len(r.stack))
	assert.Equal(t, defaultStackDepth, cap(r.stack))
	// try to access top of stack while there is nothing in it => panic.
	assert.PanicsWithValue(t,
		"frame requested: 0, but stack length: 0",
		func() {
			r.stackTop()
		})
	// try to shrink empty stack => panic.
	assert.PanicsWithValue(t,
		"stack length is empty",
		func() {
			r.shrinkStack()
		})
	newEntry1 := stackEntry{
		segDecl:  &segDecl{},
		segNode:  idr.CreateNode(idr.TextNode, "test"),
		curChild: 5,
		occurred: 10,
	}
	r.growStack(newEntry1)
	assert.Equal(t, 1, len(r.stack))
	assert.Equal(t, newEntry1, *r.stackTop())
	newEntry2 := stackEntry{
		segDecl:  &segDecl{},
		segNode:  idr.CreateNode(idr.TextNode, "test 2"),
		curChild: 10,
		occurred: 20,
	}
	r.growStack(newEntry2)
	assert.Equal(t, 2, len(r.stack))
	assert.Equal(t, newEntry2, *r.stackTop())
	// try to access a frame that doesn't exist => panic.
	assert.PanicsWithValue(t,
		"frame requested: 2, but stack length: 2",
		func() {
			r.stackTop(2)
		})
	assert.Equal(t, newEntry1, *r.shrinkStack())
	assert.Nil(t, r.shrinkStack())
}

// Adding a benchmark for stack operation to ensure there is no alloc:
// BenchmarkStack-8    	12901227	        89.0 ns/op	       0 B/op	       0 allocs/op
func BenchmarkStack(b *testing.B) {
	r := ediReader{
		stack: newStack(),
	}
	for i := 0; i < b.N; i++ {
		for j := 0; j < 20; j++ {
			r.growStack(stackEntry{})
		}
		for r.shrinkStack() != nil {
		}
	}
}

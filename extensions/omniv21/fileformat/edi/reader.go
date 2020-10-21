package edi

import (
	"bufio"
	"fmt"

	"github.com/jf-tech/omniparser/idr"
)

type rawSegElem struct {
	elemIndex int // for this piece of data, the index of which element it belongs to. 1-based.
	compIndex int // for this piece of data, the index of which component it belongs to. 1-based.
	data      []byte
}

type rawSeg struct {
	name  string       // name of the segment, e.g. 'ISA', 'GS', etc.
	raw   []byte       // the raw data of the entire segment, including segment delimiter.
	elems []rawSegElem // all the broken down pieces of elements of the segment.
}

const (
	defaultElemsPerSeg = 20
)

func newRawSeg() rawSeg {
	return rawSeg{
		elems: make([]rawSegElem, 0, defaultElemsPerSeg),
	}
}

type stackEntry struct {
	segDecl  *segDecl  // the current stack entry's segment decl
	segNode  *idr.Node // the current stack entry segment's IDR node
	curChild int       // which child segment is the current segment is processing.
	occurred int       // how many times the current segment is fully processed.
}

const (
	defaultStackDepth = 50
)

func newStack() []stackEntry {
	return make([]stackEntry, 0, defaultStackDepth)
}

type ediReader struct {
	filename           string
	scanner            *bufio.Scanner
	segDelim           string
	elemDelim          string
	compDelim          *string
	releaseChar        *rune
	stack              []stackEntry
	target             *idr.Node
	runeCount          int
	unprocessedSegData rawSeg
}

func inRange(i, lowerBoundInclusive, upperBoundInclusive int) bool {
	return i >= lowerBoundInclusive && i <= upperBoundInclusive
}

func (r *ediReader) resetRawSeg() {
	r.unprocessedSegData.name = ""
	r.unprocessedSegData.raw = nil
	r.unprocessedSegData.elems = r.unprocessedSegData.elems[:0]
}

// stackTop returns the pointer to the 'frame'-th stack entry from the top.
// 'frame' is optional, if not specified, default 0 (aka the very top of
// the stack) is assumed. Note caller NEVER owns the memory of the returned
// entry, thus caller can use the pointer and its data values inside locally
// but should never cache/save it somewhere for later usage.
func (r *ediReader) stackTop(frame ...int) *stackEntry {
	nth := 0
	if len(frame) == 1 {
		nth = frame[0]
	}
	if !inRange(nth, 0, len(r.stack)-1) {
		panic(fmt.Sprintf("frame requested: %d, but stack length: %d", nth, len(r.stack)))
	}
	return &r.stack[len(r.stack)-nth-1]
}

// shrinkStack removes the top frame of the stack and returns the pointer to the NEW TOP
// FRAME to caller. Note caller NEVER owns the memory of the returned entry, thus caller can
// use the pointer and its data values inside locally but should never cache/save it somewhere
// for later usage.
func (r *ediReader) shrinkStack() *stackEntry {
	if len(r.stack) < 1 {
		panic("stack length is empty")
	}
	r.stack = r.stack[:len(r.stack)-1]
	if len(r.stack) < 1 {
		return nil
	}
	return &r.stack[len(r.stack)-1]
}

// growStack adds a new stack entry to the top of the stack.
func (r *ediReader) growStack(e stackEntry) {
	r.stack = append(r.stack, e)
}

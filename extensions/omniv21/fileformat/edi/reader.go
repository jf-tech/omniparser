package edi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/idr"
)

// ErrInvalidEDI indicates the EDI content is corrupted. This is a fatal, non-continuable error.
type ErrInvalidEDI string

func (e ErrInvalidEDI) Error() string { return string(e) }

// IsErrInvalidEDI checks if an err is of ErrInvalidEDI type.
func IsErrInvalidEDI(err error) bool {
	switch err.(type) {
	case ErrInvalidEDI:
		return true
	default:
		return false
	}
}

type rawSegElem struct {
	elemIndex     int // for this piece of data, the index of which element it belongs to. 1-based.
	compIndex     int // for this piece of data, the index of which component it belongs to. 1-based.
	data          []byte
	dateUnescaped bool
}

type rawSeg struct {
	valid bool
	name  string       // name of the segment, e.g. 'ISA', 'GS', etc.
	raw   []byte       // the raw data of the entire segment, including segment delimiter.
	elems []rawSegElem // all the broken down pieces of elements of the segment.
}

const (
	defaultElemsPerSeg  = 32
	defaultCompsPerElem = 8
)

func newRawSeg() rawSeg {
	return rawSeg{
		// don't want to over-allocate (defaultElemsPerSeg * defaultCompsPerElem), since
		// most EDI segments don't have equal number of components for each element --
		// using defaultElemsPerSeg is probably good enough.
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
	defaultStackDepth = 10
)

func newStack() []stackEntry {
	return make([]stackEntry, 0, defaultStackDepth)
}

type strPtrByte struct {
	strptr *string
	b      []byte
}

func newStrPtrByte(strptr *string) strPtrByte {
	var b []byte
	if strptr != nil {
		b = []byte(*strptr)
	}
	return strPtrByte{
		strptr: strptr,
		b:      b,
	}
}

type ediReader struct {
	inputName          string
	scanner            *bufio.Scanner
	segDelim           strPtrByte
	elemDelim          strPtrByte
	compDelim          strPtrByte
	releaseChar        strPtrByte
	stack              []stackEntry
	target             *idr.Node
	targetXPath        *xpath.Expr
	runeBegin, runeEnd int
	unprocessedRawSeg  rawSeg
}

func inRange(i, lowerBoundInclusive, upperBoundInclusive int) bool {
	return i >= lowerBoundInclusive && i <= upperBoundInclusive
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

func (r *ediReader) resetRawSeg() {
	r.unprocessedRawSeg.valid = false
	r.unprocessedRawSeg.name = ""
	r.unprocessedRawSeg.raw = nil
	r.unprocessedRawSeg.elems = r.unprocessedRawSeg.elems[:0]
}

func runeCountAndHasOnlyCRLF(b []byte) (int, bool) {
	runeCount := 0
	onlyCRLF := true
	for {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError {
			return runeCount, onlyCRLF
		}
		if r != '\n' && r != '\r' {
			onlyCRLF = false
		}
		runeCount++
		b = b[size:]
	}
}

var (
	crBytes = []byte("\r")
)

func (r *ediReader) getUnprocessedRawSeg() (rawSeg, error) {
	if r.unprocessedRawSeg.valid {
		return r.unprocessedRawSeg, nil
	}
	var token []byte
	for r.scanner.Scan() {
		b := r.scanner.Bytes()
		// In rare occasions inputs are not strict EDI per se - they sometimes have trailing empty lines
		// with only CR and/or LF. Let's be not so strict and ignore those lines.
		count, onlyCRLF := runeCountAndHasOnlyCRLF(b)
		r.runeBegin = r.runeEnd
		r.runeEnd += count
		if onlyCRLF {
			continue
		}
		token = b
		break
	}
	// We are here because:
	// 1. we find next token (i.e. segment), great, let's process it, OR
	// 2. r.scanner.Scan() returns false and it's EOF (note scanner never returns EOF, it just returns false
	//    on Scan() and Err() returns nil). We need to return EOF, OR
	// 3. r.scanner.Scan() returns false Err() returns err, need to return the err wrapped.
	err := r.scanner.Err()
	if err != nil {
		return rawSeg{}, ErrInvalidEDI(r.fmtErrStr("cannot read segment, err: %s", err.Error()))
	}
	if token == nil {
		return rawSeg{}, io.EOF
	}
	// From now on, the important thing is to operate on token (of []byte)without modification and without
	// allocation to keep performance.
	r.unprocessedRawSeg.raw = token
	// First we need to drop the trailing segment delimiter.
	noSegDelim := token[:len(token)-len(r.segDelim.b)]
	// In rare occasions, input uses '\n' as segment delimiter, but '\r' somehow
	// gets included as well (more common in business platform running on Windows)
	// Drop that '\r' as well.
	if *r.segDelim.strptr == "\n" && bytes.HasSuffix(noSegDelim, crBytes) {
		noSegDelim = noSegDelim[:len(noSegDelim)-utf8.RuneLen('\r')]
	}
	for i, elem := range strs.ByteSplitWithEsc(noSegDelim, r.elemDelim.b, r.releaseChar.b, defaultElemsPerSeg) {
		if len(r.compDelim.b) == 0 {
			// if we don't have comp delimiter, treat the entire element as one component.
			r.unprocessedRawSeg.elems = append(
				r.unprocessedRawSeg.elems,
				rawSegElem{
					// while (element) index in schema starts with 1, it actually refers to the first element
					// AFTER the seg name element, thus we use can i as elemIndex directly.
					elemIndex: i,
					// comp_index always starts with 1
					compIndex: 1,
					data:      elem,
				})
			continue
		}
		for j, comp := range strs.ByteSplitWithEsc(elem, r.compDelim.b, r.releaseChar.b, defaultCompsPerElem) {
			r.unprocessedRawSeg.elems = append(
				r.unprocessedRawSeg.elems,
				rawSegElem{
					elemIndex: i,
					compIndex: j + 1,
					data:      comp,
				})
		}
	}
	if len(r.unprocessedRawSeg.elems) == 0 || len(r.unprocessedRawSeg.elems[0].data) == 0 {
		return rawSeg{}, ErrInvalidEDI(r.fmtErrStr("missing segment name"))
	}
	r.unprocessedRawSeg.name = string(r.unprocessedRawSeg.elems[0].data)
	r.unprocessedRawSeg.valid = true
	return r.unprocessedRawSeg, nil
}

func (r *ediReader) rawSegToNode(segDecl *segDecl) (*idr.Node, error) {
	if !r.unprocessedRawSeg.valid {
		panic("unprocessedRawSeg is not valid")
	}
	n := idr.CreateNode(idr.ElementNode, segDecl.Name)
	// Note: we assume segDecl.Elems are sorted by elemIndex/compIndex.
	// TODO: do the sorting validation.
	rawElemIndex := 0
	rawElems := r.unprocessedRawSeg.elems
	for _, elemDecl := range segDecl.Elems {
		for ; rawElemIndex < len(rawElems); rawElemIndex++ {
			if rawElems[rawElemIndex].elemIndex == elemDecl.Index &&
				rawElems[rawElemIndex].compIndex == elemDecl.compIndex() {
				break
			}
		}
		if rawElemIndex < len(rawElems) || elemDecl.EmptyIfMissing {
			elemN := idr.CreateNode(idr.ElementNode, elemDecl.Name)
			idr.AddChild(n, elemN)
			data := ""
			if rawElemIndex < len(rawElems) {
				data = string(strs.ByteUnescape(rawElems[rawElemIndex].data, r.releaseChar.b, true))
				rawElems[rawElemIndex].dateUnescaped = true
				rawElemIndex++
			}
			elemV := idr.CreateNode(idr.TextNode, data)
			idr.AddChild(elemN, elemV)
			continue
		}
		return nil, ErrInvalidEDI(
			r.fmtErrStr("unable to find element '%s' on segment '%s'", elemDecl.Name, segDecl.fqdn))
	}
	return n, nil
}

// segDone wraps up the processing of an instance of current segment (which includes the processing of
// the instances of its child segments). segDone marks streaming target if necessary. If the number of
// instance occurrences is over the current segment's max limit, segDone calls segNext to move to the
// next segment in sequence; If the number of instances is still within max limit, segDone does no more
// action so the current segment will remain on top of the stack and potentially process more instances
// of this segment. Note: segDone is potentially recursive: segDone -> segNext -> segDone -> ...
func (r *ediReader) segDone() {
	cur := r.stackTop()
	cur.curChild = 0
	cur.occurred++
	if cur.segDecl.IsTarget {
		if r.target != nil {
			panic("r.target != nil")
		}
		if cur.segNode == nil {
			panic("cur.segNode == nil")
		}
		if r.targetXPath == nil || idr.MatchAny(cur.segNode, r.targetXPath) {
			r.target = cur.segNode
		} else {
			idr.RemoveAndReleaseTree(cur.segNode)
			cur.segNode = nil
		}
	}
	if cur.occurred < cur.segDecl.maxOccurs() {
		return
	}
	// we're here because `cur.occurred >= cur.segDecl.maxOccurs()`
	// and the only path segNext() can fail is to have
	// `cur.occurred < cur.segDecl.minOccurs()`, which means
	// the calling to segNext() from segDone() will never fail,
	// if our validation makes sure min<=max.
	_ = r.segNext()
}

// segNext is called when the top-of-stack (aka current) segment is done its full processing and needs to move
// to the next segment. If the current segment has a subsequent sibling, that sibling will be the next segment;
// If not, it indicates the current segment's parent segment is fully done its processing, thus parent's segDone
// is called. Note: segNext is potentially recursive: segNext -> segDone -> segNext -> ...
func (r *ediReader) segNext() error {
	cur := r.stackTop()
	if cur.occurred < cur.segDecl.minOccurs() {
		// the current values of [begin, end] cover the current instance of the current seg. But the error
		// we're about to create is about the missing of next instance of the current seg. So just assign
		// 'end' to 'begin' to make the error msg less confusing.
		r.runeBegin = r.runeEnd
		return ErrInvalidEDI(r.fmtErrStr("segment '%s' needs min occur %d, but only got %d",
			cur.segDecl.Name, cur.segDecl.minOccurs(), cur.occurred))
	}
	if len(r.stack) <= 1 {
		return nil
	}
	cur = r.shrinkStack()
	if cur.curChild < len(cur.segDecl.Children)-1 {
		cur.curChild++
		r.growStack(stackEntry{segDecl: cur.segDecl.Children[cur.curChild]})
		return nil
	}
	r.segDone()
	return nil
}

// Read processes EDI input and returns an instance of the streaming target (aka the segment marked with is_target=true)
// The basic idea is a forever for-loop, inside which it reads out an unprocessed segment data, tries to see
// if the segment data matches what's the current segment decl we're processing: if matches, great, creates a new
// instance of the current segment decl with the data; if not, we call segNext to move the next segment decl inline, and
// continue the for-loop so next iteration, the same unprocessed data will be matched against the new segment decl.
func (r *ediReader) Read() (*idr.Node, error) {
	if r.target != nil {
		// This is just in case Release() isn't called by ingester.
		idr.RemoveAndReleaseTree(r.target)
		r.target = nil
	}
	for {
		if r.target != nil {
			return r.target, nil
		}
		rawSeg, err := r.getUnprocessedRawSeg()
		if err == io.EOF {
			// When the input is done, we still need to verified all the
			// remaining segs' min occurs are satisfied. We can do so by
			// simply keeping on moving to the next seg: we call segNext()
			// once at a time - in case after the segNext() call, the reader
			// yields another target node. We can safely do this (1 segNext()
			// call at a time after we counter EOF) is because getUnprocessedRawSeg()
			// will repeatedly return EOF.
			if len(r.stack) <= 1 {
				return nil, io.EOF
			}
			err = r.segNext()
			if err != nil {
				return nil, err
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		cur := r.stackTop()
		if !cur.segDecl.matchSegName(rawSeg.name) {
			err := r.segNext()
			if err != nil {
				return nil, err
			}
			continue
		}
		if !cur.segDecl.isGroup() {
			cur.segNode, err = r.rawSegToNode(cur.segDecl)
			if err != nil {
				return nil, err
			}
			r.resetRawSeg()
		} else {
			cur.segNode = idr.CreateNode(idr.ElementNode, cur.segDecl.Name)
		}
		if len(r.stack) > 1 {
			idr.AddChild(r.stackTop(1).segNode, cur.segNode)
		}
		if len(cur.segDecl.Children) > 0 {
			r.growStack(stackEntry{segDecl: cur.segDecl.Children[0]})
			continue
		}
		r.segDone()
	}
}

func (r *ediReader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' between character [%d,%d]: %s",
		r.inputName, r.runeBegin, r.runeEnd, fmt.Sprintf(format, args...))
}

const (
	scannerFlags = ios.ScannerByDelimFlagEofNotAsDelim | ios.ScannerByDelimFlagIncludeDelimInReturn
)

var (
	// ReaderBufSize is the default buf size for EDI reader. Making it too small might increase
	// mem-alloc and gc; making it too big increases the initial memory consumption footprint
	// (unnecessarily) for each reader creation which eventually leads to gc as well. Make it
	// exported so caller can experiment and set their optimal value.
	ReaderBufSize = 128
)

// NewReader creates an FormatReader for EDI file format.
func NewReader(inputName string, r io.Reader, decl *fileDecl, targetXPath string) (*ediReader, error) {
	segDelim := newStrPtrByte(&decl.SegDelim)
	elemDelim := newStrPtrByte(&decl.ElemDelim)
	compDelim := newStrPtrByte(decl.CompDelim)
	releaseChar := newStrPtrByte(decl.ReleaseChar)
	scanner := ios.NewScannerByDelim3(r, segDelim.b, releaseChar.b, scannerFlags, make([]byte, ReaderBufSize))
	targetXPathExpr, err := func() (*xpath.Expr, error) {
		if targetXPath == "" || targetXPath == "." {
			return nil, nil
		}
		return caches.GetXPathExpr(targetXPath)
	}()
	if err != nil {
		return nil, fmt.Errorf("invalid target xpath '%s', err: %s", targetXPath, err.Error())
	}
	reader := &ediReader{
		inputName:         inputName,
		scanner:           scanner,
		segDelim:          segDelim,
		elemDelim:         elemDelim,
		compDelim:         compDelim,
		releaseChar:       releaseChar,
		stack:             newStack(),
		targetXPath:       targetXPathExpr,
		runeBegin:         1,
		runeEnd:           1,
		unprocessedRawSeg: newRawSeg(),
	}
	reader.growStack(stackEntry{
		segDecl: &segDecl{
			Name:     rootSegName,
			Type:     strs.StrPtr(segTypeGroup),
			Children: decl.SegDecls,
			fqdn:     rootSegName,
		},
		segNode: idr.CreateNode(idr.DocumentNode, rootSegName),
	})
	if len(decl.SegDecls) > 0 {
		reader.growStack(stackEntry{
			segDecl: decl.SegDecls[0],
		})
	}
	return reader, nil
}

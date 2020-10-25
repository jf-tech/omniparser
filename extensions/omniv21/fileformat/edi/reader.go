package edi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

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
		return rawSeg{}, ErrInvalidEDI(r.fmtErrStr("segment is malformed, missing segment name"))
	}
	r.unprocessedRawSeg.name = string(r.unprocessedRawSeg.elems[0].data)
	r.unprocessedRawSeg.valid = true
	return r.unprocessedRawSeg, nil
}

func (r *ediReader) rawSegToNode(segDecl *segDecl) (*idr.Node, error) {
	if !r.unprocessedRawSeg.valid {
		panic("invalid state - unprocessedRawSeg is not valid")
	}
	if segDecl.FixedLengthInBytes != nil && len(r.unprocessedRawSeg.raw) != *segDecl.FixedLengthInBytes {
		return nil, ErrInvalidEDI(
			r.fmtErrStr("segment '%s' expected length %d byte(s), but got: %d byte(s)",
				segDecl.fqdn, *segDecl.FixedLengthInBytes, len(r.unprocessedRawSeg.raw)))
	}
	n := idr.CreateNode(idr.ElementNode, segDecl.Name)
	for _, elem := range segDecl.Elems {
		index := -1
		for i := 0; i < len(r.unprocessedRawSeg.elems); i++ {
			if r.unprocessedRawSeg.elems[i].elemIndex == elem.Index &&
				r.unprocessedRawSeg.elems[i].compIndex == elem.compIndex() {
				index = i
				break
			}
		}
		if index >= 0 || elem.EmptyIfMissing {
			data := ""
			if index >= 0 {
				data = string(strs.ByteUnescape(r.unprocessedRawSeg.elems[index].data, r.releaseChar.b, true))
				r.unprocessedRawSeg.elems[index].dateUnescaped = true
			}
			elemN := idr.CreateNode(idr.ElementNode, elem.Name)
			idr.AddChild(n, elemN)
			elemV := idr.CreateNode(idr.TextNode, data)
			idr.AddChild(elemN, elemV)
			continue
		}
		return nil, ErrInvalidEDI(r.fmtErrStr("unable to find element '%s' on segment '%s'", elem.Name, segDecl.fqdn))
	}
	return n, nil
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
func NewReader(inputName string, r io.Reader, decl *fileDecl) *ediReader {
	segDelim := newStrPtrByte(&decl.SegDelim)
	elemDelim := newStrPtrByte(&decl.ElemDelim)
	compDelim := newStrPtrByte(decl.CompDelim)
	releaseChar := newStrPtrByte(decl.ReleaseChar)
	reader := &ediReader{
		inputName:         inputName,
		scanner:           ios.NewScannerByDelim3(r, segDelim.b, releaseChar.b, scannerFlags, make([]byte, ReaderBufSize)),
		segDelim:          segDelim,
		elemDelim:         elemDelim,
		compDelim:         compDelim,
		releaseChar:       releaseChar,
		stack:             newStack(),
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
	})
	return reader
}

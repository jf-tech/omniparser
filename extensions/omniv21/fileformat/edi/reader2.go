package edi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/jf-tech/go-corelib/ios"
	"github.com/jf-tech/go-corelib/strs"
)

// ErrInvalidEDI indicates the EDI content is corrupted. This is a fatal, non-continuable error.
type ErrInvalidEDI string

func (e ErrInvalidEDI) Error() string { return string(e) }

// IsErrInvalidEDI checks if the `err` is of ErrInvalidEDI type.
func IsErrInvalidEDI(err error) bool {
	switch err.(type) {
	case ErrInvalidEDI:
		return true
	default:
		return false
	}
}

// RawSegElem represents an element or a component of a raw segment of an EDI document.
type RawSegElem struct {
	// ElemIndex is 1-based element index of this data inside the segment.
	ElemIndex int
	// CompIndex is 1-based component index of this data inside the element.
	CompIndex int
	// Data contains the element or component data.
	// WARNING: the data is just a slice of the raw input, not a copy - so no modification!
	// WARNING: data isn't unescaped if escaping sequence (release_character) is used; to
	//          unescape, use strs.ByteUnescape.
	Data []byte
}

// RawSeg represents a raw segment of an EDI document.
type RawSeg struct {
	valid bool         // only for internal use.
	Name  string       // name of the segment, e.g. 'ISA', 'GS', etc.
	Raw   []byte       // the raw data of the entire segment, including segment delimiter. not owned, no mod!
	Elems []RawSegElem // all the broken down pieces of elements/components of the segment.
}

const (
	defaultElemsPerSeg  = 32
	defaultCompsPerElem = 8
)

func newRawSeg() RawSeg {
	return RawSeg{
		// don't want to over-allocate (defaultElemsPerSeg * defaultCompsPerElem), since
		// most EDI segments don't have equal number of components for each element --
		// using defaultElemsPerSeg is probably good enough.
		Elems: make([]RawSegElem, 0, defaultElemsPerSeg),
	}
}

func resetRawSeg(raw *RawSeg) {
	raw.valid = false
	raw.Name = ""
	raw.Raw = nil
	raw.Elems = raw.Elems[:0]
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
	lfBytes = []byte("\n")
)

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

// NonValidatingReader is an EDI segment reader that only reads out raw segments (its elements and components)
// directly without doing any segment structural/hierarchical validation.
type NonValidatingReader struct {
	scanner            *bufio.Scanner
	segDelim           strPtrByte
	elemDelim          strPtrByte
	compDelim          strPtrByte
	releaseChar        strPtrByte
	runeBegin, runeEnd int
	segCount           int
	rawSeg             RawSeg
}

// Read returns a raw segment of an EDI document. Note all the []byte are not a copy, so READONLY,
// no modification.
func (r *NonValidatingReader) Read() (RawSeg, error) {
	resetRawSeg(&r.rawSeg)
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
	r.segCount++
	// We are here because:
	// 1. we find next token (i.e. segment), great, let's process it, OR
	// 2. r.scanner.Scan() returns false, and it's EOF (note scanner never returns EOF, it just returns false
	//    on Scan() and Err() returns nil). We need to return EOF, OR
	// 3. r.scanner.Scan() returns false Err() returns err, need to return the `err` wrapped.
	err := r.scanner.Err()
	if err != nil {
		return RawSeg{}, ErrInvalidEDI(fmt.Sprintf("cannot read segment, err: %s", err.Error()))
	}
	if token == nil {
		return RawSeg{}, io.EOF
	}
	// From now on, the important thing is to operate on token (of []byte) without modification and without
	// allocation to keep performance.
	r.rawSeg.Raw = token
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
			r.rawSeg.Elems = append(
				r.rawSeg.Elems,
				RawSegElem{
					// while (element) index in schema starts with 1, it actually refers to the first element
					// AFTER the seg name element, thus we can use i as ElemIndex directly.
					ElemIndex: i,
					// comp_index always starts with 1
					CompIndex: 1,
					Data:      elem,
				})
			continue
		}
		for j, comp := range strs.ByteSplitWithEsc(elem, r.compDelim.b, r.releaseChar.b, defaultCompsPerElem) {
			r.rawSeg.Elems = append(
				r.rawSeg.Elems,
				RawSegElem{
					ElemIndex: i,
					CompIndex: j + 1,
					Data:      comp,
				})
		}
	}
	if len(r.rawSeg.Elems) == 0 || len(r.rawSeg.Elems[0].Data) == 0 {
		return RawSeg{}, ErrInvalidEDI("missing segment name")
	}
	r.rawSeg.Name = string(r.rawSeg.Elems[0].Data)
	r.rawSeg.valid = true
	return r.rawSeg, nil
}

// RuneBegin returns the current reader's beginning rune position.
func (r *NonValidatingReader) RuneBegin() int {
	return r.runeBegin
}

// RuneEnd returns the current reader's ending rune position.
func (r *NonValidatingReader) RuneEnd() int {
	return r.runeEnd
}

// SegCount returns the current reader's segment count.
func (r *NonValidatingReader) SegCount() int {
	return r.segCount
}

// NewNonValidatingReader creates an instance of NonValidatingReader.
func NewNonValidatingReader(r io.Reader, decl *FileDecl) *NonValidatingReader {
	segDelim := newStrPtrByte(&decl.SegDelim)
	elemDelim := newStrPtrByte(&decl.ElemDelim)
	compDelim := newStrPtrByte(decl.CompDelim)
	releaseChar := newStrPtrByte(decl.ReleaseChar)
	if decl.IgnoreCRLF {
		r = ios.NewBytesReplacingReader(r, crBytes, nil)
		r = ios.NewBytesReplacingReader(r, lfBytes, nil)
	}
	scanner := ios.NewScannerByDelim3(r, segDelim.b, releaseChar.b, scannerFlags, make([]byte, ReaderBufSize))
	return &NonValidatingReader{
		scanner:     scanner,
		segDelim:    segDelim,
		elemDelim:   elemDelim,
		compDelim:   compDelim,
		releaseChar: releaseChar,
		runeBegin:   1,
		runeEnd:     1,
		segCount:    0,
		rawSeg:      newRawSeg(),
	}
}

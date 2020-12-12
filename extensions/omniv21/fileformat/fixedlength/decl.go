package fixedlength

import (
	"fmt"
	"unicode/utf8"

	"github.com/jf-tech/go-corelib/caches"
)

// ByHeaderFooterDecl contains the header and footer regexp patterns for a `by_header_footer` envelope.
type ByHeaderFooterDecl struct {
	Header string `json:"header"`
	Footer string `json:"footer"`
}

// ColumnDecl describes fixed-length envelope column settings for omniparser reader.
type ColumnDecl struct {
	Name        string  `json:"name"`
	StartPos    int     `json:"start_pos"` // 1-based. and rune-based.
	Length      int     `json:"length"`    // rune-based length.
	LinePattern *string `json:"line_pattern"`
}

func (c *ColumnDecl) lineMatch(line []byte) bool {
	if c.LinePattern == nil {
		return true
	}
	// validated in validation code.
	r, _ := caches.GetRegex(*c.LinePattern)
	return r.Match(line)
}

func (c *ColumnDecl) lineToColumnValue(line []byte) string {
	// StartPos is 1-based and its value >= 1 guaranteed by json schema validation done earlier.
	start := c.StartPos - 1
	// First chop off the prefix prior to c.StartPos
	for start > 0 && len(line) > 0 {
		_, adv := utf8.DecodeRune(line)
		line = line[adv:]
		start--
	}
	// Then from that position, count c.Length runes and that's the string value we need.
	lenCount := c.Length
	i := 0
	for lenCount > 0 && i < len(line) {
		_, adv := utf8.DecodeRune(line[i:])
		i += adv
		lenCount--
	}
	return string(line[:i])
}

// EnvelopeDecl describes fixed-length envelope settings for omniparser reader.
type EnvelopeDecl struct {
	Name           *string             `json:"name"`
	ByHeaderFooter *ByHeaderFooterDecl `json:"by_header_footer"`
	ByRows         *int                `json:"by_rows"`
	NotTarget      bool                `json:"not_target"`
	Columns        []*ColumnDecl       `json:"columns"`
}

func (e *EnvelopeDecl) byRows() int {
	if e.ByHeaderFooter != nil {
		panic(fmt.Sprintf("envelope '%s' type is not 'by_rows'", *e.Name))
	}
	if e.ByRows == nil {
		return 1
	}
	return *e.ByRows
}

// FileDecl describes fixed-length specific schema settings for omniparser reader.
type FileDecl struct {
	Envelopes []*EnvelopeDecl `json:"envelopes"`
}

type envelopeType int

const (
	envelopeTypeByRows envelopeType = iota
	envelopeTypeByHeaderFooter
)

func (f *FileDecl) envelopeType() envelopeType {
	if f.Envelopes[0].ByHeaderFooter != nil {
		return envelopeTypeByHeaderFooter
	}
	return envelopeTypeByRows
}

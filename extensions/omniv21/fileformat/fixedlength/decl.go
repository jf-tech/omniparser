package fixedlength

import (
	"fmt"

	"github.com/jf-tech/go-corelib/caches"
)

type byHeaderFooterDecl struct {
	Header string `json:"header"`
	Footer string `json:"footer"`
}

type columnDecl struct {
	Name     string  `json:"name"`
	StartPos int     `json:"start_pos"` // 1-based. and rune-based.
	Length   int     `json:"length"`    // rune-based length.
	Line     *string `json:"line"`
}

type envelopeDecl struct {
	Name           *string             `json:"name"`
	ByHeaderFooter *byHeaderFooterDecl `json:"by_header_footer"`
	ByRows         *int                `json:"by_rows"`
	NotTarget      bool                `json:"not_target"`
	Columns        []*columnDecl       `json:"columns"`
}

type fileDecl struct {
	Envelopes []*envelopeDecl `json:"envelopes"`
}

func (c *columnDecl) lineMatch(line []byte) bool {
	if c.Line == nil {
		return true
	}
	// validated in validation code.
	r, _ := caches.GetRegex(*c.Line)
	return r.Match(line)
}

func (c *columnDecl) lineToColumn(line []rune) []rune {
	// StartPos is 1-based and its value >= 1 guaranteed by json schema validation done earlier.
	startPosZeroBased := c.StartPos - 1
	// If [startPosZeroBased, c.Length] is partially out of range, we'll return whatever is
	// in range; if [startPosZeroBased, c.Length] is fully out of range, we'll return "".
	switch {
	case startPosZeroBased+c.Length <= len(line):
		return line[startPosZeroBased : startPosZeroBased+c.Length]
	case startPosZeroBased < len(line):
		return line[startPosZeroBased:]
	}
	return nil
}

func (e *envelopeDecl) byRows() int {
	if e.ByHeaderFooter != nil {
		panic(fmt.Sprintf("envelope '%s' type is not 'by_rows'", *e.Name))
	}
	if e.ByRows == nil {
		return 1
	}
	return *e.ByRows
}

package fixedlength

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/jf-tech/go-corelib/maths"

	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile"
)

// ColumnDecl describes a column of an envelope.
type ColumnDecl struct {
	Name        string  `json:"name,omitempty"`
	StartPos    int     `json:"start_pos,omitempty"`  // 1-based. and rune-based.
	Length      int     `json:"length,omitempty"`     // rune-based length.
	LineIndex   *int    `json:"line_index,omitempty"` // 1-based.
	LinePattern *string `json:"line_pattern,omitempty"`

	linePatternRegexp *regexp.Regexp
}

func (c *ColumnDecl) lineMatch(lineIndex int, line []byte) bool {
	if c.LineIndex != nil {
		return *c.LineIndex == lineIndex+1 // c.LineIndex is 1 based.
	}
	if c.linePatternRegexp != nil {
		return c.linePatternRegexp.Match(line)
	}
	return true
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
	// Note if c.Length is longer than what's left in the line, we'll simply take all of
	// the remaining line (and no error here, since we haven't yet seen a useful case where
	// we need to be excessively strict.)
	lenCount := c.Length
	i := 0
	for lenCount > 0 && i < len(line) {
		_, adv := utf8.DecodeRune(line[i:])
		i += adv
		lenCount--
	}
	return string(line[:i])
}

const (
	typeEnvelope = "envelope"
	typeGroup    = "envelope_group"
)

// EnvelopeDecl describes an envelope of a fixed-length input.
// If Rows/Header/Footer are all nil, then it defaults to Rows = 1.
// If Rows specified, then Header/Footer must be nil. (JSON schema validation will ensure this.)
// If Header is specified, Rows must be nil. (JSON schema validation will ensure this.)
// Footer is optional; If not specified, Header will be used for a single-line envelope matching.
type EnvelopeDecl struct {
	Name     string          `json:"name,omitempty"`
	Rows     *int            `json:"rows,omitempty"`
	Header   *string         `json:"header,omitempty"`
	Footer   *string         `json:"footer,omitempty"`
	Type     *string         `json:"type,omitempty"`
	IsTarget bool            `json:"is_target,omitempty"`
	Min      *int            `json:"min,omitempty"`
	Max      *int            `json:"max,omitempty"`
	Columns  []*ColumnDecl   `json:"columns,omitempty"`
	Children []*EnvelopeDecl `json:"child_envelopes,omitempty"`

	fqdn          string // fully hierarchical name to the envelope.
	childRecDecls []flatfile.RecDecl
	headerRegexp  *regexp.Regexp
	footerRegexp  *regexp.Regexp
}

func (e *EnvelopeDecl) DeclName() string {
	return e.Name
}

func (e *EnvelopeDecl) Target() bool {
	return e.IsTarget
}

func (e *EnvelopeDecl) Group() bool {
	return e.Type != nil && *e.Type == typeGroup
}

// MinOccurs defaults to 0. Fixed-length input most common scenario is min=0/max=unbounded.
func (e *EnvelopeDecl) MinOccurs() int {
	switch e.Min {
	case nil:
		return 0
	default:
		return *e.Min
	}
}

// MaxOccurs defaults to unbounded. Fixed-length input most common scenario is min=0/max=unbounded.
func (e *EnvelopeDecl) MaxOccurs() int {
	switch {
	case e.Max == nil:
		fallthrough
	case *e.Max < 0:
		return maths.MaxIntValue
	default:
		return *e.Max
	}
}

func (e *EnvelopeDecl) ChildDecls() []flatfile.RecDecl {
	return e.childRecDecls
}

func (e *EnvelopeDecl) rowsBased() bool {
	if e.Group() {
		panic("envelope_group is neither rows based nor header/footer based")
	}
	// for header/footer based envelope, header must be specified; otherwise, it's rows based.
	return e.Header == nil
}

// rows() defaults to 1. Fixed-length input most common scenario is rows-based single line envelope.
func (e *EnvelopeDecl) rows() int {
	if !e.rowsBased() {
		panic(fmt.Sprintf("envelope '%s' is not rows based", e.fqdn))
	}
	if e.Rows == nil {
		return 1
	}
	return *e.Rows
}

func (e *EnvelopeDecl) matchHeader(line []byte) bool {
	if e.headerRegexp == nil {
		panic(fmt.Sprintf("envelope '%s' is not header/footer based", e.fqdn))
	}
	return e.headerRegexp.Match(line)
}

// Footer is optional. If not specified, it always matches. Thus for a header/footer envelope,
// if the footer isn't specified, it effectively becomes a single-row envelope matched by header,
// given that after the header matches a line, matchFooter is called on the same line.
func (e *EnvelopeDecl) matchFooter(line []byte) bool {
	if e.footerRegexp == nil {
		return true
	}
	return e.footerRegexp.Match(line)
}

func toFlatFileRecDecls(es []*EnvelopeDecl) []flatfile.RecDecl {
	if len(es) == 0 {
		return nil
	}
	ret := make([]flatfile.RecDecl, len(es))
	for i, d := range es {
		ret[i] = d
	}
	return ret
}

// FileDecl describes fixed-length schema `file_declaration` setting.
type FileDecl struct {
	Envelopes []*EnvelopeDecl `json:"envelopes,omitempty"`
}

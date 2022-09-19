package csv

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jf-tech/go-corelib/maths"

	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile"
)

// ColumnDecl describes a column of an csv record column.
type ColumnDecl struct {
	Name        string  `json:"name,omitempty"`
	Index       *int    `json:"index,omitempty"`        // 1-based. optional.
	LineIndex   *int    `json:"line_index,omitempty"`   // 1-based. optional
	LinePattern *string `json:"line_pattern,omitempty"` // optional

	linePatternRegexp *regexp.Regexp
}

func (c *ColumnDecl) lineMatch(lineIndex int, line *line, records []string, delim string) bool {
	if c.LineIndex != nil {
		return *c.LineIndex == lineIndex+1 // c.LineIndex is 1 based.
	}
	if c.linePatternRegexp != nil {
		return matchLine(c.linePatternRegexp, line, records, delim)
	}
	return true
}

func (c *ColumnDecl) lineToColumnValue(line *line, records []string) string {
	if *c.Index < 1 || *c.Index > line.recordNum {
		return ""
	}
	return records[line.recordStart+*c.Index-1]
}

const (
	typeRecord = "record"
	typeGroup  = "record_group"
)

// RecordDecl describes an record of a csv/delimited input.
// If Rows/Header/Footer are all nil, then it defaults to Rows = 1.
// If Rows specified, then Header/Footer must be nil. (JSON schema validation will ensure this.)
// If Header is specified, Rows must be nil. (JSON schema validation will ensure this.)
// Footer is optional; If not specified, Header will be used for a single-line record matching.
type RecordDecl struct {
	Name     string        `json:"name,omitempty"`
	Rows     *int          `json:"rows,omitempty"`
	Header   *string       `json:"header,omitempty"`
	Footer   *string       `json:"footer,omitempty"`
	Type     *string       `json:"type,omitempty"`
	IsTarget bool          `json:"is_target,omitempty"`
	Min      *int          `json:"min,omitempty"`
	Max      *int          `json:"max,omitempty"`
	Columns  []*ColumnDecl `json:"columns,omitempty"`
	Children []*RecordDecl `json:"child_records,omitempty"`

	fqdn          string // fully hierarchical name to the record.
	childRecDecls []flatfile.RecDecl
	headerRegexp  *regexp.Regexp
	footerRegexp  *regexp.Regexp
}

func (r *RecordDecl) DeclName() string {
	return r.Name
}

func (r *RecordDecl) Target() bool {
	return r.IsTarget
}

func (r *RecordDecl) Group() bool {
	return r.Type != nil && *r.Type == typeGroup
}

// MinOccurs defaults to 0. CSV/delimited input most common scenario is min=0/max=unbounded.
func (r *RecordDecl) MinOccurs() int {
	switch r.Min {
	case nil:
		return 0
	default:
		return *r.Min
	}
}

// MaxOccurs defaults to unbounded. CSV/delimited input most common scenario is min=0/max=unbounded.
func (r *RecordDecl) MaxOccurs() int {
	switch {
	case r.Max == nil:
		fallthrough
	case *r.Max < 0:
		return maths.MaxIntValue
	default:
		return *r.Max
	}
}

func (r *RecordDecl) ChildDecls() []flatfile.RecDecl {
	return r.childRecDecls
}

func (r *RecordDecl) rowsBased() bool {
	if r.Group() {
		panic("record_group is neither rows based nor header/footer based")
	}
	// for header/footer based record, header must be specified; otherwise, it's rows based.
	return r.Header == nil
}

// rows() defaults to 1. csv/delimited most common scenario is rows-based single line record.
func (r *RecordDecl) rows() int {
	if !r.rowsBased() {
		panic(fmt.Sprintf("record '%s' is not rows based", r.fqdn))
	}
	if r.Rows == nil {
		return 1
	}
	return *r.Rows
}

func (r *RecordDecl) matchHeader(line *line, records []string, delim string) bool {
	if r.headerRegexp == nil {
		panic(fmt.Sprintf("record '%s' is not header/footer based", r.fqdn))
	}
	return matchLine(r.headerRegexp, line, records, delim)
}

// Footer is optional. If not specified, it always matches. Thus for a header/footer record,
// if the footer isn't specified, it effectively becomes a single-row record matched by header,
// given that after the header matches a line, matchFooter is called on the same line.
func (r *RecordDecl) matchFooter(line *line, records []string, delim string) bool {
	if r.footerRegexp == nil {
		return true
	}
	return matchLine(r.footerRegexp, line, records, delim)
}

func matchLine(re *regexp.Regexp, line *line, records []string, delim string) bool {
	if line.raw == "" {
		line.raw = strings.Join(records[line.recordStart:line.recordStart+line.recordNum], delim)
	}
	return re.MatchString(line.raw)
}

func toFlatFileRecDecls(rs []*RecordDecl) []flatfile.RecDecl {
	if len(rs) == 0 {
		return nil
	}
	ret := make([]flatfile.RecDecl, len(rs))
	for i, r := range rs {
		ret[i] = r
	}
	return ret
}

// FileDecl describes csv/delimited schema `file_declaration` setting.
type FileDecl struct {
	Delimiter           string        `json:"delimiter,omitempty"`
	ReplaceDoubleQuotes bool          `json:"replace_double_quotes,omitempty"`
	Records             []*RecordDecl `json:"records,omitempty"`
}

package csv

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"
	"github.com/jf-tech/go-corelib/maths"

	"github.com/jf-tech/omniparser/idr"
)

// ErrInvalidHeader indicates the header of the CSV input is corrupted, mismatched, or simply
// unreadable. This is a fatal, non-continuable error.
type ErrInvalidHeader string

func (e ErrInvalidHeader) Error() string { return string(e) }

// IsErrInvalidHeader checks if the `err` is of ErrInvalidHeader type.
func IsErrInvalidHeader(err error) bool {
	switch err.(type) {
	case ErrInvalidHeader:
		return true
	default:
		return false
	}
}

type reader struct {
	inputName     string
	decl          *FileDecl
	xpath         *xpath.Expr
	r             *ios.LineNumReportingCsvReader
	headerChecked bool
}

func (r *reader) Read() (*idr.Node, error) {
	if !r.headerChecked {
		err := r.checkHeader()
		r.headerChecked = true
		if err != nil {
			return nil, err
		}
	}
read:
	record, err := r.r.Read()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, r.FmtErr("failed to fetch record: %s", err.Error())
	}
	n := r.recordToNode(record)
	if r.xpath != nil && !idr.MatchAny(n, r.xpath) {
		goto read
	}
	return n, nil
}

func (r *reader) checkHeader() error {
	var err error
	var header []string
	if r.decl.HeaderRowIndex == nil {
		goto skipToDataRow
	}
	err = r.jumpTo(*r.decl.HeaderRowIndex - 1)
	if err != nil {
		return ErrInvalidHeader(r.fmtErrStr("unable to read header: %s", err.Error()))
	}
	header, err = r.r.Read()
	if err != nil {
		return ErrInvalidHeader(r.fmtErrStr("unable to read header: %s", err.Error()))
	}
	if len(header) < len(r.decl.Columns) {
		return ErrInvalidHeader(r.fmtErrStr(
			"actual header column size (%d) is less than the size (%d) declared in file_declaration.columns in schema",
			len(header), len(r.decl.Columns)))
	}
	for index, column := range r.decl.Columns {
		if strings.TrimSpace(header[index]) != strings.TrimSpace(column.Name) {
			return ErrInvalidHeader(r.fmtErrStr(
				"header column[%d] '%s' does not match declared column name '%s' in schema",
				index+1, strings.TrimSpace(header[index]), strings.TrimSpace(column.Name)))
		}
	}
skipToDataRow:
	if err = r.jumpTo(r.decl.DataRowIndex - 1); err != nil {
		return err
	}
	return nil
}

// the only possible error this jumpTo returns is io.EOF. if there is any reading error, we'll ignore
// because we really don't care about what's corrupted in a line. Now it's possible, but very very
// rarely, that the input reader's underlying media fails to read due memory/disk/IO issue. Since we
// can't reliably tease apart those failures from a simple line corruption failure, we'll choose to
// ignore them equally. And those underlying media failures most likely will repeat and cause subsequent
// read to fail, and then the reader will fail out entirely.
func (r *reader) jumpTo(rowIndex int) error {
	for r.r.LineNum() < rowIndex {
		_, err := r.r.Read()
		if err == io.EOF {
			return io.EOF
		}
	}
	return nil
}

func (r *reader) recordToNode(record []string) *idr.Node {
	root := idr.CreateNode(idr.DocumentNode, "")
	// - If actual record has more columns than declared in schema, we'll only use up to
	//   what's declared in the schema;
	// - conversely, if the actual record has fewer columns than declared in schema, we'll
	//   use all that are in the record.
	for i := 0; i < maths.MinInt(len(record), len(r.decl.Columns)); i++ {
		col := idr.CreateNode(idr.ElementNode, r.decl.Columns[i].name())
		idr.AddChild(root, col)
		data := idr.CreateNode(idr.TextNode, record[i])
		idr.AddChild(col, data)
	}
	return root
}

func (r *reader) Release(n *idr.Node) {
	if n != nil {
		idr.RemoveAndReleaseTree(n)
	}
}

func (r *reader) IsContinuableError(err error) bool {
	return !IsErrInvalidHeader(err) && err != io.EOF
}

func (r *reader) FmtErr(format string, args ...interface{}) error {
	return errors.New(r.fmtErrStr(format, args...))
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' line %d: %s", r.inputName, r.r.LineNum(), fmt.Sprintf(format, args...))
}

// NewReader creates an FormatReader for CSV file format.
func NewReader(inputName string, r io.Reader, decl *FileDecl, xpathStr string) (*reader, error) {
	var expr *xpath.Expr
	var err error
	xpathStr = strings.TrimSpace(xpathStr)
	if xpathStr != "" && xpathStr != "." {
		expr, err = caches.GetXPathExpr(xpathStr)
		if err != nil {
			return nil, fmt.Errorf("invalid xpath '%s', err: %s", xpathStr, err.Error())
		}
	}
	if decl.ReplaceDoubleQuotes {
		r = ios.NewBytesReplacingReader(r, []byte(`"`), []byte(`'`))
	}
	csv := ios.NewLineNumReportingCsvReader(r)
	delim := []rune(decl.Delimiter)
	csv.Comma = delim[0]
	csv.FieldsPerRecord = -1
	csv.ReuseRecord = true
	return &reader{
		inputName:     inputName,
		decl:          decl,
		r:             csv,
		headerChecked: false,
		xpath:         expr,
	}, nil
}

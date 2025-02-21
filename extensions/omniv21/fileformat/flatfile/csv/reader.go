package csv

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/ios"

	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/idr"
)

type line struct {
	lineNum                int // 1-based
	recordStart, recordNum int // positional references into reader.records[] slice.
	raw                    string
}

const (
	debug          = "__debug"
	debug_line_num = "line_num"
)

type reader struct {
	schemaHeader header.Header
	inputName    string
	fileDecl     *FileDecl
	r            *ios.LineNumReportingCsvReader
	hr           *flatfile.HierarchyReader
	linesBuf     []line // linesBuf contains all the unprocessed lines
	records      []string
}

// NewReader creates an FormatReader for csv file format.
func NewReader(
	schemaHeader header.Header,
	inputName string,
	r io.Reader,
	decl *FileDecl,
	targetXPathExpr *xpath.Expr) *reader {
	if decl.ReplaceDoubleQuotes {
		r = ios.NewBytesReplacingReader(r, []byte(`"`), []byte(`'`))
	}
	csv := ios.NewLineNumReportingCsvReader(r)
	delim := []rune(decl.Delimiter)
	csv.Comma = delim[0]
	csv.FieldsPerRecord = -1
	// While csv.ReuseRecord = true minimize encoding/csv.Reader slice allocations,
	// It does make our multi-line caching a bit trickier. Since the csv.Reader.Read()
	// returned []string slice will be reused, we have to have our own slice to copy
	// those record string references down: reader.records[].
	csv.ReuseRecord = true
	reader := &reader{
		schemaHeader: schemaHeader,
		inputName:    inputName,
		fileDecl:     decl,
		r:            csv,
	}
	reader.hr = flatfile.NewHierarchyReader(
		toFlatFileRecDecls(decl.Records), reader, targetXPathExpr)
	return reader
}

// Read implements fileformat.FormatReader interface, reading in data from input and returns
// target IDR node.
func (r *reader) Read() (*idr.Node, error) {
	n, err := r.hr.Read()
	switch {
	case err == nil:
		return n, nil
	case flatfile.IsErrFewerThanMinOccurs(err):
		e := err.(flatfile.ErrFewerThanMinOccurs)
		decl := e.RecDecl.(*RecordDecl)
		return nil, ErrInvalidCSV(r.fmtErrStr(r.unprocessedLineNum(),
			"record/record_group '%s' needs min occur %d, but only got %d",
			decl.fqdn, decl.MinOccurs(), e.ActualOcccurs))
	case flatfile.IsErrUnexpectedData(err):
		return nil, ErrInvalidCSV(r.fmtErrStr(r.unprocessedLineNum(), "unexpected data"))
	default:
		return nil, err
	}
}

// MoreUnprocessedData implements flatfile.RecReader, telling whether there is still unprocessed
// data or not.
func (r *reader) MoreUnprocessedData() (bool, error) {
	if len(r.linesBuf) > 0 {
		return true, nil
	}
	if err := r.readLine(); err != nil && err != io.EOF {
		return false, err
	}
	return len(r.linesBuf) > 0, nil
}

// ReadAndMatch implements flatfile.RecReader, reading unprocessed data (from buffer or from IO),
// trying to match against the given non-group typed record decl, and converting the data into IDR
// node if asked to.
func (r *reader) ReadAndMatch(
	decl flatfile.RecDecl, createIDR bool) (matched bool, node *idr.Node, err error) {
	recDecl := decl.(*RecordDecl)
	if recDecl.rowsBased() {
		return r.readAndMatchRowsBasedRecord(recDecl, createIDR)
	}
	return r.readAndMatchHeaderFooterBasedRecord(recDecl, createIDR)
}

func (r *reader) readAndMatchRowsBasedRecord(
	decl *RecordDecl, createNode bool) (bool, *idr.Node, error) {
	for len(r.linesBuf) < decl.rows() {
		if err := r.readLine(); err != nil {
			if err != io.EOF || len(r.linesBuf) == 0 {
				// So either the err isn't io.EOF, we need to return this critical error directly;
				// or it's io.EOF and our line buf is empty, i.e. all has been processed, so we
				// should return io.EOF to indicate end. Both case, we simply return "not matched"
				// plus err.
				return false, nil, err
			}
			// we're here when the err is io.EOF and line buf isn't empty, so we return
			// "no match", and "no error", hoping the non-empty line buf will be matched
			// by subsequent calls with different decls.
			return false, nil, nil
		}
	}
	if createNode {
		n := r.linesToNode(decl, decl.rows())
		// Once those rows have been converted into IDR node, we're done with them, and remove
		// them from the unprocessed line buffer.
		r.popFrontLinesBuf(decl.rows())
		return true, n, nil
	}
	return true, nil, nil
}

func (r *reader) readAndMatchHeaderFooterBasedRecord(
	decl *RecordDecl, createNode bool) (bool, *idr.Node, error) {
	if len(r.linesBuf) <= 0 {
		if err := r.readLine(); err != nil {
			// io.EOF or not, since r.linesBuf is empty, we can directly return err.
			return false, nil, err
		}
	}
	if !decl.matchHeader(&r.linesBuf[0], r.records, r.fileDecl.Delimiter) {
		return false, nil, nil
	}
	i := 0 // we'll match the footer starting on the same line.
	for {
		if decl.matchFooter(&r.linesBuf[i], r.records, r.fileDecl.Delimiter) {
			if createNode {
				n := r.linesToNode(decl, i+1)
				r.popFrontLinesBuf(i + 1)
				return true, n, nil
			}
			return true, nil, nil
		}
		// if by the end of r.linesBuf we still haven't matched footer, we need to
		// read more line in for footer match.
		if i >= len(r.linesBuf)-1 {
			if err := r.readLine(); err != nil {
				if err != io.EOF { // io reading error, directly return err.
					return false, nil, err
				}
				// io.EOF encountered and since r.linesBuf isn't empty,
				// we need to return false for matching, but nil for error (we only return io.EOF
				// when r.linesBuf is empty.
				return false, nil, nil
			}
		}
		i++
	}
}

func (r *reader) readLine() error {
	lineStart := r.r.LineNum() + 1
	record, err := r.r.Read()
	switch {
	case err == io.EOF:
		return io.EOF
	case err != nil:
		return ErrInvalidCSV(r.fmtErrStr(lineStart, err.Error()))
	}
	start, num := len(r.records), len(record)
	r.records = append(r.records, record...)
	r.linesBuf = append(r.linesBuf, line{
		lineNum:     lineStart,
		recordStart: start,
		recordNum:   num,
	})
	return nil
}

func (r *reader) linesToNode(decl *RecordDecl, n int) *idr.Node {
	if len(r.linesBuf) < n {
		panic(fmt.Sprintf(
			"linesBuf has %d lines but requested %d lines to convert", len(r.linesBuf), n))
	}
	node := idr.CreateNode(idr.ElementNode, decl.Name)
	r.insertDebugInfo(node)
	for col := range decl.Columns {
		colDecl := decl.Columns[col]
		for i := 0; i < n; i++ {
			if !colDecl.lineMatch(i, &(r.linesBuf[i]), r.records, r.fileDecl.Delimiter) {
				continue
			}
			colNode := idr.CreateNode(idr.ElementNode, colDecl.Name)
			idr.AddChild(node, colNode)
			colVal := idr.CreateNode(
				idr.TextNode, colDecl.lineToColumnValue(&r.linesBuf[i], r.records))
			idr.AddChild(colNode, colVal)
			break
		}
	}
	return node
}

func (r *reader) insertDebugInfo(node *idr.Node) {
	if r.schemaHeader.ParserSettings.Debug == 0 {
		return
	}
	debugNode := idr.CreateNode(idr.ElementNode, debug)
	idr.AddChild(node, debugNode)
	lineNumNode := idr.CreateNode(idr.ElementNode, debug_line_num)
	idr.AddChild(debugNode, lineNumNode)
	lineNumValNode := idr.CreateNode(idr.TextNode, strconv.Itoa(r.linesBuf[0].lineNum))
	idr.AddChild(lineNumNode, lineNumValNode)
}

func (r *reader) popFrontLinesBuf(n int) {
	if n > len(r.linesBuf) {
		panic(fmt.Sprintf(
			"less lines (%d) in r.linesBuf than requested pop front count (%d)",
			len(r.linesBuf), n))
	}

	recordShift := 0
	for i := 0; i < n; i++ {
		recordShift += r.linesBuf[i].recordNum
	}
	copy(r.records, r.records[recordShift:])
	r.records = r.records[:len(r.records)-recordShift]

	newLinesBufLen := len(r.linesBuf) - n
	for i := 0; i < newLinesBufLen; i++ {
		r.linesBuf[i] = r.linesBuf[i+n]
		r.linesBuf[i].recordStart -= recordShift
	}
	r.linesBuf = r.linesBuf[:newLinesBufLen]
}

func (r *reader) unprocessedLineNum() int {
	if len(r.linesBuf) > 0 {
		return r.linesBuf[0].lineNum
	}
	return r.r.LineNum() + 1
}

// Release implements fileformat.FormatReader interface, releasing a finished IDR target node.
func (r *reader) Release(n *idr.Node) {
	r.hr.Release(n)
}

// IsContinuableError implements fileformat..FormatReader interface, checking if an error is
// fatal or not.
func (r *reader) IsContinuableError(err error) bool {
	return !IsErrInvalidCSV(err) && err != io.EOF
}

// FmtErr implements errs.CtxAwareErr embedded in fileformat.FormatReader, formatting an error
// with line info.
func (r *reader) FmtErr(format string, args ...interface{}) error {
	return errors.New(r.fmtErrStr(r.unprocessedLineNum(), format, args...))
}

func (r *reader) fmtErrStr(line int, format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' line %d: %s",
		r.inputName, line, fmt.Sprintf(format, args...))
}

// ErrInvalidCSV indicates the csv content is corrupted or IO failure.
// This is a fatal, non-continuable error.
type ErrInvalidCSV string

// Error implements error interface.
func (e ErrInvalidCSV) Error() string { return string(e) }

// IsErrInvalidCSV checks if the `err` is of ErrInvalidCSV type.
func IsErrInvalidCSV(err error) bool {
	switch err.(type) {
	case ErrInvalidCSV:
		return true
	default:
		return false
	}
}

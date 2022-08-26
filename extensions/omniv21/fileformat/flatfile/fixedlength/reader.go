package fixedlength

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/ios"

	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/flatfile"
	"github.com/jf-tech/omniparser/idr"
)

type line struct {
	lineNum int // 1-based
	b       []byte
}

type reader struct {
	inputName string
	r         *bufio.Reader
	hr        *flatfile.HierarchyReader
	linesRead int    // total number of lines read in so far
	linesBuf  []line // linesBuf contains all the unprocessed lines
}

// NewReader creates an FormatReader for fixed-length file format.
func NewReader(
	inputName string, r io.Reader, decl *FileDecl, targetXPathExpr *xpath.Expr) *reader {
	reader := &reader{
		inputName: inputName,
		r:         bufio.NewReader(r),
	}
	reader.hr = flatfile.NewHierarchyReader(
		toFlatFileRecDecls(decl.Envelopes), reader, targetXPathExpr)
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
		envelopeDecl := e.RecDecl.(*EnvelopeDecl)
		return nil, ErrInvalidFixedLength(r.fmtErrStr(r.unprocessedLineNum(),
			"envelope/envelope_group '%s' needs min occur %d, but only got %d",
			envelopeDecl.fqdn, envelopeDecl.MinOccurs(), e.ActualOcccurs))
	case flatfile.IsErrUnexpectedData(err):
		return nil, ErrInvalidFixedLength(r.fmtErrStr(r.unprocessedLineNum(), "unexpected data"))
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
	envelopeDecl := decl.(*EnvelopeDecl)
	if envelopeDecl.rowsBased() {
		return r.readAndMatchRowsBasedEnvelope(envelopeDecl, createIDR)
	}
	return r.readAndMatchHeaderFooterBasedEnvelope(envelopeDecl, createIDR)
}

func (r *reader) readAndMatchRowsBasedEnvelope(
	decl *EnvelopeDecl, createNode bool) (bool, *idr.Node, error) {
	for len(r.linesBuf) < decl.rows() {
		if err := r.readLine(); err != nil {
			if err != io.EOF || len(r.linesBuf) == 0 {
				// So either the err isn't io.EOF, we need to return this critical error directly;
				// or it's io.EOF and our line buf is empty, i.e. all has been procssed, so we
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

func (r *reader) readAndMatchHeaderFooterBasedEnvelope(
	decl *EnvelopeDecl, createNode bool) (bool, *idr.Node, error) {
	if len(r.linesBuf) <= 0 {
		if err := r.readLine(); err != nil {
			// io.EOF or not, since r.linesBuf is empty, we can directly return err.
			return false, nil, err
		}
	}
	if !decl.matchHeader(r.linesBuf[0].b) {
		return false, nil, nil
	}
	i := 0 // we'll match the footer starting on the same line.
	for {
		if decl.matchFooter(r.linesBuf[i].b) {
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
	for {
		// note1: ios.ByteReadLine returns a ln with trailing '\n' (and/or '\r') dropped.
		// note2: ios.ByteReadLine won't return io.EOF if ln returned isn't empty.
		b, err := ios.ByteReadLine(r.r)
		switch {
		case err == io.EOF:
			return io.EOF
		case err != nil:
			return ErrInvalidFixedLength(r.fmtErrStr(r.linesRead+1, err.Error()))
		}
		r.linesRead++
		if len(b) > 0 {
			r.linesBuf = append(r.linesBuf, line{lineNum: r.linesRead, b: b})
			return nil
		}
	}
}

func (r *reader) linesToNode(decl *EnvelopeDecl, n int) *idr.Node {
	if len(r.linesBuf) < n {
		panic(
			fmt.Sprintf("linesBuf has %d lines but requested %d lines to convert",
				len(r.linesBuf), n))
	}
	node := idr.CreateNode(idr.ElementNode, decl.Name)
	for col := range decl.Columns {
		colDecl := decl.Columns[col]
		for i := 0; i < n; i++ {
			if !colDecl.lineMatch(i, r.linesBuf[i].b) {
				continue
			}
			colNode := idr.CreateNode(idr.ElementNode, colDecl.Name)
			idr.AddChild(node, colNode)
			colVal := idr.CreateNode(idr.TextNode, colDecl.lineToColumnValue(r.linesBuf[i].b))
			idr.AddChild(colNode, colVal)
		}
	}
	return node
}

func (r *reader) popFrontLinesBuf(n int) {
	if n > len(r.linesBuf) {
		panic(fmt.Sprintf(
			"less lines (%d) in r.linesBuf than requested pop front count (%d)",
			len(r.linesBuf), n))
	}
	newLen := len(r.linesBuf) - n
	for i := 0; i < newLen; i++ {
		r.linesBuf[i] = r.linesBuf[i+n]
	}
	r.linesBuf = r.linesBuf[:newLen]
}

func (r *reader) unprocessedLineNum() int {
	if len(r.linesBuf) > 0 {
		return r.linesBuf[0].lineNum
	}
	return r.linesRead + 1
}

// Release implements fileformat.FormatReader interface, releasing a finished IDR target node.
func (r *reader) Release(n *idr.Node) {
	r.hr.Release(n)
}

// IsContinuableError implements fileformat..FormatReader interface, checking if an error is
// fatal or not.
func (r *reader) IsContinuableError(err error) bool {
	return !IsErrInvalidFixedLength(err) && err != io.EOF
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

// ErrInvalidFixedLength indicates the fixed-length content is corrupted or IO failure.
// This is a fatal, non-continuable error.
type ErrInvalidFixedLength string

// Error implements error interface.
func (e ErrInvalidFixedLength) Error() string { return string(e) }

// IsErrInvalidFixedLength checks if the `err` is of ErrInvalidFixedLength type.
func IsErrInvalidFixedLength(err error) bool {
	switch err.(type) {
	case ErrInvalidFixedLength:
		return true
	default:
		return false
	}
}

package fixedlength

import (
	"bufio"
	"fmt"
	"io"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/ios"

	"github.com/jf-tech/omniparser/idr"
)

type ErrInvalidEnvelope string

func (e ErrInvalidEnvelope) Error() string { return string(e) }

func IsErrInvalidEnvelope(err error) bool {
	switch err.(type) {
	case ErrInvalidEnvelope:
		return true
	default:
		return false
	}
}

type reader struct {
	inputName     string
	r             *bufio.Reader
	decl          *fileDecl
	xpath         *xpath.Expr
	root          *idr.Node
	target        *idr.Node
	envelopeIndex int
	line          int // 1-based
}

// Note the returned []byte is only valid before the next readLine() call.
func (r *reader) readLine() ([]byte, error) {
	for {
		line, err := ios.ByteReadLine(r.r)
		switch err {
		case nil:
			r.line++
		case io.EOF:
			return nil, err
		default:
			r.line++
			return nil, err
		}
		// skip only truly empty lines.
		if len(line) == 0 {
			continue
		}
		return line, nil
	}
}

func (r *reader) readByRowsEnvelope() (*idr.Node, error) {
	envelopeDecl := r.decl.Envelopes[r.envelopeIndex]
	node := idr.CreateNode(idr.ElementNode, *envelopeDecl.Name)
	columnsDone := make([]bool, len(envelopeDecl.Columns))
	for i := 0; i < envelopeDecl.byRows(); i++ {
		line, err := r.readLine()
		if err != nil {
			if err == io.EOF && i == 0 {
				return nil, err
			}
			return nil, ErrInvalidEnvelope(
				r.fmtErrStr("incomplete envelope, missing %d row(s)", envelopeDecl.byRows()-i))
		}
		for col := range envelopeDecl.Columns {
			if columnsDone[col] {
				continue
			}
			colDecl := envelopeDecl.Columns[col]
			if !colDecl.lineMatch(line) {
				continue
			}
			colNode := idr.CreateNode(idr.ElementNode, colDecl.Name)
			idr.AddChild(node, colNode)
			colVal := idr.CreateNode(idr.TextNode, colDecl.lineToColumnValue(line))
			idr.AddChild(colNode, colVal)
			columnsDone[col] = true
		}
	}
	return node, nil
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' line %d: %s", r.inputName, r.line, fmt.Sprintf(format, args...))
}

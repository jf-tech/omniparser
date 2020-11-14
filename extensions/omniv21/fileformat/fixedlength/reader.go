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
	envelopeLines [][]byte
}

const (
	EnvelopeLinesCap = 20
)

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

func (r *reader) readByRowsEnvelope() ([][]byte, error) {
	r.envelopeLines = r.envelopeLines[:0]
	copyNeeded := r.decl.Envelopes[r.envelopeIndex].byRows() > 1
	for i := 0; i < r.decl.Envelopes[r.envelopeIndex].byRows(); i++ {
		line, err := r.readLine()
		switch {
		case err == nil:
			if copyNeeded {
				cp := make([]byte, len(line))
				copy(cp, line)
				r.envelopeLines = append(r.envelopeLines, cp)
			} else {
				r.envelopeLines = append(r.envelopeLines, line)
			}
			continue
		case err == io.EOF && i == 0:
			return nil, err
		default:
			return nil, ErrInvalidEnvelope(r.fmtErrStr("envelope incomplete: %s", err.Error()))
		}
	}
	return r.envelopeLines, nil
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' line %d: %s", r.inputName, r.line, fmt.Sprintf(format, args...))
}

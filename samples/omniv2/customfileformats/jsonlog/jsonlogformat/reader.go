package jsonlogformat

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/idr"
)

// ErrLogReadingFailed indicates the reader fails to read out a complete non-corrupted
// log line. This is a fatal, non-continuable error.
type ErrLogReadingFailed string

func (e ErrLogReadingFailed) Error() string { return string(e) }

// IsErrLogReadingFailed checks if an err is of ErrLogReadingFailed type.
func IsErrLogReadingFailed(err error) bool {
	switch err.(type) {
	case ErrLogReadingFailed:
		return true
	default:
		return false
	}
}

type reader struct {
	inputName string
	r         *bufio.Reader
	line      int
	filter    *xpath.Expr
}

func (r *reader) Read() (*idr.Node, error) {
	for {
		r.line++
		l, err := ios.ReadLine(r.r)
		if err == io.EOF {
			return nil, errs.ErrEOF
		}
		if err != nil {
			// If we fail to read a log line out (permission issue, disk issue, whatever)
			// there is really no point to continue anymore, thus wrap the error in this
			// non-continuable error ErrLogReadingFailed.
			return nil, ErrLogReadingFailed(r.fmtErrStr(err.Error()))
		}
		if !strs.IsStrNonBlank(l) {
			continue
		}
		n, err := parseJSON([]byte(l))
		if err != nil {
			// If we read out a log line fine, but unable to parse it, that shouldn't be
			// a fatal error, thus not wrapping the error in non-continuable error
			// ErrLogReadingFailed.
			return nil, r.FmtErr(err.Error())
		}
		// Now we test this log-line-translated node (and its subtree) against the filter,
		// if no match, then we'll move onto the next line.
		if !idr.MatchAny(n, r.filter) {
			continue
		}
		return n, nil
	}
}

func parseJSON(b []byte) (*idr.Node, error) {
	p, err := idr.NewJSONStreamReader(bytes.NewReader(b), ".")
	if err != nil {
		return nil, err
	}
	return p.Read()
}

func (r *reader) IsContinuableError(err error) bool {
	return !IsErrLogReadingFailed(err) && err != errs.ErrEOF
}

func (r *reader) FmtErr(format string, args ...interface{}) error {
	return errors.New(r.fmtErrStr(format, args...))
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' line %d: %s", r.inputName, r.line, fmt.Sprintf(format, args...))
}

// NewReader creates an InputReader for this sample jsonlog file format.
func NewReader(inputName string, src io.Reader, filterXPath string) (*reader, error) {
	filter, err := caches.GetXPathExpr(filterXPath)
	if err != nil {
		return nil, err
	}
	return &reader{
		inputName: inputName,
		r:         bufio.NewReader(src),
		line:      0,
		filter:    filter,
	}, nil
}

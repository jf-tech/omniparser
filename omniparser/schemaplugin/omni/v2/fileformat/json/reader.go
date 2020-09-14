package omniv2json

import (
	"errors"
	"fmt"
	"io"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/nodes"
)

// ErrNodeReadingFailed indicates the reader fails to read out a complete non-corrupted
// JSON node. This is a fatal, non-continuable error.
type ErrNodeReadingFailed string

func (e ErrNodeReadingFailed) Error() string { return string(e) }

// IsErrNodeReadingFailed checks if an err is of ErrNodeReadingFailed type.
func IsErrNodeReadingFailed(err error) bool {
	switch err.(type) {
	case ErrNodeReadingFailed:
		return true
	default:
		return false
	}
}

type reader struct {
	inputName string
	r         *nodes.JSONStreamParser
}

func (r *reader) Read() (*node.Node, error) {
	n, err := r.r.Read()
	if err == io.EOF {
		return nil, errs.ErrEOF
	}
	if err != nil {
		return nil, ErrNodeReadingFailed(r.fmtErrStr(err.Error()))
	}
	return n, nil
}

func (r *reader) IsContinuableError(err error) bool {
	return !IsErrNodeReadingFailed(err) && err != errs.ErrEOF
}

func (r *reader) FmtErr(format string, args ...interface{}) error {
	return errors.New(r.fmtErrStr(format, args...))
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' before/near line %d: %s", r.inputName, r.r.AtLine(), fmt.Sprintf(format, args...))
}

// NewReader creates an InputReader for JSON file format for omniv2 schema plugin.
func NewReader(inputName string, src io.Reader, xpath string) (*reader, error) {
	sp, err := nodes.NewJSONStreamParser(src, xpath)
	if err != nil {
		return nil, err
	}
	return &reader{inputName: inputName, r: sp}, nil
}

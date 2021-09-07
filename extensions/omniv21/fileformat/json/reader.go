package json

import (
	"errors"
	"fmt"
	"io"

	"github.com/jf-tech/omniparser/idr"
)

// ErrNodeReadingFailed indicates the reader fails to read out a complete non-corrupted
// JSON node. This is a fatal, non-continuable error.
type ErrNodeReadingFailed string

func (e ErrNodeReadingFailed) Error() string { return string(e) }

// IsErrNodeReadingFailed checks if the `err` is of ErrNodeReadingFailed type.
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
	r         *idr.JSONStreamReader
}

func (r *reader) Read() (*idr.Node, error) {
	n, err := r.r.Read()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, ErrNodeReadingFailed(r.fmtErrStr(err.Error()))
	}
	return n, nil
}

func (r *reader) Release(n *idr.Node) {
	if n != nil {
		r.r.Release(n)
	}
}

func (r *reader) IsContinuableError(err error) bool {
	return !IsErrNodeReadingFailed(err) && err != io.EOF
}

func (r *reader) FmtErr(format string, args ...interface{}) error {
	return errors.New(r.fmtErrStr(format, args...))
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' before/near line %d: %s", r.inputName, r.r.AtLine(), fmt.Sprintf(format, args...))
}

// NewReader creates an FormatReader for JSON file format.
func NewReader(inputName string, src io.Reader, xpath string) (*reader, error) {
	sp, err := idr.NewJSONStreamReader(src, xpath)
	if err != nil {
		return nil, err
	}
	return &reader{inputName: inputName, r: sp}, nil
}

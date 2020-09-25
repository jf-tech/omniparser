package omniv2xml

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/nodes"
)

// ErrNodeReadingFailed indicates the reader fails to read out a complete non-corrupted
// XML element node. This is a fatal, non-continuable error.
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
	reader    *node.StreamParser
}

func (r *reader) Read() (*node.Node, error) {
	n, err := r.reader.Read()
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
	return fmt.Sprintf("input '%s' near line %d: %s", r.inputName, r.lineNumber(), fmt.Sprintf(format, args...))
}

func (r *reader) lineNumber() int {
	// We want to return an approx line number for error reporting purpose. But
	// the 'line' field is buried at: r.reader.p.decoder.line, and none of them
	// are exported. So using this hack to get the number. Given all the libraries
	// are of fixed versions in go modules, we're fine. If in the future, something
	// changes and breaks due to library upgrade, we'll have test failures to remind
	// us to fix.
	return int(reflect.ValueOf(r.reader).Elem().
		FieldByName("p").Elem().
		FieldByName("decoder").Elem().
		FieldByName("line").Int())
}

// NewReader creates an InputReader for XML file format for omniv2 schema handler.
func NewReader(inputName string, src io.Reader, xpath string) (*reader, error) {
	xpath = strings.TrimSpace(xpath)
	xpathWithoutLastFilter := nodes.RemoveLastFilterInXPath(xpath)
	var sp *node.StreamParser
	var err error
	if xpathWithoutLastFilter == xpath {
		sp, err = node.CreateStreamParser(src, xpath)
	} else {
		sp, err = node.CreateStreamParser(src, xpathWithoutLastFilter, xpath)
	}
	if err != nil {
		return nil, err
	}
	return &reader{inputName: inputName, reader: sp}, nil
}

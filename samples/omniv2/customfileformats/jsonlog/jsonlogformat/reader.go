package jsonlogformat

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"

	node "github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/jf-tech/iohelper"

	"github.com/jf-tech/omniparser/cache"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/strs"
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

func (r *reader) Read() (*node.Node, error) {
	for {
		r.line++
		l, err := iohelper.ReadLine(r.r)
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
		if node.QuerySelector(n, r.filter) == nil {
			continue
		}
		return n, nil
	}
}

// Note parseJSON and parseJSONValue are borrowed and adapted from
// https://github.com/antchfx/jsonquery/blob/master/node.go.

func parseJSONValue(x interface{}, parent *node.Node) {
	switch v := x.(type) {
	case []interface{}:
		for _, vv := range v {
			n := &node.Node{Type: node.ElementNode}
			node.AddChild(parent, n)
			parseJSONValue(vv, n)
		}
	case map[string]interface{}:
		var keys []string
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			n := &node.Node{Data: key, Type: node.ElementNode}
			node.AddChild(parent, n)
			parseJSONValue(v[key], n)
		}
	case string:
		node.AddChild(parent, &node.Node{Data: v, Type: node.TextNode})
	case float64:
		// The format fmt with 'f' means (-ddd.dddd, no exponent),
		// The special precision -1 uses the smallest number of digits
		s := strconv.FormatFloat(v, 'f', -1, 64)
		node.AddChild(parent, &node.Node{Data: s, Type: node.TextNode})
	case bool:
		s := strconv.FormatBool(v)
		node.AddChild(parent, &node.Node{Data: s, Type: node.TextNode})
	}
}

func parseJSON(b []byte) (*node.Node, error) {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	doc := &node.Node{Type: node.DocumentNode}
	parseJSONValue(v, doc)
	return doc, nil
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
	filter, err := cache.GetXPathExpr(filterXPath)
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

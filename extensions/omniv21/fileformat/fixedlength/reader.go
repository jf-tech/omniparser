package fixedlength

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"

	"github.com/jf-tech/omniparser/idr"
)

// ErrInvalidEnvelope indicates a fixed-length input envelope is invalid. This is a fatal, non-continuable error.
type ErrInvalidEnvelope string

func (e ErrInvalidEnvelope) Error() string { return string(e) }

// IsErrInvalidEnvelope checks if an error is of ErrInvalidEnvelope type.
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
	decl          *FileDecl
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
		default:
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

func (r *reader) readByHeaderFooterEnvelope() (*idr.Node, error) {
	line, err := r.readLine()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, ErrInvalidEnvelope(r.fmtErrStr("incomplete envelope: %s", err.Error()))
	}
	for ; r.envelopeIndex < len(r.decl.Envelopes); r.envelopeIndex++ {
		// regex is already validated
		headerRegex, _ := caches.GetRegex(r.decl.Envelopes[r.envelopeIndex].ByHeaderFooter.Header)
		if headerRegex.Match(line) {
			break
		}
	}
	if r.envelopeIndex >= len(r.decl.Envelopes) {
		return nil, io.EOF
	}
	envelopeDecl := r.decl.Envelopes[r.envelopeIndex]
	footerRegex, _ := caches.GetRegex(envelopeDecl.ByHeaderFooter.Footer)
	node := idr.CreateNode(idr.ElementNode, *envelopeDecl.Name)
	columnsDone := make([]bool, len(envelopeDecl.Columns))
	for {
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
		if footerRegex.Match(line) {
			return node, nil
		}
		line, err = r.readLine()
		// Since the envelope has started, any reading error, including EOF, indicates incomplete envelope error.
		if err != nil {
			return nil, ErrInvalidEnvelope(r.fmtErrStr("incomplete envelope: %s", err.Error()))
		}
	}
}

func (r *reader) Read() (node *idr.Node, err error) {
	if r.target != nil {
		// This is just in case Release() isn't called by ingester.
		idr.RemoveAndReleaseTree(r.target)
		r.target = nil
	}
readEnvelope:
	if r.decl.envelopeType() == envelopeTypeByRows {
		node, err = r.readByRowsEnvelope()
		if err != nil {
			return nil, err
		}
		idr.AddChild(r.root, node)
	} else {
		node, err = r.readByHeaderFooterEnvelope()
		if err != nil {
			return nil, err
		}
		idr.AddChild(r.root, node)
		if r.decl.Envelopes[r.envelopeIndex].NotTarget {
			// If this by_header_footer envelope isn't target envelope then we consider it
			// a global envelope and keep it in the idr tree.
			goto readEnvelope
		}
	}
	// now the envelope is the target envelope, let's do a target xpath filtering.
	// if it filters out, then we need to remove it from the idr tree.
	if r.xpath != nil && !idr.MatchAny(node, r.xpath) {
		idr.RemoveAndReleaseTree(node)
		goto readEnvelope
	}
	r.target = node
	return node, err
}

func (r *reader) Release(n *idr.Node) {
	if r.target == n {
		r.target = nil
	}
	idr.RemoveAndReleaseTree(n)
}

func (r *reader) IsContinuableError(err error) bool {
	return !IsErrInvalidEnvelope(err) && err != io.EOF
}

func (r *reader) FmtErr(format string, args ...interface{}) error {
	return errors.New(r.fmtErrStr(format, args...))
}

func (r *reader) fmtErrStr(format string, args ...interface{}) string {
	return fmt.Sprintf("input '%s' line %d: %s", r.inputName, r.line, fmt.Sprintf(format, args...))
}

// NewReader creates an FormatReader for fixed-length file format.
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
	return &reader{
		inputName: inputName,
		r:         bufio.NewReader(r),
		decl:      decl,
		xpath:     expr,
		root:      idr.CreateNode(idr.DocumentNode, "#root"),
		line:      1,
	}, nil
}

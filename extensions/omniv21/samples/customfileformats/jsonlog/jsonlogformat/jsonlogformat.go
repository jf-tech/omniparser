package jsonlogformat

import (
	"fmt"
	"io"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/header"
)

const (
	// FileFormatJSONLog is the constant for this special/sample file format.
	FileFormatJSONLog = "jsonlog"
)

type jsonLogFileFormat struct {
	schemaName string
}

// NewJSONLogFileFormat creates a new FileFormat for this special/sample jsonlog.
func NewJSONLogFileFormat(schemaName string) fileformat.FileFormat {
	return &jsonLogFileFormat{schemaName: schemaName}
}

func (p *jsonLogFileFormat) ValidateSchema(
	format string, _ []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != FileFormatJSONLog {
		return nil, errs.ErrSchemaNotSupported
	}
	if finalOutputDecl == nil {
		return nil, p.FmtErr("'FINAL_OUTPUT' decl is nil")
	}
	if !strs.IsStrPtrNonBlank(finalOutputDecl.XPath) {
		return nil, p.FmtErr("'FINAL_OUTPUT' must have 'xpath' specified")
	}
	_, err := caches.GetXPathExpr(*finalOutputDecl.XPath)
	if err != nil {
		return nil, p.FmtErr("'xpath' on 'FINAL_OUTPUT' (value: '%s') is invalid, err: %s",
			*finalOutputDecl.XPath, err.Error())
	}
	return *finalOutputDecl.XPath, nil
}

func (p *jsonLogFileFormat) CreateFormatReader(
	_ header.Header,
	name string,
	r io.Reader,
	runtime interface{}) (fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (p *jsonLogFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", p.schemaName, fmt.Sprintf(format, args...))
}

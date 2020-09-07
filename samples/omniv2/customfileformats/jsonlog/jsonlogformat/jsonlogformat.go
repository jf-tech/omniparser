package jsonlogformat

import (
	"fmt"
	"io"

	"github.com/antchfx/xpath"

	"github.com/jf-tech/omniparser/omniparser/errs"
	omniv2fileformat "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/strs"
)

const (
	FileFormatJSONLog = "jsonlog"
)

type jsonLogFileFormat struct {
	schemaName string
}

func NewJSONLogFileFormat(schemaName string) omniv2fileformat.FileFormat {
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
	_, err := xpath.Compile(*finalOutputDecl.XPath)
	if err != nil {
		return nil, p.FmtErr("'xpath' on 'FINAL_OUTPUT' (value: '%s') is invalid, err: %s",
			*finalOutputDecl.XPath, err.Error())
	}
	return *finalOutputDecl.XPath, nil
}

func (p *jsonLogFileFormat) CreateFormatReader(
	name string, r io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (p *jsonLogFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", p.schemaName, fmt.Sprintf(format, args...))
}

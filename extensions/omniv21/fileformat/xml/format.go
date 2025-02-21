package xml

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
	fileFormatXML = "xml"
)

type xmlFileFormat struct {
	schemaName string
}

// NewXMLFileFormat creates a FileFormat for XML.
func NewXMLFileFormat(schemaName string) fileformat.FileFormat {
	return &xmlFileFormat{schemaName: schemaName}
}

func (f *xmlFileFormat) ValidateSchema(format string, _ []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != fileFormatXML {
		return nil, errs.ErrSchemaNotSupported
	}
	if finalOutputDecl == nil {
		return nil, f.FmtErr("'FINAL_OUTPUT' is missing")
	}
	xpath := strs.StrPtrOrElse(finalOutputDecl.XPath, ".")
	_, err := caches.GetXPathExpr(xpath)
	if err != nil {
		return nil, f.FmtErr("'FINAL_OUTPUT.xpath' (value: '%s') is invalid, err: %s", xpath, err.Error())
	}
	return xpath, nil
}

func (f *xmlFileFormat) CreateFormatReader(
	_ header.Header,
	name string,
	r io.Reader,
	runtime interface{}) (fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (f *xmlFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

package omniv2xml

import (
	"fmt"
	"io"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/errs"
	omniv2fileformat "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
)

const (
	// fileFormatXML is the file format for XML for omniv2 schema handler.
	fileFormatXML = "xml"
)

type xmlFileFormat struct {
	schemaName string
}

// NewXMLFileFormat creates a FileFormat for XML for omniv2 schema handler.
func NewXMLFileFormat(schemaName string) omniv2fileformat.FileFormat {
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
	name string, r io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (f *xmlFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

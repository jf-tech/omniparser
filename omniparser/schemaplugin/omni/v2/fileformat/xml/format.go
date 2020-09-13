package omniv2xml

import (
	"fmt"
	"io"

	"github.com/jf-tech/omniparser/cache"
	"github.com/jf-tech/omniparser/omniparser/errs"
	omniv2fileformat "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/strs"
)

const (
	// FileFormatXML is the file format for XML for omniv2 schema plugin.
	FileFormatXML = "xml"
)

type xmlFileFormat struct {
	schemaName string
}

// NewXMLFileFormat creates a FileFormat for XML for omniv2 schema plugin.
func NewXMLFileFormat(schemaName string) omniv2fileformat.FileFormat {
	return &xmlFileFormat{schemaName: schemaName}
}

func (p *xmlFileFormat) ValidateSchema(format string, _ []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != FileFormatXML {
		return nil, errs.ErrSchemaNotSupported
	}
	if finalOutputDecl == nil {
		return nil, p.FmtErr("'FINAL_OUTPUT' decl is nil")
	}
	if !strs.IsStrPtrNonBlank(finalOutputDecl.XPath) {
		return nil, p.FmtErr("'FINAL_OUTPUT' must have 'xpath' specified")
	}
	_, err := cache.GetXPathExpr(*finalOutputDecl.XPath)
	if err != nil {
		return nil, p.FmtErr("'xpath' on 'FINAL_OUTPUT' (value: '%s') is invalid, err: %s",
			*finalOutputDecl.XPath, err.Error())
	}
	return *finalOutputDecl.XPath, nil
}

func (p *xmlFileFormat) CreateFormatReader(
	name string, r io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (p *xmlFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", p.schemaName, fmt.Sprintf(format, args...))
}

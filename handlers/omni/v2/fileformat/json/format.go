package omniv2json

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
	fileFormatJSON = "json"
)

type jsonFileFormat struct {
	schemaName string
}

// NewJSONFileFormat creates a FileFormat for JSON for omniv2 schema handler.
func NewJSONFileFormat(schemaName string) omniv2fileformat.FileFormat {
	return &jsonFileFormat{schemaName: schemaName}
}

func (f *jsonFileFormat) ValidateSchema(format string, _ []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != fileFormatJSON {
		return nil, errs.ErrSchemaNotSupported
	}
	if finalOutputDecl == nil {
		return nil, f.FmtErr("'FINAL_OUTPUT' is missing")
	}
	if !strs.IsStrPtrNonBlank(finalOutputDecl.XPath) {
		return nil, f.FmtErr("'FINAL_OUTPUT' must have 'xpath' specified")
	}
	_, err := caches.GetXPathExpr(*finalOutputDecl.XPath)
	if err != nil {
		return nil, f.FmtErr("'FINAL_OUTPUT.xpath' (value: '%s') is invalid, err: %s",
			*finalOutputDecl.XPath, err.Error())
	}
	return *finalOutputDecl.XPath, nil
}

func (f *jsonFileFormat) CreateFormatReader(
	name string, r io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (f *jsonFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

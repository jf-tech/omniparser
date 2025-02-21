package json

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
	fileFormatJSON = "json"
)

type jsonFileFormat struct {
	schemaName string
}

// NewJSONFileFormat creates a FileFormat for JSON.
func NewJSONFileFormat(schemaName string) fileformat.FileFormat {
	return &jsonFileFormat{schemaName: schemaName}
}

func (f *jsonFileFormat) ValidateSchema(format string, _ []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != fileFormatJSON {
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

func (f *jsonFileFormat) CreateFormatReader(
	_ header.Header,
	name string,
	r io.Reader,
	runtime interface{}) (fileformat.FormatReader, error) {
	return NewReader(name, r, runtime.(string))
}

func (f *jsonFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

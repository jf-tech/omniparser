package csv

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	v21validation "github.com/jf-tech/omniparser/extensions/omniv21/validation"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/validation"
)

const (
	fileFormatCSV = "csv2"
)

type csvFormat struct {
	schemaName string
}

// NewCSVFileFormat creates a FileFormat for 'csv2'.
func NewCSVFileFormat(schemaName string) fileformat.FileFormat {
	return &csvFormat{schemaName: schemaName}
}

type csvFormatRuntime struct {
	Decl  *FileDecl `json:"file_declaration"`
	XPath string
}

func (f *csvFormat) ValidateSchema(
	format string, schemaContent []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != fileFormatCSV {
		return nil, errs.ErrSchemaNotSupported
	}
	err := validation.SchemaValidate(
		f.schemaName, schemaContent, v21validation.JSONSchemaCSV2FileDeclaration)
	if err != nil {
		// err is already context formatted.
		return nil, err
	}
	var runtime csvFormatRuntime
	_ = json.Unmarshal(schemaContent, &runtime) // JSON schema validation earlier guarantees Unmarshal success.
	err = f.validateFileDecl(runtime.Decl)
	if err != nil {
		// err is already context formatted.
		return nil, err
	}
	if finalOutputDecl == nil {
		return nil, f.FmtErr("'FINAL_OUTPUT' is missing")
	}
	runtime.XPath = strings.TrimSpace(strs.StrPtrOrElse(finalOutputDecl.XPath, ""))
	if runtime.XPath != "" {
		_, err := caches.GetXPathExpr(runtime.XPath)
		if err != nil {
			return nil, f.FmtErr("'FINAL_OUTPUT.xpath' (value: '%s') is invalid, err: %s",
				runtime.XPath, err.Error())
		}
	}
	return &runtime, nil
}

func (f *csvFormat) validateFileDecl(decl *FileDecl) error {
	err := (&validateCtx{}).validateFileDecl(decl)
	if err != nil {
		return f.FmtErr(err.Error())
	}
	return err
}

func (f *csvFormat) CreateFormatReader(
	schemaHeader header.Header,
	name string,
	r io.Reader,
	runtime interface{}) (fileformat.FormatReader, error) {
	rt := runtime.(*csvFormatRuntime)
	targetXPathExpr, err := func() (*xpath.Expr, error) {
		if rt.XPath == "" || rt.XPath == "." {
			return nil, nil
		}
		return caches.GetXPathExpr(rt.XPath)
	}()
	if err != nil {
		return nil, f.FmtErr("xpath '%s' on 'FINAL_OUTPUT' is invalid: %s", rt.XPath, err.Error())
	}
	return NewReader(schemaHeader, name, r, rt.Decl, targetXPathExpr), nil
}

func (f *csvFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

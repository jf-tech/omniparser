package omniv2csv

import (
	"encoding/json"
	"fmt"
	"github.com/jf-tech/omniparser/validation"
	"io"
	"strings"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/errs"
	omniv2fileformat "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
)

const (
	fileFormatCSV       = "csv"
	fileFormatDelimited = "delimited"
)

type csvFileFormat struct {
	schemaName string
}

// NewCSVFileFormat creates a FileFormat for CSV for omniv2 schema handler.
func NewCSVFileFormat(schemaName string) omniv2fileformat.FileFormat {
	return &csvFileFormat{schemaName: schemaName}
}

type csvFormatRuntime struct {
	Decl  *fileDecl `json:"file_declaration"`
	XPath string
}

func (f *csvFileFormat) ValidateSchema(
	format string, schemaContent []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != fileFormatCSV && format != fileFormatDelimited {
		return nil, errs.ErrSchemaNotSupported
	}
	err := validation.SchemaValidate(f.schemaName, schemaContent, validation.JSONSchemaCSVFileDeclaration)
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

func (f *csvFileFormat) validateFileDecl(decl *fileDecl) error {
	// If header_row_index is specified, then it must be < data_row_index
	if decl.HeaderRowIndex != nil && *decl.HeaderRowIndex >= decl.DataRowIndex {
		return f.FmtErr(
			"file_declaration.header_row_index(%d) must be smaller than file_declaration.data_row_index(%d)",
			*decl.HeaderRowIndex, decl.DataRowIndex)
	}
	if err := f.validateColumns(decl.Columns); err != nil {
		return err
	}
	return nil
}

func (f *csvFileFormat) validateColumns(columns []column) error {
	namesSeen := map[string]bool{}
	aliasesSeen := map[string]bool{}
	for _, column := range columns {
		if _, found := namesSeen[column.Name]; found {
			return f.FmtErr("file_declaration.columns contains duplicate name '%s'", column.Name)
		}
		namesSeen[column.Name] = true
		if column.Alias != nil {
			if _, found := aliasesSeen[*column.Alias]; found {
				return f.FmtErr("file_declaration.columns contains duplicate alias '%s'", *column.Alias)
			}
			aliasesSeen[*column.Alias] = true
		}
	}
	return nil
}

func (f *csvFileFormat) CreateFormatReader(
	name string, r io.Reader, runtime interface{}) (omniv2fileformat.FormatReader, error) {
	csv := runtime.(*csvFormatRuntime)
	return NewReader(name, r, csv.Decl, csv.XPath)
}

func (f *csvFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

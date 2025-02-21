package fixedlength

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	fileFormatFixedLength = "fixed-length"
)

type fixedLengthFileFormat struct {
	schemaName               string
	autoGenEnvelopeNameIndex int
}

// NewFixedLengthFileFormat creates a FileFormat for fixed-length files.
func NewFixedLengthFileFormat(schemaName string) fileformat.FileFormat {
	return &fixedLengthFileFormat{schemaName: schemaName}
}

type fixedLengthFormatRuntime struct {
	Decl  *FileDecl `json:"file_declaration"`
	XPath string
}

func (f *fixedLengthFileFormat) ValidateSchema(
	format string, schemaContent []byte, finalOutputDecl *transform.Decl) (interface{}, error) {
	if format != fileFormatFixedLength {
		return nil, errs.ErrSchemaNotSupported
	}
	err := validation.SchemaValidate(f.schemaName, schemaContent, v21validation.JSONSchemaFixedLengthFileDeclaration)
	if err != nil {
		// err is already context formatted.
		return nil, err
	}
	var runtime fixedLengthFormatRuntime
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

func (f *fixedLengthFileFormat) validateFileDecl(decl *FileDecl) error {
	targetSeen := false
	namesSeen := map[string]bool{}
	for _, envelope := range decl.Envelopes {
		if targetSeen && !envelope.NotTarget {
			return f.FmtErr("cannot have more than one target envelope")
		}
		targetSeen = targetSeen || !envelope.NotTarget
		if envelope.Name == nil {
			f.autoGenEnvelopeNameIndex++
			envelope.Name = strs.StrPtr(strconv.Itoa(f.autoGenEnvelopeNameIndex))
		}
		if _, found := namesSeen[*envelope.Name]; found {
			return f.FmtErr("more than one envelope has the name '%s'", *envelope.Name)
		}
		namesSeen[*envelope.Name] = true
		if err := f.validateByHeaderFooter(envelope.ByHeaderFooter); err != nil {
			return err
		}
		if err := f.validateColumns(envelope.Columns); err != nil {
			return err
		}
	}
	if !targetSeen {
		return f.FmtErr("missing target envelope")
	}
	return nil
}

func (f *fixedLengthFileFormat) validateByHeaderFooter(decl *ByHeaderFooterDecl) error {
	if decl == nil {
		return nil
	}
	_, err := caches.GetRegex(decl.Header)
	if err != nil {
		return f.FmtErr("invalid 'header' regex '%s': %s", decl.Header, err.Error())
	}
	_, err = caches.GetRegex(decl.Footer)
	if err != nil {
		return f.FmtErr("invalid 'footer' regex '%s': %s", decl.Footer, err.Error())
	}
	return nil
}

func (f *fixedLengthFileFormat) validateColumns(cols []*ColumnDecl) error {
	columnNamesSeen := map[string]bool{}
	for _, col := range cols {
		if _, found := columnNamesSeen[col.Name]; found {
			return f.FmtErr("more than one column has the name '%s'", col.Name)
		}
		columnNamesSeen[col.Name] = true
		if col.LinePattern != nil {
			if _, err := caches.GetRegex(*col.LinePattern); err != nil {
				return f.FmtErr("invalid 'line_pattern' regex '%s': %s", *col.LinePattern, err.Error())
			}
		}
	}
	return nil
}

func (f *fixedLengthFileFormat) CreateFormatReader(
	_ header.Header,
	name string,
	r io.Reader,
	runtime interface{}) (fileformat.FormatReader, error) {
	rt := runtime.(*fixedLengthFormatRuntime)
	return NewReader(name, r, rt.Decl, rt.XPath)
}

func (f *fixedLengthFileFormat) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("schema '%s': %s", f.schemaName, fmt.Sprintf(format, args...))
}

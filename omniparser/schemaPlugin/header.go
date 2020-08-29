package schemaPlugin

import (
	"io"

	"golang.org/x/text/encoding/charmap"

	"github.com/jf-tech/omniparser/strs"
)

// ParserSettings defines the common header (and its JSON format) for all schemas across all schema plugins.
// It contains vital information about what schema plugin a schema wants to use, and what file format the
// input stream is of (e.g. fixed-length txt, CSV/TSV, XML, JSON, EDI, etc).
// Also optionally, it specifies the expected the encoding scheme for the input streams this schema is used
// for.
type ParserSettings struct {
	Version        string  `json:"version,omitempty"`
	FileFormatType string  `json:"file_format_type,omitempty"`
	Encoding       *string `json:"encoding,omitempty"`
}

const (
	// EncodingUTF8 is the UTF-8 (golang's default) encoding scheme.
	EncodingUTF8 = "utf-8"
	// EncodingISO8859_1 is the ISO 8859-1 encoding.
	EncodingISO8859_1 = "iso-8859-1"
	// EncodingWindows1252 is the Windows 1252 encoding.
	EncodingWindows1252 = "windows-1252"
)

type encodingMappingFunc func(reader io.Reader) io.Reader

// SupportedEncodingMappings provides mapping between input stream reader and a func that does
// encoding specific translation.
var SupportedEncodingMappings = map[string]encodingMappingFunc{
	EncodingUTF8:        func(r io.Reader) io.Reader { return r },
	EncodingISO8859_1:   func(r io.Reader) io.Reader { return charmap.ISO8859_1.NewDecoder().Reader(r) },
	EncodingWindows1252: func(r io.Reader) io.Reader { return charmap.Windows1252.NewDecoder().Reader(r) },
}

// GetEncoding returns the encoding of the schema. If no encoding is specified in the schema, which
// the most comment default case, it assumes the input stream will be in UTF-8.
func (p ParserSettings) GetEncoding() string {
	return strs.StrPtrOrElse(p.Encoding, EncodingUTF8)
}

// Header contains the common ParserSettings for all schemas.
type Header struct {
	ParserSettings ParserSettings `json:"parser_settings,omitempty"`
}

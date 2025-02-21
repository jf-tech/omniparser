package header

import (
	"io"

	"github.com/jf-tech/go-corelib/strs"
	"golang.org/x/text/encoding/charmap"
)

// ParserSettings defines the common header (and its JSON format) for all schemas across all schema handlers.
// It contains vital information about which handler a schema wants to use, and what file format the input
// stream is of (e.g. fixed-length txt, CSV/TSV, XML, JSON, EDI, etc). Optionally, it specifies the expected
// encoding scheme for the input streams this schema is used for, as well as other auxiliary but global
// parser settings.
type ParserSettings struct {
	Version        string  `json:"version,omitempty"`
	FileFormatType string  `json:"file_format_type,omitempty"`
	Encoding       *string `json:"encoding,omitempty"`
	Debug          int     `json:"debug,omitempty"`
}

const (
	encodingUTF8        = "utf-8"
	encodingISO8859_1   = "iso-8859-1"
	encodingWindows1252 = "windows-1252"
)

type encodingMappingFunc func(reader io.Reader) io.Reader

var supportedEncodingMappings = map[string]encodingMappingFunc{
	encodingUTF8:        func(r io.Reader) io.Reader { return r },
	encodingISO8859_1:   func(r io.Reader) io.Reader { return charmap.ISO8859_1.NewDecoder().Reader(r) },
	encodingWindows1252: func(r io.Reader) io.Reader { return charmap.Windows1252.NewDecoder().Reader(r) },
}

// WrapEncoding returns an io.Reader that ensures the encoding scheme matches what's specified
// in 'parser_settings.encoding' setting.
func (p ParserSettings) WrapEncoding(input io.Reader) io.Reader {
	f, found := supportedEncodingMappings[strs.StrPtrOrElse(p.Encoding, encodingUTF8)]
	if !found {
		f = supportedEncodingMappings[encodingUTF8]
	}
	return f(input)
}

// Header contains the common ParserSettings for all schemas.
type Header struct {
	ParserSettings ParserSettings `json:"parser_settings,omitempty"`
}

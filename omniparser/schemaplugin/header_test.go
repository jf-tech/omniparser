package schemaplugin

import (
	"bytes"
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/charmap"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/testlib"
)

func TestSupportedEncodingMappingsDump(t *testing.T) {
	var supported []string
	for k := range supportedEncodingMappings {
		supported = append(supported, k)
	}
	sort.Strings(supported)
	cupaloy.SnapshotT(t, jsons.BPM(supported))
}

func TestSupportedEncodingMappings(t *testing.T) {
	for encoding, mappingFn := range supportedEncodingMappings {
		t.Run(encoding, func(t *testing.T) {
			actual, err := ioutil.ReadAll(mappingFn(strings.NewReader("test")))
			assert.NoError(t, err)
			assert.Equal(t, []byte("test"), actual)
		})
	}
}

func TestWrapEncoding(t *testing.T) {
	readAll := func(r io.Reader) string {
		b, err := ioutil.ReadAll(r)
		assert.NoError(t, err)
		return string(b)
	}
	// No 'parser_settings.encoding' ==> UTF-8
	assert.Equal(t, "test", readAll(ParserSettings{}.WrapEncoding(strings.NewReader("test"))))
	// 'parser_settings.encoding' = UTF-8
	assert.Equal(t, "test", readAll(
		ParserSettings{Encoding: testlib.StrPtr(encodingUTF8)}.WrapEncoding(strings.NewReader("test"))))
	// 'parser_settings.encoding' = <unknown> ==> UTF-8
	assert.Equal(t, "test", readAll(
		ParserSettings{Encoding: testlib.StrPtr("unknown")}.WrapEncoding(strings.NewReader("test"))))
	// 'parser_settings.encoding' = ISO-8859-1
	iso88591bytes, err := charmap.ISO8859_1.NewEncoder().Bytes([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "test", readAll(
		ParserSettings{Encoding: testlib.StrPtr(encodingISO8859_1)}.WrapEncoding(bytes.NewReader(iso88591bytes))))
	// 'parser_settings.encoding' = windows-1252
	windows1252bytes, err := charmap.Windows1252.NewEncoder().Bytes([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "test", readAll(
		ParserSettings{Encoding: testlib.StrPtr(encodingWindows1252)}.WrapEncoding(bytes.NewReader(windows1252bytes))))
}

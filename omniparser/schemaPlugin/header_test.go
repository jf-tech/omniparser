package schemaPlugin

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/testlib"
)

func TestSupportedEncodingMappingsDump(t *testing.T) {
	var supported []string
	for k, _ := range SupportedEncodingMappings {
		supported = append(supported, k)
	}
	sort.Strings(supported)
	cupaloy.SnapshotT(t, jsons.BPM(supported))
}

func TestSupportedEncodingMappings(t *testing.T) {
	for encoding, mappingFn := range SupportedEncodingMappings {
		t.Run(encoding, func(t *testing.T) {
			actual, err := ioutil.ReadAll(mappingFn(strings.NewReader("test")))
			assert.NoError(t, err)
			assert.Equal(t, []byte("test"), actual)
		})
	}
}

func TestGetEncoding(t *testing.T) {
	assert.Equal(
		t, EncodingUTF8, (ParserSettings{Encoding: testlib.StrPtr(EncodingUTF8)}).GetEncoding())
	assert.Equal(
		t, EncodingISO8859_1, (ParserSettings{Encoding: testlib.StrPtr(EncodingISO8859_1)}).GetEncoding())
	assert.Equal(
		t, EncodingWindows1252, (ParserSettings{Encoding: testlib.StrPtr(EncodingWindows1252)}).GetEncoding())
	assert.Equal(
		t, EncodingUTF8, (ParserSettings{}).GetEncoding())
	assert.Equal(
		t, "whatever", (ParserSettings{Encoding: testlib.StrPtr("whatever")}).GetEncoding())
}

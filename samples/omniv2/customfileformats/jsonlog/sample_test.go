package jsonlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/customfuncs"
	omniv2 "github.com/jf-tech/omniparser/handlers/omni/v2"
	omniv2fileformat "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/samples/omniv2/customfileformats/jsonlog/jsonlogformat"
	"github.com/jf-tech/omniparser/transformctx"
)

func normalizeSeverity(_ *transformctx.Ctx, sev string) (string, error) {
	switch strings.ToUpper(sev) {
	case "DEBUG":
		return "D", nil
	case "INFO":
		return "I", nil
	case "WARNING":
		return "W", nil
	case "ERROR":
		return "E", nil
	case "CRITICAL":
		return "F", nil
	default:
		return "?", nil
	}
}

func TestSample(t *testing.T) {
	schemaFile := "./sample_schema.json"
	schemaFileBaseName := filepath.Base(schemaFile)
	schemaFileReader, err := os.Open(schemaFile)
	assert.NoError(t, err)
	defer schemaFileReader.Close()

	inputFile := "./sample.log"
	inputFileBaseName := filepath.Base(inputFile)
	inputFileReader, err := os.Open(inputFile)
	assert.NoError(t, err)
	defer inputFileReader.Close()

	schema, err := omniparser.NewSchema(
		schemaFileBaseName,
		schemaFileReader,
		// Use this Extension to effectively replace the
		// builtin omniv2 schema handler (and its builtin fileformats)
		// with, well, omniv2 schema handler :) with our own custom
		// fileformat. Also let's demo how to add a new custom func.
		omniparser.Extension{
			CreateHandler: omniv2.CreateHandler, // Use the same omniv2 handler
			HandlerParams: &omniv2.HandlerParams{
				// But use our own FileFormat.
				CustomFileFormats: []omniv2fileformat.FileFormat{
					jsonlogformat.NewJSONLogFileFormat(schemaFileBaseName),
				},
			},
			CustomFuncs: customfuncs.CustomFuncs{
				"normalize_severity": normalizeSeverity,
			},
		})
	assert.NoError(t, err)
	transform, err := schema.NewTransform(inputFileBaseName, inputFileReader, &transformctx.Ctx{})
	assert.NoError(t, err)

	var records []string
	for transform.Next() {
		recordBytes, err := transform.Read()
		assert.NoError(t, err)
		records = append(records, string(recordBytes))
	}
	cupaloy.SnapshotT(t, jsons.BPJ("["+strings.Join(records, ",")+"]"))
}

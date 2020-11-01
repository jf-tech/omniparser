package jsonlog

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/extensions/omniv21"
	v21 "github.com/jf-tech/omniparser/extensions/omniv21/customfuncs"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat"
	"github.com/jf-tech/omniparser/extensions/omniv21/samples/customfileformats/jsonlog/jsonlogformat"
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
		// builtin schema handler (and its builtin fileformats)
		// with, well, the same schema handler but with our own custom
		// fileformat. Also let's demo how to add a new custom func.
		omniparser.Extension{
			CreateSchemaHandler: omniv21.CreateSchemaHandler,
			CreateSchemaHandlerParams: &omniv21.CreateParams{
				// But use our own FileFormat.
				CustomFileFormats: []fileformat.FileFormat{
					jsonlogformat.NewJSONLogFileFormat(schemaFileBaseName),
				},
			},
			CustomFuncs: customfuncs.Merge(
				customfuncs.CommonCustomFuncs,
				v21.OmniV21CustomFuncs,
				customfuncs.CustomFuncs{
					"normalize_severity": normalizeSeverity,
				}),
		})
	assert.NoError(t, err)
	transform, err := schema.NewTransform(inputFileBaseName, inputFileReader, &transformctx.Ctx{})
	assert.NoError(t, err)

	var records []string
	for {
		recordBytes, err := transform.Read()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		records = append(records, string(recordBytes))
	}
	cupaloy.SnapshotT(t, jsons.BPJ("["+strings.Join(records, ",")+"]"))
}

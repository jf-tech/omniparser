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
	omniv2 "github.com/jf-tech/omniparser/schemaplugin/omni/v2"
	"github.com/jf-tech/omniparser/transformctx"
	"github.com/jf-tech/omniparser/samples/omniv2/customfileformats/jsonlog/jsonlogformat"
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

	parser, err := omniparser.NewParser(
		schemaFileBaseName,
		schemaFileReader,
		// Use this SchemaPluginConfig to effectively replace the
		// builtin omniv2 schema plugin (and its builtin fileformats)
		// with, well, omniv2 schema plugin :) with our own custom
		// fileformat. Also let's demo how to add a new custom func.
		omniparser.SchemaPluginConfig{
			ParseSchema: omniv2.ParseSchema,
			PluginParams: &omniv2.PluginParams{
				CustomFileFormat: jsonlogformat.NewJSONLogFileFormat(schemaFileBaseName),
			},
			CustomFuncs: customfuncs.CustomFuncs{
				"normalize_severity": normalizeSeverity,
			},
		})
	assert.NoError(t, err)
	op, err := parser.GetTransformOp(
		inputFileBaseName,
		inputFileReader,
		&transformctx.Ctx{})
	assert.NoError(t, err)

	var records []string
	for op.Next() {
		recordBytes, err := op.Read()
		assert.NoError(t, err)
		records = append(records, string(recordBytes))
	}
	cupaloy.SnapshotT(t, jsons.BPJ("["+strings.Join(records, ",")+"]"))
}

package sampleutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

func SampleTestCommon(t *testing.T, schemaFile, inputFile string) string {
	schemaFileBaseName := filepath.Base(schemaFile)
	schemaFileReader, err := os.Open(schemaFile)
	assert.NoError(t, err)
	defer schemaFileReader.Close()

	inputFileBaseName := filepath.Base(inputFile)
	inputFileReader, err := os.Open(inputFile)
	assert.NoError(t, err)
	defer inputFileReader.Close()

	parser, err := omniparser.NewParser(schemaFileBaseName, schemaFileReader)
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

	return "[" + strings.Join(records, ",") + "]"
}

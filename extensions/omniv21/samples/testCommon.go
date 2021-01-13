package samples

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

// SampleTestCommon is a test helper for sample tests
func SampleTestCommon(t *testing.T, schemaFile, inputFile string) string {
	schemaFileBaseName := filepath.Base(schemaFile)
	schemaFileReader, err := os.Open(schemaFile)
	assert.NoError(t, err)
	defer schemaFileReader.Close()

	inputFileBaseName := filepath.Base(inputFile)
	inputFileReader, err := os.Open(inputFile)
	assert.NoError(t, err)
	defer inputFileReader.Close()

	schema, err := omniparser.NewSchema(schemaFileBaseName, schemaFileReader)
	assert.NoError(t, err)
	transform, err := schema.NewTransform(inputFileBaseName, inputFileReader, &transformctx.Ctx{})
	assert.NoError(t, err)

	type record struct {
		RawRecord         string
		RawRecordHash     string
		TransformedRecord interface{}
	}
	var records []record
	for {
		recordBytes, err := transform.Read()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		var transformed interface{}
		err = json.Unmarshal(recordBytes, &transformed)
		assert.NoError(t, err)

		raw, err := transform.RawRecord()
		assert.NoError(t, err)
		records = append(records, record{
			RawRecord:         idr.JSONify2(raw.Raw().(*idr.Node)),
			RawRecordHash:     raw.Checksum(),
			TransformedRecord: transformed,
		})
	}
	return jsons.BMM(records)
}

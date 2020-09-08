package customparse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	node "github.com/antchfx/xmlquery"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/omniparser"
	omniv2 "github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

func TestSample(t *testing.T) {
	schemaFile := "./sample_schema.json"
	schemaFileBaseName := filepath.Base(schemaFile)
	schemaFileReader, err := os.Open(schemaFile)
	assert.NoError(t, err)
	defer schemaFileReader.Close()

	inputFile := "./sample.xml"
	inputFileBaseName := filepath.Base(inputFile)
	inputFileReader, err := os.Open(inputFile)
	assert.NoError(t, err)
	defer inputFileReader.Close()

	parser, err := omniparser.NewParser(
		schemaFileBaseName,
		schemaFileReader,
		omniparser.SchemaPluginConfig{
			ParseSchema: omniv2.ParseSchema,
			PluginParams: &omniv2.PluginParams{
				CustomParseFuncs: transform.CustomParseFuncs{
					"employee_personal_details_lookup": employeePersonalDetailsLookup,
					"employee_business_details_lookup": employeeBusinessDetailsLookup,
					"employee_team_lookup":             employeeTempLookup,
				},
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

// Pretend we need to do some secure database look-up for employee's various info based on id.
// The result can be one of these:
// - you can directly return a string value, if this lookup yields a simple string value.
// - you can return a generic map[string]interface{}, if lookup yields a complex structure.
// - you can return a struct that is JSON marshalable, if lookup yields a complex structure.

// These are very contrived examples, only for illustration purposes.
func employeePersonalDetailsLookup(_ *transformctx.Ctx, node *node.Node) (interface{}, error) {
	id := node.InnerText()
	// Pretend some complex logic and/or RPC calls...
	// This custom_parse demonstrates how to return a complex object with map[string]interface{}
	return map[string]interface{}{
		"name": "name-" + id,
		"age":  time.Now().Nanosecond()%50 + 20,
		"home_address": map[string]interface{}{
			"street": "street-" + id,
			"city":   "city-" + id,
			"state":  "state-" + id,
			"zip":    id,
		},
	}, nil
}

func employeeBusinessDetailsLookup(_ *transformctx.Ctx, node *node.Node) (interface{}, error) {
	id := node.InnerText()
	// Pretend some complex logic and/or RPC calls...
	// This custom_parse demonstrates how to return a complex object with golang struct
	type employeeReview struct {
		Year   int    `json:"year"`
		Rating string `json:"rating"`
	}
	type employeeBusinessDetails struct {
		Title          string           `json:"title"`
		YearsOfService int              `json:"years_of_service"`
		Reviews        []employeeReview `json:"reviews"`
	}
	return employeeBusinessDetails{
		Title:          "Sr. Engineer (" + id + ")",
		YearsOfService: 3,
		Reviews: []employeeReview{
			{Year: 2020, Rating: "Exceeded Expectation"},
			{Year: 2019, Rating: "Met Expectation"},
			{Year: 2018, Rating: "Below Expectation"},
		},
	}, nil
}

func employeeTempLookup(_ *transformctx.Ctx, node *node.Node) (interface{}, error) {
	id := node.InnerText()
	// Pretend some complex logic and/or RPC calls...
	// This custom_parse demonstrates how to return a single string value.
	// You can also in schema specify different result_type (such as int, float64) to
	// fit your needs.
	return "TEAM-" + id, nil
}

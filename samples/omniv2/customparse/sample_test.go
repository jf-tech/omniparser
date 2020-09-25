package customparse

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser"
	omniv2 "github.com/jf-tech/omniparser/handlers/omni/v2"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
	"github.com/jf-tech/omniparser/transformctx"
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

	schema, err := omniparser.NewSchema(
		schemaFileBaseName,
		schemaFileReader,
		omniparser.Extension{
			CreateHandler: omniv2.CreateHandler,
			HandlerParams: &omniv2.HandlerParams{
				CustomParseFuncs: transform.CustomParseFuncs{
					"employee_personal_details_lookup": employeePersonalDetailsLookup,
					"employee_business_details_lookup": employeeBusinessDetailsLookup,
					"employee_team_lookup":             employeeTempLookup,
				},
			},
		})
	assert.NoError(t, err)
	tfm, err := schema.NewTransform(inputFileBaseName, inputFileReader, &transformctx.Ctx{})
	assert.NoError(t, err)

	var records []string
	for tfm.Next() {
		recordBytes, err := tfm.Read()
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
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"name": "name-" + id,
		"age":  idInt%40 + 21, // whatever this means :)
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

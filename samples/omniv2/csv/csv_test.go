package xml

import (
	"bytes"
	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/transformctx"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"

	"github.com/jf-tech/omniparser/samples/sampleutil"
)

func Test1_Weather_Data_CSV(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(sampleutil.SampleTestCommon(
		t, "./1_weather_data_csv.schema.json", "./1_weather_data_csv.input.csv")))
}

var benchSchemaFile = "./1_weather_data_csv.schema.json"
var benchInputFile = "./1_weather_data_csv.input.csv"
var benchSchema omniparser.Schema
var benchInput []byte

func init() {
	schema, err := ioutil.ReadFile(benchSchemaFile)
	if err != nil {
		panic(err)
	}
	benchSchema, err = omniparser.NewSchema("bench", bytes.NewReader(schema))
	if err != nil {
		panic(err)
	}
	benchInput, err = ioutil.ReadFile(benchInputFile)
	if err != nil {
		panic(err)
	}
}

// go test -bench=. -benchmem -benchtime=30s
// Benchmark1_Weather_Data_CSV-8   	  148292	    240290 ns/op	   78183 B/op	    1622 allocs/op

func Benchmark1_Weather_Data_CSV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transform, err := benchSchema.NewTransform(
			"bench", bytes.NewReader(benchInput), &transformctx.Ctx{})
		if err != nil {
			b.FailNow()
		}
		for transform.Next() {
			_, err := transform.Read()
			if err != nil {
				b.FailNow()
			}
		}
	}
}

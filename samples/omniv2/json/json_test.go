package json

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/samples/sampleutil"
	"github.com/jf-tech/omniparser/transformctx"
)

func Test1_Single_Object(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(sampleutil.SampleTestCommon(
		t, "./1_single_object.schema.json", "./1_single_object.input.json")))
}

func Test2_Multiple_Objects(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(sampleutil.SampleTestCommon(t,
		"./2_multiple_objects.schema.json", "./2_multiple_objects.input.json")))
}

func Test3_XPathDynamic(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(sampleutil.SampleTestCommon(t,
		"./3_xpathdynamic.schema.json", "./3_xpathdynamic.input.json")))
}

var benchSchemaFile = "./2_multiple_objects.schema.json"
var benchInputFile = "./2_multiple_objects.input.json"
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
// Benchmark2_Multiple_Objects-4   	   95775	    381437 ns/op	   84870 B/op	    2180 allocs/op

func Benchmark2_Multiple_Objects(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transform, err := benchSchema.NewTransform(
			"bench", bytes.NewReader(benchInput), &transformctx.Ctx{})
		if err != nil {
			panic(err)
		}
		for transform.Next() {
			_, err := transform.Read()
			if err != nil {
				panic(err)
			}
		}
	}
}

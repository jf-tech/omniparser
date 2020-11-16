package fixedlength

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/extensions/omniv21/samples"
	"github.com/jf-tech/omniparser/transformctx"
)

func Test1_Single_Row(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./1_single_row.schema.json", "./1_single_row.input.txt")))
}

func Test2_Multi_Rows(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./2_multi_rows.schema.json", "./2_multi_rows.input.txt")))
}

func Test3_Header_Footer(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./3_header_footer.schema.json", "./3_header_footer.input.txt")))
}

var benchSchemaFile = "./3_header_footer.schema.json"
var benchInputFile = "./3_header_footer.input.txt"
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

// Benchmark3_Header_Footer-8   	    2798	    385303 ns/op	   77909 B/op	    1891 allocs/op
func Benchmark3_Header_Footer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transform, err := benchSchema.NewTransform(
			"bench", bytes.NewReader(benchInput), &transformctx.Ctx{})
		if err != nil {
			b.FailNow()
		}
		for {
			_, err := transform.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.FailNow()
			}
		}
	}
}

package edi

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/extensions/omniv21/samples"
	"github.com/jf-tech/omniparser/transformctx"
)

func Test1_CanadaPost_EDI_214(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./1_canadapost_edi_214.schema.json", "./1_canadapost_edi_214.input.txt")))
}

var benchSchemaFile = "./1_canadapost_edi_214.schema.json"
var benchInputFile = "./1_canadapost_edi_214.input.txt"
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

func Benchmark1_CanadaPost_EDI_214(b *testing.B) {
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

package edi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat/edi"
	"github.com/jf-tech/omniparser/extensions/omniv21/samples"
	"github.com/jf-tech/omniparser/transformctx"
)

func Test1_CanadaPost_EDI_214(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./1_canadapost_edi_214.schema.json", "./1_canadapost_edi_214.input.txt")))
}

func Test2_UPS_EDI_210(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./2_ups_edi_210.schema.json", "./2_ups_edi_210.input.txt")))
}

func Test3_NonValidatingReader(t *testing.T) {
	schemaFileReader, err := os.Open("./2_ups_edi_210.schema.json")
	assert.NoError(t, err)
	defer schemaFileReader.Close()

	inputFileReader, err := os.Open("./2_ups_edi_210.input.txt")
	assert.NoError(t, err)
	defer inputFileReader.Close()

	schemaContent, err := ioutil.ReadAll(schemaFileReader)
	assert.NoError(t, err)

	type ediSchema struct {
		FileDecl *edi.FileDecl `json:"file_declaration"`
	}
	var schema ediSchema
	err = json.Unmarshal(schemaContent, &schema)
	assert.NoError(t, err)

	type rawElem struct {
		ElemIndex int
		CompIndex int
		Data      string
	}
	type rawSeg struct {
		Name  string
		Raw   string
		Elems []rawElem
	}
	r := edi.NewNonValidatingReader(inputFileReader, schema.FileDecl)
	var rawSegs []rawSeg
	for {
		seg, err := r.Read()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		rawSegs = append(rawSegs, rawSeg{
			Name: seg.Name,
			Raw:  string(seg.Raw),
			Elems: func() []rawElem {
				var elems []rawElem
				for _, e := range seg.Elems {
					elems = append(elems, rawElem{
						ElemIndex: e.ElemIndex,
						CompIndex: e.CompIndex,
						Data:      string(e.Data),
					})
				}
				return elems
			}(),
		})
	}
	cupaloy.SnapshotT(t, jsons.BPM(rawSegs))
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

// Benchmark1_CanadaPost_EDI_214-8   	     963	   1248482 ns/op	  362593 B/op	    8301 allocs/op
func Benchmark1_CanadaPost_EDI_214(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transform, err := benchSchema.NewTransform(
			"bench", bytes.NewReader(benchInput), &transformctx.Ctx{})
		if err != nil {
			b.FailNow()
		}
		for {
			_, err = transform.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.FailNow()
			}
		}
	}
}

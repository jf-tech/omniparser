package csv2

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

type testCase struct {
	schemaFile string
	inputFile  string
	schema     omniparser.Schema
	input      []byte
}

const (
	test1_Single_Row = iota
	test2_Fixed_Multi_Row
	test3_Multi_Row_HeaderFooter
	test4_Nested
)

var tests = []testCase{
	{
		// test1_Single_Row
		schemaFile: "./1_single_row.schema.json",
		inputFile:  "./1_single_row.input.csv",
	},
	{
		// test2_Fixed_Multi_Row
		schemaFile: "./2_fixed_multi_row.schema.json",
		inputFile:  "./2_fixed_multi_row.input.csv",
	},
	{
		// test3_Multi_Row_HeaderFooter
		schemaFile: "./3_multi_row_headerfooter.schema.json",
		inputFile:  "./3_multi_row_headerfooter.input.csv",
	},
	{
		// test4_Nested
		schemaFile: "./4_nested.schema.json",
		inputFile:  "./4_nested.input.csv",
	},
}

func init() {
	for i := range tests {
		schema, err := ioutil.ReadFile(tests[i].schemaFile)
		if err != nil {
			panic(err)
		}
		tests[i].schema, err = omniparser.NewSchema("bench", bytes.NewReader(schema))
		if err != nil {
			panic(err)
		}
		tests[i].input, err = ioutil.ReadFile(tests[i].inputFile)
		if err != nil {
			panic(err)
		}
	}
}

func (tst testCase) doTest(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(t, tst.schemaFile, tst.inputFile)))
}

func (tst testCase) doBenchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transform, err := tst.schema.NewTransform(
			"bench", bytes.NewReader(tst.input), &transformctx.Ctx{})
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

func Test1_Single_Row(t *testing.T) {
	tests[test1_Single_Row].doTest(t)
}

func Test2_Fixed_Multi_Row(t *testing.T) {
	tests[test2_Fixed_Multi_Row].doTest(t)
}

func Test3_Multi_Row_HeaderFooter(t *testing.T) {
	tests[test3_Multi_Row_HeaderFooter].doTest(t)
}

func Test4_Nested(t *testing.T) {
	tests[test4_Nested].doTest(t)
}

// Benchmark1_Single_Row-8   	       11401	    105326 ns/op	   80756 B/op	    1424 allocs/op
func Benchmark1_Single_Row(b *testing.B) {
	tests[test1_Single_Row].doBenchmark(b)
}

// Benchmark2_Fixed_Multi_Row-8   	   82680	     14498 ns/op	   13268 B/op	     261 allocs/op
func Benchmark2_Fixed_Multi_Row(b *testing.B) {
	tests[test2_Fixed_Multi_Row].doBenchmark(b)
}

// Benchmark3_Multi_Row_HeaderFooter-8   	   67137	     17708 ns/op	   14525 B/op	     290 allocs/op
func Benchmark3_Multi_Row_HeaderFooter(b *testing.B) {
	tests[test3_Multi_Row_HeaderFooter].doBenchmark(b)
}

// Benchmark4_Nested-8       	   26906	     44804 ns/op	   32291 B/op	     719 allocs/op
func Benchmark4_Nested(b *testing.B) {
	tests[test4_Nested].doBenchmark(b)
}

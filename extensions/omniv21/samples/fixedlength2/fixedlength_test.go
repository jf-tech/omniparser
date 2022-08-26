package fixedlength2

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
	test2_Multi_Rows
	test3_Header_Footer
	test4_Nested
)

var tests = []testCase{
	{
		// test1_Single_Row
		schemaFile: "./1_single_row.schema.json",
		inputFile:  "./1_single_row.input.txt",
	},
	{
		// test2_Multi_Rows
		schemaFile: "./2_multi_rows.schema.json",
		inputFile:  "./2_multi_rows.input.txt",
	},

	{
		// test3_Header_Footer
		schemaFile: "./3_header_footer.schema.json",
		inputFile:  "./3_header_footer.input.txt",
	},
	{
		// test4_Nested
		schemaFile: "./4_nested.schema.json",
		inputFile:  "./4_nested.input.txt",
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

func Test2_Multi_Rows(t *testing.T) {
	tests[test2_Multi_Rows].doTest(t)
}

func Test3_Header_Footer(t *testing.T) {
	tests[test3_Header_Footer].doTest(t)
}

func Test4_Nested(t *testing.T) {
	tests[test4_Nested].doTest(t)
}

// Benchmark1_Single_Row-8      	   25898	     45728 ns/op	   28169 B/op	     647 allocs/op
func Benchmark1_Single_Row(b *testing.B) {
	tests[test1_Single_Row].doBenchmark(b)
}

// Benchmark2_Multi_Rows-8      	   15338	     78273 ns/op	   30179 B/op	     618 allocs/op
func Benchmark2_Multi_Rows(b *testing.B) {
	tests[test2_Multi_Rows].doBenchmark(b)
}

// Benchmark3_Header_Footer-8   	    6390	    178301 ns/op	   76571 B/op	    1501 allocs/op
func Benchmark3_Header_Footer(b *testing.B) {
	tests[test3_Header_Footer].doBenchmark(b)
}

// Benchmark4_Nested-8          	   10000	    107948 ns/op	   78986 B/op	    1535 allocs/op
func Benchmark4_Nested(b *testing.B) {
	tests[test4_Nested].doBenchmark(b)
}

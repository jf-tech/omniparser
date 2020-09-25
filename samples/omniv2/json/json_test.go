package json

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"

	"github.com/jf-tech/omniparser/samples/sampleutil"
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

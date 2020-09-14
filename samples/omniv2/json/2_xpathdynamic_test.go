package json

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/samples/sampleutil"
)

func Test2_XPathDynamic(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(
		sampleutil.SampleTestCommon(t, "./2_xpathdynamic_schema.json", "./2_xpathdynamic.json")))
}

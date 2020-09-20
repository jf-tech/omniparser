package xml

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/samples/sampleutil"
)

func Test2_Multiple_Objects(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(sampleutil.SampleTestCommon(
		t, "./2_multiple_objects.schema.json", "./2_multiple_objects.input.xml")))
}

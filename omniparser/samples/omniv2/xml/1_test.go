package xml

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/omniparser/samples/sampleutil"
)

func Test1(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(sampleutil.SampleTestCommon(t, "./1_schema.json", "./1.xml")))
}

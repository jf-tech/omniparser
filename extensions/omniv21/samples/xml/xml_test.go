package xml

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"

	"github.com/jf-tech/omniparser/extensions/omniv21/samples"
)

func Test1_DateTime_Parse_And_Format(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./1_datetime_parse_and_format.schema.json", "./1_datetime_parse_and_format.input.xml")))
}

func Test2_Multiple_Objects(t *testing.T) {
	cupaloy.SnapshotT(t, jsons.BPJ(samples.SampleTestCommon(
		t, "./2_multiple_objects.schema.json", "./2_multiple_objects.input.xml")))
}

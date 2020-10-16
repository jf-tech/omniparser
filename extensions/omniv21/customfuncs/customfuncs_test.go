package customfuncs

import (
	"sort"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

func TestDumpOmniV21CustomFuncNames(t *testing.T) {
	var names []string
	for name := range OmniV21CustomFuncs {
		names = append(names, name)
	}
	sort.Strings(names)
	cupaloy.SnapshotT(t, jsons.BPM(names))
}

func TestCopyFunc(t *testing.T) {
	j := `{ "a": 1, "b": "2", "c": true, "d": null, "e": { "f": "three" }, "g": [ 1, "2", "three"] }`
	r, err := idr.NewJSONStreamReader(strings.NewReader(j), ".")
	assert.NoError(t, err)
	n, err := r.Read()
	assert.NoError(t, err)
	dest, err := CopyFunc(nil, n)
	assert.NoError(t, err)
	cupaloy.SnapshotT(t, jsons.BPM(dest))
}

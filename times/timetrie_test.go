package times

import (
	"sort"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/strs"
)

func TestAddToTrie(t *testing.T) {
	trie := strs.NewRuneTrie()
	addToTrie(trie, trieEntry{pattern: "abc", layout: "123"})
	assert.PanicsWithValue(t,
		"pattern 'abc' caused a collision",
		func() {
			addToTrie(trie, trieEntry{pattern: "abc", layout: "dup"})
		})
}

func TestDumpDateTimeTrie(t *testing.T) {
	t.Logf("total trie nodes created: %d", dateTimeTrie.NodeCount())
	cupaloy.SnapshotT(t, jsons.BPM(dateTimeTrie))
}

func TestDumpAllTimezones(t *testing.T) {
	var tzs []string
	for tz := range tzList {
		tzs = append(tzs, tz)
	}
	sort.Strings(tzs)
	cupaloy.SnapshotT(t, jsons.BPM(tzs))
}

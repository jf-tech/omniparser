package times

import (
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

func TestInitDateTimeTrie(t *testing.T) {
	trie := initDateTimeTrie()
	t.Logf("total trie nodes created: %d", trie.NodeCount())
	cupaloy.SnapshotT(t, jsons.BPM(trie))
}

func TestInitTimezones(t *testing.T) {
	m := initTimezones()
	t.Logf("total timezones: %d", len(m))
	cupaloy.SnapshotT(t, jsons.BPM(m))
}

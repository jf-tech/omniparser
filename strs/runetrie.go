package strs

import (
	"encoding/json"
)

type trieNode struct {
	valueStored bool
	value       interface{}
	children    map[string]*trieNode
}

func (tn *trieNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ValueStored bool
		Value       interface{}
		Children    map[string]*trieNode
	}{
		ValueStored: tn.valueStored,
		Value:       tn.value,
		Children:    tn.children,
	})
}

// KeyMapper maps one rune or a segment of consecutive runes into a key for inserting
// into RuneTrie, and returns how many runes needs to be advanced
type KeyMapper func(rs []rune, index int) (advance int, key string)

// RuneTrie is a trie of strings. By default, each trie node corresponds to a rune
// in a string, however, user of RuneTrie can provide a custom mapper that can map
// one or multiple consecutive runes into a trie key.
type RuneTrie struct {
	root      *trieNode
	mapper    KeyMapper
	nodeCount int
}

func (t *RuneTrie) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Root *trieNode
	}{
		Root: t.root,
	})
}

// NewRuneTrie creates a new RuneTrie.
func NewRuneTrie(mappers ...KeyMapper) *RuneTrie {
	mapper := KeyMapper(nil)
	switch len(mappers) {
	case 0:
	case 1:
		mapper = mappers[0]
	default:
		panic("must not call with more than one mapper")
	}
	return &RuneTrie{
		root:      &trieNode{},
		mapper:    mapper,
		nodeCount: 1,
	}
}

func (t *RuneTrie) key(rs []rune, index int) (advance int, key string) {
	if t.mapper == nil {
		return 1, string(rs[index])
	}
	return t.mapper(rs, index)
}

// Add inserts a string and its associated value into RuneTrie and returns true
// if the value is added; false if an existing value is replaced.
func (t *RuneTrie) Add(s string, value interface{}) bool {
	n := t.root
	rs := []rune(s)
	for i := 0; i < len(rs); {
		adv, k := t.key(rs, i)
		i += adv
		child, _ := n.children[k]
		if child == nil {
			if n.children == nil {
				n.children = map[string]*trieNode{}
			}
			child = &trieNode{}
			t.nodeCount++
			n.children[k] = child
		}
		n = child
	}
	added := !n.valueStored
	n.value = value
	n.valueStored = true
	return added
}

// Get looks up a string in RuneTrie and returns its associated value if found.
func (t *RuneTrie) Get(s string) (interface{}, bool) {
	n := t.root
	rs := []rune(s)
	for i := 0; i < len(rs); {
		adv, k := t.key(rs, i)
		i += adv
		n = n.children[k]
		if n == nil {
			return nil, false
		}
	}
	return n.value, n.valueStored
}

func (t *RuneTrie) NodeCount() int {
	return t.nodeCount
}

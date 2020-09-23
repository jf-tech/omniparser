package strs

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// trieNode represents a node in a RuneTrie. In general a key of
// 'children' is a single rune (which is int32 cast to uint64), if a
// RuneTrie created with the default mapper. However, custom mapper
// can choose to map one or consecutive runes into some custom uint64
// value as key. Given a rune is int32, custom mapper is advised to
// utilize upper 32-bit for those beyond a single/direct rune to key
// mapping so that special mapping can be distinguished apart from
// direct rune mapping. Using uint64 instead of a more flexible string
// key type is to save allocation and uint64 range is large enough to
// be useful.
type trieNode struct {
	valueStored bool
	value       interface{}
	children    map[uint64]*trieNode
}

// MarshalJSON json marshals a trieNode. Should only be used in tests.
func (tn *trieNode) MarshalJSON() ([]byte, error) {
	var children map[string]*trieNode
	if tn.children != nil {
		children = map[string]*trieNode{}
		for k, v := range tn.children {
			if k <= utf8.MaxRune {
				children[string(rune(uint32(k)))] = v
			} else {
				// for non-rune key, break the uint64 into two uin32s
				// so the marshaled out key of the map looks somewhat
				// readable.
				k1 := (k >> 32) & 0xFFFF
				k2 := k & 0xFFFF
				children[fmt.Sprintf("0x%X|0x%X", k1, k2)] = v
			}
		}
	}
	return json.Marshal(&struct {
		ValueStored bool
		Value       interface{}
		Children    map[string]*trieNode
	}{
		ValueStored: tn.valueStored,
		Value:       tn.value,
		Children:    children,
	})
}

// KeyMapper maps one rune or a segment of consecutive runes into a key for inserting
// into RuneTrie, and returns how many bytes needs to be advanced
type KeyMapper func(s string, index int) (advance int, key uint64)

// RuneTrie is a trie of strings. By default, each trie node corresponds to a rune
// in a string, however, user of RuneTrie can provide a custom mapper that can map
// one or multiple consecutive runes into a trie key.
type RuneTrie struct {
	root      *trieNode
	mapper    KeyMapper
	nodeCount int
}

// MarshalJSON json marshals a RuneTrie. Should only be used in tests.
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

func (t *RuneTrie) key(s string, index int) (advance int, key uint64) {
	if t.mapper == nil {
		r, size := utf8.DecodeRuneInString(s[index:])
		return size, uint64(r)
	}
	return t.mapper(s, index)
}

// Add inserts a string and its associated value into RuneTrie and returns true
// if the value is added; false if an existing value is replaced.
func (t *RuneTrie) Add(s string, value interface{}) bool {
	n := t.root
	for i := 0; i < len(s); {
		adv, k := t.key(s, i)
		i += adv
		child, _ := n.children[k]
		if child == nil {
			if n.children == nil {
				n.children = map[uint64]*trieNode{}
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
	for i := 0; i < len(s); {
		adv, k := t.key(s, i)
		i += adv
		n = n.children[k]
		if n == nil {
			return nil, false
		}
	}
	return n.value, n.valueStored
}

// NodeCount returns the total number of nodes in the trie.
func (t *RuneTrie) NodeCount() int {
	return t.nodeCount
}

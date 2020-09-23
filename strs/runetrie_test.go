package strs

import (
	"testing"
	"unicode"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/jsons"
)

func intPtr(n int) *int {
	return &n
}

func TestRuneTrie(t *testing.T) {
	type get struct {
		s        string
		expected *int
	}
	for _, test := range []struct {
		name   string
		mapper KeyMapper
		input  []string
		count  int
		gets   []get
	}{
		{
			name:  "empty trie",
			count: 1,
			gets:  []get{{s: "abc", expected: nil}},
		},
		{
			name:  "single string",
			input: []string{"hello, こんにちは!"},
			count: 14,
			gets: []get{
				{s: "hello, こんにちは!", expected: intPtr(0)},
				{s: "hello,     ", expected: nil},
			},
		},
		{
			name: "multiple strings with mapper",
			mapper: func(rs []rune, index int) (advance int, key string) {
				switch {
				case unicode.IsDigit(rs[index]):
					for advance = index + 1; advance < len(rs) && unicode.IsDigit(rs[advance]); advance++ {
					}
					return advance - index, "#digit"
				case unicode.IsSpace(rs[index]):
					for advance = index + 1; advance < len(rs) && unicode.IsSpace(rs[advance]); advance++ {
					}
					return advance - index, "#space"
				default:
					return 1, string(rs[index])
				}
			},
			count: 8,
			input: []string{"a   b c", "a b\t99d", "a\nb\r12"},
			gets: []get{
				{s: "a\t\tb\n\nc", expected: intPtr(0)},
				{s: "a\nb\r12345d", expected: intPtr(1)},
				{s: "a\tb     98765", expected: intPtr(2)},
				{s: "a   bc", expected: nil},
				{s: "a   b123x", expected: nil},
				{s: "a   b123dx", expected: nil},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var trie *RuneTrie
			if test.mapper == nil {
				trie = NewRuneTrie()
			} else {
				trie = NewRuneTrie(test.mapper)
			}
			for i, s := range test.input {
				trie.Add(s, i)
			}
			assert.Equal(t, test.count, trie.NodeCount())
			cupaloy.SnapshotT(t, jsons.BPM(trie))
			for _, g := range test.gets {
				v, found := trie.Get(g.s)
				if g.expected == nil {
					assert.False(t, found)
					assert.Nil(t, v)
				} else {
					assert.True(t, found)
					assert.Equal(t, *g.expected, v.(int))
				}
			}
		})
	}

	// Also testing the panic for more than one mapper provided
	assert.PanicsWithValue(t, "must not call with more than one mapper", func() {
		NewRuneTrie(
			func([]rune, int) (int, string) { return 0, "" },
			func([]rune, int) (int, string) { return 0, "" })
	})
}

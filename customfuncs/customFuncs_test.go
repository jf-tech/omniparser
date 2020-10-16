package customfuncs

import (
	"sort"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"
)

func TestDumpCommonCustomFuncNames(t *testing.T) {
	var names []string
	for name := range CommonCustomFuncs {
		names = append(names, name)
	}
	sort.Strings(names)
	cupaloy.SnapshotT(t, jsons.BPM(names))
}

func TestMerge(t *testing.T) {
	fs1 := CustomFuncs{
		"a": 1,
		"b": 2,
	}
	fs2 := CustomFuncs{
		"a": 3,
		"c": 4,
	}
	assert.Equal(t,
		CustomFuncs{
			"a": 3,
			"b": 2,
			"c": 4,
		},
		Merge(fs1, fs2, nil))
}

func TestConcat(t *testing.T) {
	for _, test := range []struct {
		name     string
		strs     []string
		expected string
	}{
		{
			name:     "empty input",
			strs:     nil,
			expected: "",
		},
		{
			name:     "one empty string input",
			strs:     []string{""},
			expected: "",
		},
		{
			name:     "one non-empty string input",
			strs:     []string{"a b c"},
			expected: "a b c",
		},
		{
			name:     "multiple strings",
			strs:     []string{"", "a b c", "", " e f"},
			expected: "a b c e f",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := concat(nil, test.strs...)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestLower(t *testing.T) {
	s, err := lower(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "", s)

	s, err = lower(nil, "AbCeDfG 0123456789")
	assert.NoError(t, err)
	assert.Equal(t, "abcedfg 0123456789", s)
}

func TestUpper(t *testing.T) {
	s, err := upper(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "", s)

	s, err = upper(nil, "abCeDfG 0123456789")
	assert.NoError(t, err)
	assert.Equal(t, "ABCEDFG 0123456789", s)
}

func TestUUIDV3(t *testing.T) {
	result, err := uuidv3(nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "4ae71336-e44b-39bf-b9d2-752e234818a5", result)

	result, err = uuidv3(nil, "abc")
	assert.NoError(t, err)
	assert.Equal(t, "522ec739-ca63-3ec5-b082-08ce08ad65e2", result)
}

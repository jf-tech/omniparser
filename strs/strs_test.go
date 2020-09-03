package strs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/testlib"
)

func TestIsStrNonBlank(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		nonBlank bool
	}{
		{
			name:     "empty string",
			input:    "",
			nonBlank: false,
		},
		{
			name:     "blank string",
			input:    "      ",
			nonBlank: false,
		},
		{
			name:     "non blank",
			input:    "abc",
			nonBlank: true,
		},
		{
			name:     "non blank after trimming",
			input:    "  abc  ",
			nonBlank: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.nonBlank, IsStrNonBlank(test.input))
			inputCopy := test.input
			assert.Equal(t, test.nonBlank, IsStrPtrNonBlank(&inputCopy))
		})
	}
	assert.False(t, IsStrPtrNonBlank(nil))
}

func TestFirstNonBlank(t *testing.T) {
	assert.Equal(t, "abc", FirstNonBlank("", "   ", "abc", "def"))
	assert.Equal(t, "", FirstNonBlank("", "   ", "           "))
	assert.Equal(t, "", FirstNonBlank())
}

func TestStrPtrOrElse(t *testing.T) {
	assert.Equal(t, "this", StrPtrOrElse(testlib.StrPtr("this"), "that"))
	assert.Equal(t, "that", StrPtrOrElse(nil, "that"))
}

func TestCopyStrPtr(t *testing.T) {
	assert.True(t, CopyStrPtr(nil) == nil)
	src := testlib.StrPtr("abc")
	dst := CopyStrPtr(src)
	assert.Equal(t, *src, *dst)
	assert.True(t, fmt.Sprintf("%p", src) != fmt.Sprintf("%p", dst))
}

func TestBuildFQDN(t *testing.T) {
	for _, test := range []struct {
		name     string
		namelets []string
		expected string
	}{
		{
			name:     "nil",
			namelets: nil,
			expected: "",
		},
		{
			name:     "empty",
			namelets: []string{},
			expected: "",
		},
		{
			name:     "single",
			namelets: []string{"one"},
			expected: "one",
		},
		{
			name:     "multiple",
			namelets: []string{"one", "", "three", "four"},
			expected: "one..three.four",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, BuildFQDN(test.namelets...))
		})
	}
}

func TestCopySlice(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          []string
		expectedOutput []string
	}{
		{
			name:           "nil",
			input:          nil,
			expectedOutput: nil,
		},
		{
			name:           "empty slice",
			input:          []string{},
			expectedOutput: nil,
		},
		{
			name:           "non-empty slice",
			input:          []string{"abc", ""},
			expectedOutput: []string{"abc", ""},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cp := CopySlice(test.input)
			// First make sure the copy contains what's expected.
			assert.Equal(t, test.expectedOutput, cp)
			if len(test.input) >= 2 {
				// Second test if modifying the original won't affect the copy
				// (that's what this copy func is all about)
				test.input[0] = test.input[1]
				assert.NotEqual(t, test.input, cp)
			}
		})
	}
}

func TestMergeSlices(t *testing.T) {
	for _, test := range []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "both nil",
			slice1:   nil,
			slice2:   nil,
			expected: nil,
		},
		{
			name:     "1 nil, 2 not nil",
			slice1:   nil,
			slice2:   []string{"", "abc"},
			expected: []string{"", "abc"},
		},
		{
			name:     "1 not nil, 2 nil",
			slice1:   []string{"abc", ""},
			slice2:   nil,
			expected: []string{"abc", ""},
		},
		{
			name:     "both not nil",
			slice1:   []string{"abc", ""},
			slice2:   []string{"", "abc"},
			expected: []string{"abc", "", "", "abc"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			merged := MergeSlices(test.slice1, test.slice2)
			// also very importantly to make sure the resulting merged is a new copy so modifying
			// the input slices won't affect the merged slice.
			if len(test.slice1) > 0 {
				test.slice1[0] = "modified"
			}
			if len(test.slice2) > 0 {
				test.slice2[0] = "modified"
			}
			assert.Equal(t, test.expected, merged)
		})
	}
}

func TestHasDup(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected bool
	}{
		{
			name:     "nil",
			input:    nil,
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: false,
		},
		{
			name:     "non-empty slice with no dups",
			input:    []string{"abc", ""},
			expected: false,
		},
		{
			name:     "non-empty slice with dups",
			input:    []string{"", "abc", ""},
			expected: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, HasDup(test.input))
		})
	}
}

func TestMapSlice(t *testing.T) {
	t.Run("map error", func(t *testing.T) {
		errorMap := func(_ string) (string, error) {
			return "abc", errors.New("map error")
		}
		result, err := MapSlice([]string{"abc", ""}, errorMap)
		assert.Error(t, err)
		assert.Equal(t, "map error", err.Error())
		assert.Nil(t, result)
	})

	t.Run("map success", func(t *testing.T) {
		input := []string{"abc", ""}
		index := 0
		mirrorMap := func(_ string) (string, error) {
			index++
			return input[len(input)-index], nil
		}
		result, err := MapSlice(input, mirrorMap)
		assert.NoError(t, err)
		assert.Equal(t, []string{"", "abc"}, result)
	})

	t.Run("map nil", func(t *testing.T) {
		result, err := MapSlice(nil, func(s string) (string, error) {
			return s + "...", nil
		})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestNoErrMapSlice(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "non-empty slice",
			input:    []string{"abc", ""},
			expected: []string{"", "abc"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			index := 0
			mirrorMap := func(s string) string {
				index++
				return test.input[len(test.input)-index]
			}
			assert.Equal(t, test.expected, NoErrMapSlice(test.input, mirrorMap))
		})
	}
}

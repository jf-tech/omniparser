package customfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

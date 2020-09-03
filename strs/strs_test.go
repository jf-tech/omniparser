package strs

import (
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

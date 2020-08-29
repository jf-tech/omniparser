package strs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/testlib"
)

func TestStrPtrOrElse(t *testing.T) {
	assert.Equal(t, "this", StrPtrOrElse(testlib.StrPtr("this"), "that"))
	assert.Equal(t, "that", StrPtrOrElse(nil, "that"))
}

package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/testutil"
)

func TestStrPtrOrElse(t *testing.T) {
	assert.Equal(t, "this", StrPtrOrElse(testutil.StrPtr("this"), "that"))
	assert.Equal(t, "that", StrPtrOrElse(nil, "that"))
}

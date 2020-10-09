package errs

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsErrTransformFailed(t *testing.T) {
	assert.True(t, IsErrTransformFailed(ErrTransformFailed("test")))
	assert.Equal(t, "test", ErrTransformFailed("test").Error())
	assert.False(t, IsErrTransformFailed(io.EOF))
}

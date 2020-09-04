package testlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntPtr(t *testing.T) {
	np := IntPtr(31415926)
	assert.NotNil(t, np)
	assert.Equal(t, 31415926, *np)
}

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegex(t *testing.T) {
	RegexCache = NewLoadingCache()
	assert.Equal(t, 0, len(RegexCache.DumpForTest()))
	// failure case
	expr, err := GetRegex("[")
	assert.Error(t, err)
	assert.Equal(t, "error parsing regexp: missing closing ]: `[`", err.Error())
	assert.Nil(t, expr)
	assert.Equal(t, 0, len(RegexCache.DumpForTest()))
	// success case
	expr, err = GetRegex("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
	assert.NoError(t, err)
	assert.NotNil(t, expr)
	assert.Equal(t, 1, len(RegexCache.DumpForTest()))
	// repeat success case shouldn't case any cache growth
	expr, err = GetRegex("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
	assert.NoError(t, err)
	assert.NotNil(t, expr)
	assert.Equal(t, 1, len(RegexCache.DumpForTest()))
}

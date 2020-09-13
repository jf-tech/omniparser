package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetXPathExpr(t *testing.T) {
	XPathExprCache = NewLoadingCache()
	assert.Equal(t, 0, len(XPathExprCache.DumpForTest()))
	// failure case
	expr, err := GetXPathExpr(">")
	assert.Error(t, err)
	assert.Equal(t, "expression must evaluate to a node-set", err.Error())
	assert.Nil(t, expr)
	assert.Equal(t, 0, len(XPathExprCache.DumpForTest()))
	// success case
	expr, err = GetXPathExpr("/A/B[.!='test']")
	assert.NoError(t, err)
	assert.NotNil(t, expr)
	assert.Equal(t, 1, len(XPathExprCache.DumpForTest()))
	// repeat success case shouldn't case any cache growth
	expr, err = GetXPathExpr("/A/B[.!='test']")
	assert.NoError(t, err)
	assert.NotNil(t, expr)
	assert.Equal(t, 1, len(XPathExprCache.DumpForTest()))
}

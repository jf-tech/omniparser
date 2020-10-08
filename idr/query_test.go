package idr

import (
	"fmt"
	"testing"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/stretchr/testify/assert"
)

func TestLoadXPathExpr(t *testing.T) {
	for _, test := range []struct {
		name          string
		exprStr       string
		flags         []uint
		existsInCache bool
		err           string
	}{
		{
			name:          "valid expr; not in cache before; added to cache",
			exprStr:       ".",
			flags:         nil,
			existsInCache: false,
			err:           "",
		},
		{
			name:          "valid expr; not in cache before; not added to cache",
			exprStr:       ".",
			flags:         []uint{DisableXPathCache},
			existsInCache: false,
			err:           "",
		},
		{
			name:          "valid expr; in cache before; not added to cache again",
			exprStr:       ".",
			flags:         nil,
			existsInCache: true,
			err:           "",
		},
		{
			name:          "not valid expr",
			exprStr:       "]",
			flags:         nil,
			existsInCache: false,
			err:           "xpath ']' compilation failed: expression must evaluate to a node-set",
		},
		{
			name:          "more than one flag",
			exprStr:       ".",
			flags:         []uint{DisableXPathCache, DisableXPathCache},
			existsInCache: false,
			err:           "only one flag is allowed, instead got: 2",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			addedToCache := true
			for _, f := range test.flags {
				if f&DisableXPathCache != 0 {
					addedToCache = false
				}
			}
			caches.XPathExprCache = caches.NewLoadingCache()
			if test.existsInCache {
				_, err := caches.GetXPathExpr(test.exprStr)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
			}
			expr, err := loadXPathExpr(test.exprStr, test.flags)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, expr)
				if test.existsInCache {
					assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
				} else {
					assert.Equal(t, 0, len(caches.XPathExprCache.DumpForTest()))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, expr)
				if addedToCache {
					exprInCache, err := caches.XPathExprCache.Get(test.exprStr, func(interface{}) (interface{}, error) {
						return nil, fmt.Errorf("expr '%s' should've already existed in cache, but not", test.exprStr)
					})
					assert.NoError(t, err)
					assert.True(t, expr == exprInCache)
				}
			}
		})
	}
}

func TestQueryIter(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	expr, err := caches.GetXPathExpr(".")
	assert.NoError(t, err)
	iter := QueryIter(tt.elemB, expr)
	assert.True(t, tt.elemB == iter.Current().(*navigator).root)
	assert.True(t, tt.elemB == iter.Current().(*navigator).cur)
}

func TestAnyMatch(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	expr, err := caches.GetXPathExpr(".")
	assert.NoError(t, err)
	assert.True(t, MatchAny(tt.elemB, expr))
	expr, err = caches.GetXPathExpr("not_existing")
	assert.NoError(t, err)
	assert.False(t, MatchAny(tt.elemB, expr))
}

func TestMatchAll_Dot(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	nodes, err := MatchAll(tt.elemB, ".")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nodes))
	assert.True(t, tt.elemB == nodes[0])
	assert.Equal(t, 0, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchAll_InvalidExpr(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	nodes, err := MatchAll(tt.elemB, "]")
	assert.Error(t, err)
	assert.Equal(t, "xpath ']' compilation failed: expression must evaluate to a node-set", err.Error())
	assert.Equal(t, 0, len(nodes))
	assert.Equal(t, 0, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchAll_NoMatch(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	nodes, err := MatchAll(tt.elemC, "non_existing")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nodes))
	assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchAll_MultipleMatches(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	nodes, err := MatchAll(tt.elemC, "*")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodes))
	assert.True(t, tt.elemC3 == nodes[0])
	assert.True(t, tt.elemC4 == nodes[1])
	assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchSingle_Dot(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	n, err := MatchSingle(tt.elemB, ".")
	assert.NoError(t, err)
	assert.True(t, tt.elemB == n)
	assert.Equal(t, 0, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchSingle_InvalidExpr(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	n, err := MatchSingle(tt.elemB, "]")
	assert.Error(t, err)
	assert.Equal(t, "xpath ']' compilation failed: expression must evaluate to a node-set", err.Error())
	assert.Nil(t, n)
	assert.Equal(t, 0, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchSingle_NoMatch(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	n, err := MatchSingle(tt.elemC, "non_existing")
	assert.Equal(t, ErrNoMatch, err)
	assert.Nil(t, n)
	assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchSingle_SingleMatch(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	n, err := MatchSingle(tt.elemC, "elemC4")
	assert.NoError(t, err)
	assert.True(t, tt.elemC4 == n)
	assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
}

func TestMatchSingle_MoreThanOneMatch(t *testing.T) {
	tt, _, _ := navTestSetup(t)
	caches.XPathExprCache = caches.NewLoadingCache()
	n, err := MatchSingle(tt.elemA, "*")
	assert.Equal(t, ErrMoreThanExpected, err)
	assert.Nil(t, n)
	assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest()))
}

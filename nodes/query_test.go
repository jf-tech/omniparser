package nodes

import (
	"strings"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/stretchr/testify/assert"
)

func TestMatchAll(t *testing.T) {
	s := `
	<AAA>
		<BBB id="1"/>
		<CCC id="2">
			<DDD/>
		</CCC>
		<CCC id="3">
			<DDD/>
		</CCC>
	</AAA>`
	top, err := node.Parse(strings.NewReader(s))
	assert.NoError(t, err)
	assert.NotNil(t, top)

	caches.XPathExprCache = caches.NewLoadingCache()
	assert.Equal(t, 0, len(caches.XPathExprCache.DumpForTest()))

	top, err = MatchSingle(top, "/AAA")
	assert.NoError(t, err)
	assert.NotNil(t, top)
	assert.Equal(t, 1, len(caches.XPathExprCache.DumpForTest())) // "/AAA" added to xpath expr cache.

	n, err := MatchAll(top, "BBB")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(n))
	assert.Equal(t, `<BBB id="1"></BBB>`, n[0].OutputXML(true))
	assert.Equal(t, 2, len(caches.XPathExprCache.DumpForTest())) // "BBB" added to xpath expr cache.

	n, err = MatchAll(top, "CCC", DisableXPathCache)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(n))
	assert.Equal(t, `<CCC id="2"><DDD></DDD></CCC>`, n[0].OutputXML(true))
	assert.Equal(t, `<CCC id="3"><DDD></DDD></CCC>`, n[1].OutputXML(true))
	assert.Equal(t, 2, len(caches.XPathExprCache.DumpForTest())) // "CCC" shouldn't be added to cache.

	n, err = MatchAll(top, "CCC[@id='2']")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(n))
	assert.Equal(t, `<CCC id="2"><DDD></DDD></CCC>`, n[0].OutputXML(true))
	n2, err := MatchAll(n[0], ".")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(n2))
	assert.Equal(t, n[0], n2[0])

	// only one flag can be passed.
	n, err = MatchAll(top, "CCC[@id='2']", 0, 1)
	assert.Error(t, err)
	assert.Equal(t, "only one flag is allowed, instead got: [0 1]", err.Error())
	assert.Nil(t, n)

	// invalid xpath
	n, err = MatchAll(top, "[invalid")
	assert.Error(t, err)
	assert.Equal(t, "xpath '[invalid' compilation failed: expression must evaluate to a node-set", err.Error())
	assert.Nil(t, n)
}

func TestMatchSingle(t *testing.T) {
	s := `
	<AAA>
		<BBB id="1"/>
		<CCC id="2">
			<DDD/>
		</CCC>
		<CCC id="3">
			<DDD/>
		</CCC>
	</AAA>`
	top, err := node.Parse(strings.NewReader(s))
	assert.NoError(t, err)
	assert.NotNil(t, top)

	n, err := MatchSingle(top, "[invalid")
	assert.Error(t, err)
	assert.Equal(t, "xpath '[invalid' compilation failed: expression must evaluate to a node-set", err.Error())
	assert.Nil(t, n)

	n, err = MatchSingle(top, "/NON_EXISTING")
	assert.Equal(t, ErrNoMatch, err)
	assert.Nil(t, n)

	n, err = MatchSingle(top, "/AAA/CCC")
	assert.Equal(t, ErrMoreThanExpected, err)
	assert.Nil(t, n)

	n, err = MatchSingle(top, "/AAA/CCC[@id=2]")
	assert.NoError(t, err)
	assert.Equal(t, `<CCC id="2"><DDD></DDD></CCC>`, n.OutputXML(true))
}

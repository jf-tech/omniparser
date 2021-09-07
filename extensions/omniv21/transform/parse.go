package transform

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

type parseCtx struct {
	transformCtx          *transformctx.Ctx
	customFuncs           customfuncs.CustomFuncs
	customParseFuncs      CustomParseFuncs // Deprecated.
	disableTransformCache bool             // by default, we have caching on. only in some tests we turn caching off.
	transformCache        map[string]interface{}
}

// NewParseCtx creates new context for parsing and transforming a *Node (and its sub-tree) into an output record.
func NewParseCtx(
	transformCtx *transformctx.Ctx,
	customFuncs customfuncs.CustomFuncs,
	customParseFuncs CustomParseFuncs) *parseCtx {
	return &parseCtx{
		transformCtx:          transformCtx,
		customFuncs:           customFuncs,
		customParseFuncs:      customParseFuncs,
		disableTransformCache: false,
		transformCache:        map[string]interface{}{},
	}
}

func (p *parseCtx) ParseNode(n *idr.Node, decl *Decl) (interface{}, error) {
	var cacheKey string
	if !p.disableTransformCache {
		cacheKey = strconv.FormatInt(n.ID, 16) + "/" + decl.hash
		if cacheValue, found := p.transformCache[cacheKey]; found {
			return cacheValue, nil
		}
	}
	saveIntoCache := func(value interface{}, err error) (interface{}, error) {
		if !p.disableTransformCache {
			if err != nil {
				return value, err
			}
			p.transformCache[cacheKey] = value
		}
		return value, err
	}
	switch decl.kind {
	case kindConst:
		return saveIntoCache(p.parseConst(decl))
	case kindExternal:
		return saveIntoCache(p.parseExternal(decl))
	case kindField:
		return saveIntoCache(p.parseField(n, decl))
	case kindObject:
		return saveIntoCache(p.parseObject(n, decl))
	case kindArray:
		return saveIntoCache(p.parseArray(n, decl))
	case kindCustomFunc:
		return saveIntoCache(p.parseCustomFunc(n, decl))
	case kindCustomParse:
		return saveIntoCache(p.parseCustomParse(n, decl))
	default:
		return nil, fmt.Errorf("unexpected decl kind '%s' on '%s'", decl.kind, decl.fqdn)
	}
}

func (p *parseCtx) parseConst(decl *Decl) (interface{}, error) {
	return normalizeAndReturnValue(decl, *decl.Const)
}

func (p *parseCtx) parseExternal(decl *Decl) (interface{}, error) {
	if v, found := p.transformCtx.External(*decl.External); found {
		return normalizeAndReturnValue(decl, v)
	}
	return nil, fmt.Errorf("cannot find external property '%s' on '%s'", *decl.External, decl.fqdn)
}

func xpathQueryNeeded(decl *Decl) bool {
	// For a given transform, we only do xpath query, if
	// - it has "xpath" or "xpath_dynamic" defined in its decl AND
	// - it is not a child of array decl.
	// The second condition is because for array's child transform, the xpath query is done at array level.
	// See details in parseArray().
	// Now, if the transform is FINAL_OUTPUT, we never do xpath query on that, FINAL_OUTPUT's content node
	// is always supplied by reader.
	return decl.fqdn != finalOutput &&
		decl.isXPathSet() &&
		(decl.parent == nil || decl.parent.kind != kindArray)
}

func (p *parseCtx) computeXPath(n *idr.Node, decl *Decl) (xpath string, dynamic bool, err error) {
	switch {
	case strs.IsStrPtrNonBlank(decl.XPath):
		xpath, dynamic, err = *(decl.XPath), false, nil
	case decl.XPathDynamic != nil:
		dynamic = true
		xpath, err = p.computeXPathDynamic(n, decl.XPathDynamic)
	default:
		xpath, dynamic, err = ".", false, nil
	}
	return xpath, dynamic, err
}

func (p *parseCtx) computeXPathDynamic(n *idr.Node, xpathDynamicDecl *Decl) (string, error) {
	v, err := p.ParseNode(n, xpathDynamicDecl)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", fmt.Errorf("xpath_dynamic on '%s' yields empty value", xpathDynamicDecl.fqdn)
	}
	if reflect.ValueOf(v).Kind() != reflect.String {
		return "", fmt.Errorf("xpath_dynamic on '%s' yields a non-string value '%v'", xpathDynamicDecl.fqdn, v)
	}
	xpathDynamic := v.(string)
	if !strs.IsStrNonBlank(xpathDynamic) {
		return "", fmt.Errorf("xpath_dynamic on '%s' yields empty value", xpathDynamicDecl.fqdn)
	}
	return xpathDynamic, nil
}

func xpathMatchFlags(dynamic bool) uint {
	if dynamic {
		return idr.DisableXPathCache
	}
	return 0
}

func (p *parseCtx) querySingleNodeFromXPath(n *idr.Node, decl *Decl) (*idr.Node, error) {
	if !xpathQueryNeeded(decl) {
		return n, nil
	}
	xpath, dynamic, err := p.computeXPath(n, decl)
	if err != nil {
		return nil, nil
	}
	resultNode, err := idr.MatchSingle(n, xpath, xpathMatchFlags(dynamic))
	switch {
	case err == idr.ErrNoMatch:
		return nil, nil
	case err == idr.ErrMoreThanExpected:
		return nil, fmt.Errorf("xpath query '%s' on '%s' yielded more than one result", xpath, decl.fqdn)
	case err != nil:
		return nil, fmt.Errorf("xpath query '%s' on '%s' failed: %s", xpath, decl.fqdn, err.Error())
	}
	return resultNode, nil
}

func (p *parseCtx) parseField(n *idr.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	return normalizeAndReturnValue(decl, n.InnerText())
}

func (p *parseCtx) parseCustomFunc(n *idr.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	funcResult, err := p.invokeCustomFunc(n, decl.CustomFunc)
	if err != nil {
		return nil, err
	}
	return normalizeAndReturnValue(decl, funcResult)
}

func (p *parseCtx) parseCustomParse(n *idr.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	v, err := p.invokeCustomParse(p.customParseFuncs[*decl.CustomParse], n)
	if err != nil {
		return nil, err
	}
	return normalizeAndReturnValue(decl, v)
}

func (p *parseCtx) parseObject(n *idr.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	obj := map[string]interface{}{}
	for _, childDecl := range decl.children {
		childValue, err := p.ParseNode(n, childDecl)
		if err != nil {
			return nil, err
		}
		// value returned by p.ParseNode is already normalized, thus this
		// normalizeAndSaveValue won't fail.
		_ = normalizeAndSaveValue(childDecl, childValue, func(normalizedValue interface{}) {
			obj[strs.LastNameletOfFQDNWithEsc(childDecl.fqdn)] = normalizedValue
		})
	}
	return normalizeAndReturnValue(decl, obj)
}

func (p *parseCtx) parseArray(n *idr.Node, decl *Decl) (interface{}, error) {
	var array []interface{}
	for _, childDecl := range decl.children {
		// if a particular child Decl has xpath, then we'll multi-select nodes based on that
		// xpath, transform each of the nodes based on the child Decl, and save to the array.
		// if a particular child Decl has no xpath, then we'll simply use its parent n, i.e.
		// the current n, and do child Decl transform and save to the array.
		// Note computeXPath() already does this for us: if xpath/xpath_dynamic both null, it
		// returns xpath "." which gives us the current node when we use it to query the current
		// node.
		xpath, dynamic, err := p.computeXPath(n, childDecl)
		if err != nil {
			continue
		}
		childNodes, err := idr.MatchAll(n, xpath, xpathMatchFlags(dynamic))
		if err != nil {
			return nil, fmt.Errorf("xpath query '%s' on '%s' failed: %s", xpath, childDecl.fqdn, err.Error())
		}
		for _, childNode := range childNodes {
			childValue, err := p.ParseNode(childNode, childDecl)
			if err != nil {
				return nil, err
			}
			// value returned by p.ParseNode is already normalized, thus this
			// normalizeAndSaveValue won't fail.
			_ = normalizeAndSaveValue(childDecl, childValue, func(normalizedValue interface{}) {
				array = append(array, normalizedValue)
			})
		}
	}
	return normalizeAndReturnValue(decl, array)
}

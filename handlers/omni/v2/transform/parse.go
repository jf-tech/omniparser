package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	node "github.com/antchfx/xmlquery"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/nodes"
	"github.com/jf-tech/omniparser/transformctx"
)

type parseCtx struct {
	opCtx                 *transformctx.Ctx
	customFuncs           customfuncs.CustomFuncs
	customParseFuncs      CustomParseFuncs
	disableTransformCache bool // by default we have caching on. only in some tests we turn caching off.
	transformCache        map[string]interface{}
}

// NewParseCtx creates new context for parsing a *Node (and its sub-tree) into an output record.
func NewParseCtx(
	opCtx *transformctx.Ctx,
	customFuncs customfuncs.CustomFuncs,
	customParseFuncs CustomParseFuncs) *parseCtx {
	return &parseCtx{
		opCtx:                 opCtx,
		customFuncs:           customFuncs,
		customParseFuncs:      customParseFuncs,
		disableTransformCache: false,
		transformCache:        map[string]interface{}{},
	}
}

func nodePtrAddrStr(n *node.Node) string {
	// `uintptr` is faster than `fmt.Sprintf("%p"...)`
	return strconv.FormatUint(uint64(uintptr(unsafe.Pointer(n))), 16)
}

func resultTypeConversion(decl *Decl, value string) (interface{}, error) {
	if decl.resultType() == ResultTypeString {
		return value, nil
	}
	// after this point, result type isn't of string.
	// Omit the field in final result if it is empty with non-string type.
	if !strs.IsStrNonBlank(value) {
		return nil, nil
	}
	switch decl.resultType() {
	case ResultTypeInt:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return int64(f), nil
	case ResultTypeFloat:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	case ResultTypeBoolean:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return b, nil
	default:
		return value, nil
	}
}

func normalizeAndSaveValue(decl *Decl, value interface{}, save func(interface{})) error {
	if value == nil {
		if decl.KeepEmptyOrNull {
			save(nil)
		}
		return nil
	}
	// Now value != nil
	switch reflect.ValueOf(value).Kind() {
	case reflect.String:
		strValue := value.(string)
		if !decl.KeepLeadingTrailingSpace {
			strValue = strings.TrimSpace(strValue)
		}
		if strValue == "" && !decl.KeepEmptyOrNull {
			return nil
		}
		typedResult, err := resultTypeConversion(decl, strValue)
		if err != nil {
			return fmt.Errorf("fail to convert value '%s' to type '%s' on '%s', err: %s",
				strValue, decl.resultType(), decl.fqdn, err.Error())
		}
		if typedResult != nil || decl.KeepEmptyOrNull {
			save(typedResult)
		}
		return nil
	case reflect.Slice:
		if len(value.([]interface{})) > 0 || decl.KeepEmptyOrNull {
			save(value)
		}
		return nil
	case reflect.Map:
		if len(value.(map[string]interface{})) > 0 || decl.KeepEmptyOrNull {
			save(value)
		}
		return nil
	default:
		save(value)
		return nil
	}
}

func normalizeAndReturnValue(decl *Decl, value interface{}) (interface{}, error) {
	var returnValue interface{}
	err := normalizeAndSaveValue(decl, value, func(normalizedValue interface{}) {
		returnValue = normalizedValue
	})
	if err != nil {
		return nil, err
	}
	return returnValue, nil
}

func (p *parseCtx) ParseNode(n *node.Node, decl *Decl) (interface{}, error) {
	var cacheKey string
	if !p.disableTransformCache {
		cacheKey = nodePtrAddrStr(n) + "/" + decl.hash
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
	case KindConst:
		return saveIntoCache(p.parseConst(decl))
	case KindExternal:
		return saveIntoCache(p.parseExternal(decl))
	case KindField:
		return saveIntoCache(p.parseField(n, decl))
	case KindObject:
		return saveIntoCache(p.parseObject(n, decl))
	case KindArray:
		return saveIntoCache(p.parseArray(n, decl))
	case KindCustomFunc:
		return saveIntoCache(p.parseCustomFunc(n, decl))
	case KindCustomParse:
		return saveIntoCache(p.parseCustomParse(n, decl))
	default:
		return nil, fmt.Errorf("unexpected decl kind '%s' on '%s'", decl.kind, decl.fqdn)
	}
}

func (p *parseCtx) parseConst(decl *Decl) (interface{}, error) {
	return normalizeAndReturnValue(decl, *decl.Const)
}

func (p *parseCtx) parseExternal(decl *Decl) (interface{}, error) {
	if v, found := p.opCtx.External(*decl.External); found {
		return normalizeAndReturnValue(decl, v)
	}
	return "", fmt.Errorf("cannot find external property '%s' on '%s'", *decl.External, decl.fqdn)
}

func xpathQueryNeeded(decl *Decl) bool {
	// For a given transform, we only do xpath query, if
	// - it has "xpath" or "xpath_dynamic" defined in its decl AND
	// - it is not a child of array decl.
	// The second condition is because for array's child transform, the xpath query is done at array level.
	// See details in parseArray().
	// Now, if the transform is FINAL_OUTPUT, we never do xpath query on that, FINAL_OUTPUT's content node
	// is always supplied by reader.
	return decl.fqdn != FinalOutput &&
		decl.isXPathSet() &&
		(decl.parent == nil || decl.parent.kind != KindArray)
}

func (p *parseCtx) computeXPath(n *node.Node, decl *Decl) (xpath string, dynamic bool, err error) {
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

func (p *parseCtx) computeXPathDynamic(n *node.Node, xpathDynamicDecl *Decl) (string, error) {
	v, err := p.ParseNode(n, xpathDynamicDecl)
	if err != nil {
		return "", err
	}
	// if v is straight out nil, then we should fail out
	// if v isn't nil, it could be an interface{} type whose value is nil; or it could be some valid values.
	// note we need to guard the IsNil call as it would panic if v kind isn't interface/chan/func/map/slice/ptr.
	// note we only need to ensure for kind == interface, because  ParseNode will never return
	// chan/func/ptr. It's possible to return map/slice, but in earlier validation (validateXPath) we already
	// ensured `xpath_dynamic` result type is string.
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Interface && reflect.ValueOf(v).IsNil()) {
		return "", fmt.Errorf("'%s' failed to yield a single value: no node matched", xpathDynamicDecl.fqdn)
	}
	return v.(string), nil
}

func xpathMatchFlags(dynamic bool) uint {
	if dynamic {
		return nodes.DisableXPathCache
	}
	return 0
}

func (p *parseCtx) querySingleNodeFromXPath(n *node.Node, decl *Decl) (*node.Node, error) {
	if !xpathQueryNeeded(decl) {
		return n, nil
	}
	xpath, dynamic, err := p.computeXPath(n, decl)
	if err != nil {
		return nil, nil
	}
	resultNode, err := nodes.MatchSingle(n, xpath, xpathMatchFlags(dynamic))
	switch {
	case err == nodes.ErrNoMatch:
		return nil, nil
	case err == nodes.ErrMoreThanExpected:
		return nil, fmt.Errorf("xpath query '%s' on '%s' yielded more than one result", xpath, decl.fqdn)
	case err != nil:
		return nil, fmt.Errorf("xpath query '%s' on '%s' failed: %s", xpath, decl.fqdn, err.Error())
	}
	return resultNode, nil
}

func (p *parseCtx) parseField(n *node.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return normalizeAndReturnValue(decl, nil)
	}
	if decl.resultType() == ResultTypeObject && n.Type == node.ElementNode {
		// When a field's result_type is marked "object", we'll simply copy the selected
		// node and all its children over directly. Note it doesn't/won't work pretty with
		// XML input files, as XML might contains attributes, which can't really nicely
		// translate into map[string]interface{}. All other file format types
		// (csv/edi/fixed-length/json) are fine. Well, so be it the limitation.
		return normalizeAndReturnValue(decl, nodes.J2NodeToInterface(n))
	}
	return normalizeAndReturnValue(decl, n.InnerText())
}

func (p *parseCtx) parseCustomFunc(n *node.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return normalizeAndReturnValue(decl, nil)
	}
	funcValue, err := p.invokeCustomFunc(n, decl.CustomFunc)
	if err != nil {
		return nil, err
	}
	if decl.resultType() == ResultTypeObject && funcValue != "" {
		var obj interface{}
		if err := json.Unmarshal([]byte(funcValue), &obj); err != nil {
			return nil, err
		}
		return normalizeAndReturnValue(decl, obj)
	}
	return normalizeAndReturnValue(decl, funcValue)
}

func (p *parseCtx) parseCustomParse(n *node.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return normalizeAndReturnValue(decl, nil)
	}
	v, err := p.invokeCustomParse(p.customParseFuncs[*decl.CustomParse], n)
	if err != nil {
		return nil, err
	}
	return normalizeAndReturnValue(decl, v)
}

func (p *parseCtx) parseObject(n *node.Node, decl *Decl) (interface{}, error) {
	n, err := p.querySingleNodeFromXPath(n, decl)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return normalizeAndReturnValue(decl, nil)
	}
	object := map[string]interface{}{}
	for _, childDecl := range decl.children {
		childValue, err := p.ParseNode(n, childDecl)
		if err != nil {
			return nil, err
		}
		// value returned by p.ParseNode is already normalized, thus this
		// normalizeAndSaveValue won't fail.
		_ = normalizeAndSaveValue(childDecl, childValue, func(normalizedValue interface{}) {
			object[strs.LastNameletOfFQDN(childDecl.fqdn)] = normalizedValue
		})
	}
	return normalizeAndReturnValue(decl, object)
}

func (p *parseCtx) parseArray(n *node.Node, decl *Decl) (interface{}, error) {
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
		childNodes, err := nodes.MatchAll(n, xpath, xpathMatchFlags(dynamic))
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

package customfuncs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	node "github.com/antchfx/xmlquery"
	"github.com/dop251/goja"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/nodes"
	"github.com/jf-tech/omniparser/transformctx"
)

const (
	argTypeString  = "string"
	argTypeInt     = "int"
	argTypeFloat   = "float"
	argTypeBoolean = "boolean"
)

const (
	argNameNode = "_node"
)

func parseArgTypeAndValue(argDecl, argValue string) (name string, value interface{}, err error) {
	col := strings.IndexRune(argDecl, ':')
	if col < 0 {
		return "", nil, fmt.Errorf(
			"arg decl must be in format of '<arg_name>:<arg_type>', instead got '%s'", argDecl)
	}
	name = argDecl[:col]
	if !strs.IsStrNonBlank(name) {
		return "", nil, fmt.Errorf(
			"arg_name in '<arg_name>:<arg_type>' cannot be a blank string, instead got '%s'", argDecl)
	}
	switch argDecl[col+1:] {
	case argTypeString:
		return name, argValue, nil
	case argTypeInt:
		f, err := strconv.ParseFloat(argValue, 64)
		if err != nil {
			return "", nil, err
		}
		return name, int64(f), nil
	case argTypeFloat:
		f, err := strconv.ParseFloat(argValue, 64)
		if err != nil {
			return "", nil, err
		}
		return name, f, nil
	case argTypeBoolean:
		b, err := strconv.ParseBool(argValue)
		if err != nil {
			return "", nil, err
		}
		return name, b, nil
	default:
		return "", nil, fmt.Errorf("arg_type '%s' in '<arg_name>:<arg_type>' is not supported", argDecl[col+1:])
	}
}

// JSProgramCache caches *goja.Program. A *goja.Program is compiled javascript and it can be used
// across multiple goroutines and across different *goja.Runtime. If default loading cache capacity
// is not desirable, change JSProgramCache to a loading cache with a different capacity at package
// init time. Be mindful this will be shared across all use cases inside your process.
var JSProgramCache *caches.LoadingCache

// jsRuntimePool caches *goja.Runtime whose creation is expensive such that we want to have a pool
// of them to amortize the initialization cost. However, a *goja.Runtime cannot be used by two/more
// javascript's at the same time, thus the use of sync.Pool. Not user customizable.
var jsRuntimePool sync.Pool

// NodeToJSONCache caches *node.Node to JSON translations.
var NodeToJSONCache *caches.LoadingCache

// For debugging/testing purpose so we can easily disable all the caches. But not exported. We always
// want caching in production.
var disableCaching = false

func resetCaches() {
	JSProgramCache = caches.NewLoadingCache()
	jsRuntimePool = sync.Pool{
		New: func() interface{} {
			return goja.New()
		},
	}
	NodeToJSONCache = caches.NewLoadingCache()
}

func init() {
	resetCaches()
}

func getProgram(js string) (*goja.Program, error) {
	if disableCaching {
		return goja.Compile("", js, false)
	}
	p, err := JSProgramCache.Get(js, func(interface{}) (interface{}, error) {
		return goja.Compile("", js, false)
	})
	if err != nil {
		return nil, err
	}
	return p.(*goja.Program), nil
}

func getNodeJSON(n *node.Node) string {
	if disableCaching {
		return nodes.JSONify2(n)
	}
	// TODO if in the future we have *node.Node allocation recycling, then this by-addr caching
	// won't work. Ideally, we should have a node ID which refreshes upon recycling.
	addr := strconv.FormatUint(uint64(uintptr(unsafe.Pointer(n))), 16)
	j, _ := NodeToJSONCache.Get(addr, func(interface{}) (interface{}, error) {
		return nodes.JSONify2(n), nil
	})
	return j.(string)
}

func execProgram(program *goja.Program, args map[string]interface{}) (goja.Value, error) {
	var vm *goja.Runtime
	var poolObj interface{}
	if disableCaching {
		vm = goja.New()
	} else {
		poolObj = jsRuntimePool.Get()
		vm = poolObj.(*goja.Runtime)
	}
	defer func() {
		if vm != nil {
			// wipe out all the args in prep for next exec.
			for arg := range args {
				_ = vm.GlobalObject().Delete(arg)
			}
		}
		if poolObj != nil {
			jsRuntimePool.Put(poolObj)
		}
	}()
	for arg, val := range args {
		vm.Set(arg, val)
	}
	return vm.RunProgram(program)
}

// javascriptWithContext is a custom_func that runs a javascript with optional arguments and
// with current node JSON, if node is provided.
func javascriptWithContext(ctx *transformctx.Ctx, n *node.Node, js string, args ...string) (string, error) {
	if len(args)%2 != 0 {
		return "", errors.New("invalid number of args to 'javascript'")
	}
	program, err := getProgram(js)
	if err != nil {
		return "", fmt.Errorf("invalid javascript: %s", err.Error())
	}
	vmArgs := make(map[string]interface{})
	for i := 0; i < len(args)/2; i++ {
		argName, val, err := parseArgTypeAndValue(args[i*2], args[i*2+1])
		if err != nil {
			return "", err
		}
		vmArgs[argName] = val
	}
	if n != nil {
		vmArgs[argNameNode] = getNodeJSON(n)
	}
	v, err := execProgram(program, vmArgs)
	if err != nil {
		return "", err
	}
	switch {
	case goja.IsNaN(v), goja.IsInfinity(v), goja.IsNull(v), goja.IsUndefined(v):
		return "", fmt.Errorf("result is %s", v.String())
	default:
		return v.String(), nil
	}
}

// javascript is a custom_func that runs a javascript with optional arguments and without context
// node JSON provided.
func javascript(ctx *transformctx.Ctx, js string, args ...string) (string, error) {
	return javascriptWithContext(ctx, nil, js, args...)
}

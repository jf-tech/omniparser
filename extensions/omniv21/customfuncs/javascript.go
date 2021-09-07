package customfuncs

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/jf-tech/go-corelib/caches"

	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

const (
	argNameNode = "_node"
)

// JSProgramCache caches *goja.Program. A *goja.Program is compiled javascript, and it can be used
// across multiple goroutines and across different *goja.Runtime. If default loading cache capacity
// is not desirable, change JSProgramCache to a loading cache with a different capacity at package
// init time. Be mindful this will be shared across all use cases inside your process.
var JSProgramCache *caches.LoadingCache

// jsRuntimePool caches *goja.Runtime whose creation is expensive such that we want to have a pool
// of them to amortize the initialization cost. However, a *goja.Runtime cannot be used by two/more
// javascript's at the same time, thus the use of sync.Pool. Not user customizable.
var jsRuntimePool sync.Pool

// NodeToJSONCache caches *idr.Node to JSON translations.
var NodeToJSONCache *caches.LoadingCache

// For debugging/testing purpose, so we can easily disable all the caches. But not exported. We always
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

func getNodeJSON(n *idr.Node) string {
	if disableCaching {
		return idr.JSONify2(n)
	}
	j, _ := NodeToJSONCache.Get(n.ID, func(interface{}) (interface{}, error) {
		return idr.JSONify2(n), nil
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

// JavaScriptWithContext is a custom_func that runs a javascript with optional arguments and
// with contextual '_node' JSON, if idr.Node is provided.
func JavaScriptWithContext(_ *transformctx.Ctx, n *idr.Node, js string, args ...interface{}) (interface{}, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("number of args must be even, but got %d", len(args))
	}
	program, err := getProgram(js)
	if err != nil {
		return nil, fmt.Errorf("invalid javascript: %s", err.Error())
	}
	vmArgs := make(map[string]interface{})
	for i := 0; i < len(args)/2; i++ {
		vmArgs[args[i*2].(string)] = args[i*2+1]
	}
	if n != nil {
		vmArgs[argNameNode] = getNodeJSON(n)
	}
	v, err := execProgram(program, vmArgs)
	if err != nil {
		return nil, err
	}
	switch {
	case goja.IsNaN(v), goja.IsInfinity(v), goja.IsNull(v), goja.IsUndefined(v):
		return nil, fmt.Errorf("result is %s", v.String())
	default:
		return v.Export(), nil
	}
}

// JavaScript is a custom_func that runs a javascript with optional arguments and without contextual
// '_node' JSON provided.
func JavaScript(ctx *transformctx.Ctx, js string, args ...interface{}) (interface{}, error) {
	return JavaScriptWithContext(ctx, nil, js, args...)
}

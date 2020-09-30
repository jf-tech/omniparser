package customfuncs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	node "github.com/antchfx/xmlquery"
	"github.com/dop251/goja"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"
	pool "github.com/jolestar/go-commons-pool"

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
	declParts := strings.Split(argDecl, ":")
	if len(declParts) != 2 {
		return "", nil, fmt.Errorf(
			"arg decl must be in format of '<arg_name>:<arg_type>', instead got '%s'", argDecl)
	}
	name = declParts[0]
	if !strs.IsStrNonBlank(name) {
		return "", nil, fmt.Errorf(
			"arg_name in '<arg_name>:<arg_type>' cannot be a blank string, instead got '%s'", argDecl)
	}
	switch declParts[1] {
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
		return "", nil, fmt.Errorf("arg_type '%s' in '<arg_name>:<arg_type>' is not supported", declParts[1])
	}
}

// For debugging/testing purpose so we can easily disable all the caches. But not exported. We always
// want caching in production.
var disableCaching = false

func resetCaches() {
	// per schema so won't have too many, no need to put a small cap (default is 64k)
	JSProgramCache = caches.NewLoadingCache()
	JSRuntimePool = pool.NewObjectPool(
		context.Background(),
		pool.NewPooledObjectFactorySimple(
			func(context.Context) (interface{}, error) {
				return goja.New(), nil
			}),
		func() *pool.ObjectPoolConfig {
			cfg := pool.NewDefaultPoolConfig()
			cfg.MaxTotal = -1 // no limit
			cfg.MaxIdle = -1  // no limit
			// keep a minimum 10 idling VMs always around to reduce startup latency.
			cfg.MinIdle = 10
			cfg.TimeBetweenEvictionRuns = cfg.MinEvictableIdleTime // turn on eviction
			return cfg
		}())
	// per transform, plus expensive, a smaller cap.
	NodeToJSONCache = caches.NewLoadingCache(100)
}

// JSProgramCache caches *goja.Program. A *goja.Program is compiled javascript and it can be used
// across multiple goroutines and across different *goja.Runtime. If default loading cache capacity
// is not desirable, change JSProgramCache to a loading cache with a different capacity at package
// init time. Be mindful this will be shared across all use cases inside your process.
var JSProgramCache *caches.LoadingCache

// JSRuntimePool caches *goja.Runtime whose creation is expensive such that we want to have a pool
// of them to amortize the initialization cost. However, a *goja.Runtime cannot be used by two/more
// javascript's at the same time, thus the use of resource pool. If the default pool configuration is
// not desirable, change JSRuntimePool with a different config during package init time. Be mindful
// this will be shared across all use cases inside your process.
var JSRuntimePool *pool.ObjectPool

// NodeToJSONCache caches *node.Node to JSON translations.
var NodeToJSONCache *caches.LoadingCache

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
	if disableCaching {
		vm = goja.New()
	} else {
		poolObj, err := JSRuntimePool.BorrowObject(context.Background())
		if err != nil {
			vm = goja.New()
		} else {
			vm = poolObj.(*goja.Runtime)
		}
	}
	defer func() {
		if vm != nil {
			// wipe out all the args (by setting them to undefined) in prep for next exec.
			for arg := range args {
				vm.Set(arg, goja.Undefined())
			}
			_ = JSRuntimePool.ReturnObject(context.Background(), vm)
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

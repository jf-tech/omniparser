package customfuncs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
var disableCache = false

// JSProgramCache caches *goja.Program. A *goja.Program is compiled javascript and it can be used
// across multiple goroutines and across different *goja.Runtime.
var JSProgramCache = caches.NewLoadingCache() // per schema so won't have too many, no need to put a hard cap.
// JSRuntimeCache caches *goja.Runtime. A *goja.Runtime is a javascript VM. It can *not* be shared
// across multiple goroutines.
var JSRuntimeCache = caches.NewLoadingCache(100) // per transform, plus expensive, a smaller cap.
// NodeToJSONCache caches *node.Node tree to translated JSON string.
var NodeToJSONCache = caches.NewLoadingCache(100) // per transform, plus expensive, a smaller cap.

func getProgram(js string) (*goja.Program, error) {
	if disableCache {
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

func ptrAddrStr(p unsafe.Pointer) string {
	return strconv.FormatUint(uint64(uintptr(p)), 16)
}

func getRuntime(ctx *transformctx.Ctx) *goja.Runtime {
	if disableCache {
		return goja.New()
	}
	// a VM can be reused as long as not across thread. We don't have access to
	// thread/goroutine id (nor do we want to use some hack to get it, see
	// https://golang.org/doc/faq#no_goroutine_id). Instead, we use ctx as an
	// indicator - omniparser runs on a single thread per transform. And ctx is
	// is per transform.
	addr := ptrAddrStr(unsafe.Pointer(ctx))
	vm, _ := JSRuntimeCache.Get(addr, func(interface{}) (interface{}, error) {
		return goja.New(), nil
	})
	return vm.(*goja.Runtime)
}

func getNodeJSON(n *node.Node) string {
	if disableCache {
		return nodes.JSONify2(n)
	}
	addr := ptrAddrStr(unsafe.Pointer(n))
	j, _ := NodeToJSONCache.Get(addr, func(interface{}) (interface{}, error) {
		return nodes.JSONify2(n), nil
	})
	return j.(string)
}

// javascriptWithContext is a custom_func that runs a javascript with optional arguments and
// with current node JSON, if the context node is provided.
func javascriptWithContext(ctx *transformctx.Ctx, n *node.Node, js string, args ...string) (string, error) {
	if len(args)%2 != 0 {
		return "", errors.New("invalid number of args to 'javascript'")
	}
	program, err := getProgram(js)
	if err != nil {
		return "", fmt.Errorf("invalid javascript: %s", err.Error())
	}
	runtime := getRuntime(ctx)
	var varnames []string
	defer func() {
		for i := range varnames {
			runtime.Set(varnames[i], goja.Undefined())
		}
	}()
	for i := 0; i < len(args)/2; i++ {
		varname, val, err := parseArgTypeAndValue(args[i*2], args[i*2+1])
		if err != nil {
			return "", err
		}
		runtime.Set(varname, val)
		varnames = append(varnames, varname)
	}
	if n != nil {
		runtime.Set(argNameNode, getNodeJSON(n))
	}
	v, err := runtime.RunProgram(program)
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

package customfuncs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	node "github.com/antchfx/xmlquery"
	"github.com/dop251/goja"

	"github.com/jf-tech/omniparser/omniparser/nodes"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/strs"
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

func javascript(_ *transformctx.Ctx, n *node.Node, js string, args ...string) (string, error) {
	if len(args)%2 != 0 {
		return "", errors.New("invalid number of args to 'javascript'")
	}
	vm := goja.New()
	vm.Set(argNameNode, nodes.JSONify2(n))
	for i := 0; i < len(args)/2; i++ {
		n, v, err := parseArgTypeAndValue(args[i*2], args[i*2+1])
		if err != nil {
			return "", err
		}
		vm.Set(n, v)
	}
	v, err := vm.RunString(js)
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

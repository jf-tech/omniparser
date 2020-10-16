package transform

import (
	"fmt"
	"reflect"

	"github.com/jf-tech/omniparser/idr"
)

func (p *parseCtx) invokeCustomFunc(n *idr.Node, customFuncDecl *CustomFuncDecl) (interface{}, error) {
	// In validation, we've validated the custom func exists.
	fn, _ := p.customFuncs[customFuncDecl.Name]
	fnType := reflect.TypeOf(fn)
	argValues, err := p.prepArgValues(n, customFuncDecl, fnType)
	if err != nil {
		return nil, err
	}
	result := reflect.ValueOf(fn).Call(argValues)
	// result[0] - result from custom function
	// result[1] - error from custom function
	if result[1].Interface() == nil {
		return result[0].Interface(), nil
	}
	if customFuncDecl.IgnoreError {
		return nil, nil
	}
	return nil, fmt.Errorf("'%s' failed: %s", customFuncDecl.fqdn, result[1].Interface().(error).Error())
}

func (p *parseCtx) prepArgValues(
	n *idr.Node, customFuncDecl *CustomFuncDecl, fnType reflect.Type) ([]reflect.Value, error) {

	argVals := make([]reflect.Value, 0, 2+len(customFuncDecl.Args))
	// All custom_func's has 0-th arg as *transformctx.Ctx
	argVals = append(argVals, reflect.ValueOf(p.transformCtx))
	// Some newer custom_func's can have *idr.Node as secondary default arg.
	fnArgIndex := 1
	if fnType.NumIn() >= 2 && fnType.In(1) == reflect.TypeOf((*idr.Node)(nil)) {
		argVals = append(argVals, reflect.ValueOf(n))
		fnArgIndex = 2
	}
	// Now process all normal args to the custom_func.
	for _, argDecl := range customFuncDecl.Args {
		val, err := p.ParseNode(n, argDecl)
		if err != nil {
			return nil, err
		}
		if val == nil {
			argVals = append(argVals, reflect.Zero(getFuncArgType(fnType, fnArgIndex)))
		} else {
			argVals = append(argVals, reflect.ValueOf(val))
		}
		fnArgIndex++
	}
	return argVals, nil
}

func getFuncArgType(fnType reflect.Type, argIndex int) reflect.Type {
	isVariadic := fnType.IsVariadic()
	if argIndex >= fnType.NumIn() {
		argIndex = fnType.NumIn() - 1
	}
	typ := fnType.In(argIndex)
	if isVariadic {
		typ = typ.Elem()
	}
	return typ
}

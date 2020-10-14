package transform

import (
	"fmt"
	"reflect"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/idr"
)

func (p *parseCtx) invokeCustomFunc(n *idr.Node, customFuncDecl *CustomFuncDecl) (interface{}, error) {
	// In validation, we've validated the custom func exists.
	fn, _ := p.customFuncs[customFuncDecl.Name]
	argValues, err := p.prepArgValues(n, customFuncDecl, fn)
	if err != nil {
		return "", err
	}

	result := reflect.ValueOf(fn).Call(argValues)

	// result[0] - result from custom function
	// result[1] - error from custom function
	if result[1].Interface() == nil {
		return result[0].Interface(), nil
	}
	err = result[1].Interface().(error)

	if customFuncDecl.ignoreError() {
		return nil, nil
	}

	return "", fmt.Errorf("'%s' failed: %s", customFuncDecl.fqdn, err.Error())
}

func (p *parseCtx) prepArgValues(
	n *idr.Node, customFuncDecl *CustomFuncDecl, fn customfuncs.CustomFuncType) ([]reflect.Value, error) {

	fnType := reflect.TypeOf(fn)
	argVals := make([]reflect.Value, 0, 2+len(customFuncDecl.Args))

	// All custom_func's has 0-th arg as *transformctx.Ctx
	argVals = append(argVals, reflect.ValueOf(p.opCtx))

	// Some newer custom_func's can have *idr.Node as secondary default arg.
	if fnType.NumIn() >= 2 && fnType.In(1) == reflect.TypeOf((*idr.Node)(nil)) {
		argVals = append(argVals, reflect.ValueOf(n))
	}

	// Now process all normal args to the custom_func.
	for _, argDecl := range customFuncDecl.Args {
		val, err := p.ParseNode(n, argDecl)
		if err != nil {
			return nil, err
		}
		argVals = append(argVals, reflect.ValueOf(val))
	}
	return argVals, nil
}

package transform

import (
	"fmt"
	"reflect"

	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/nodes"
)

func (p *parseCtx) invokeCustomFunc(n *node.Node, customFuncDecl *CustomFuncDecl) (string, error) {
	// In validation, we've validated the custom func exists.
	fn, _ := p.customFuncs[customFuncDecl.Name]
	argValues, err := p.prepCustomFuncArgValues(n, customFuncDecl, fn)
	if err != nil {
		return "", err
	}

	result := reflect.ValueOf(fn).Call(argValues)

	// result[0] - result from custom function
	// result[1] - error from custom function
	if result[1].Interface() == nil {
		return result[0].String(), nil
	}
	err = result[1].Interface().(error)

	if customFuncDecl.IgnoreErrorAndReturnEmptyStr {
		return "", nil
	}

	return "", fmt.Errorf("'%s' failed: %s", customFuncDecl.fqdn, err.Error())
}

func (p *parseCtx) prepCustomFuncArgValues(
	n *node.Node, customFuncDecl *CustomFuncDecl, fn customfuncs.CustomFuncType) ([]reflect.Value, error) {

	argValues := []reflect.Value{reflect.ValueOf(p.opCtx)}
	appendArgValue := func(argDecl *Decl, value interface{}) {
		v, _ := normalizeAndReturnValue(argDecl, value)
		// if v is nil for some reason, e.g:
		// -----------
		// "date": { "custom_func": {
		//     "name": "dateTimeToRfc3339",
		//     "args": [
		//         { "xpath": "DATE" },
		//         { "const": "", "_comment": "input timezone" },
		//         { "const": "", "_comment": "output timezone" }
		//     ]
		// }},
		// -----------
		// In the example above, arg[2] and arg[3] is empty string, which will be converted to nil by
		// normalizeAndReturnValue because keep_empty_or_null isn't specified, which is the typical case
		// for schema author, we always want to use empty string in case of nil value for custom func args.
		switch v {
		case nil:
			v = ""
		default:
			v = v.(string)
		}
		argValues = append(argValues, reflect.ValueOf(v))
	}

	for _, argDecl := range customFuncDecl.Args {
		// We'd love to delegate all the value calculation to ParseNode but here we have
		// one special case, when we deal with a field.
		// We have situations we need to support aggregation func such as sum/avg. In those cases
		// the arg to the custom func can be a field with xpath/xpath_dynamic that we want it to
		// yield multiple values to feed into those agg funcs.
		switch argDecl.kind {
		case KindField:
			xpath, dynamic, err := p.computeXPath(n, argDecl)
			if err != nil {
				return nil, err
			}
			argValueNodes, err := nodes.MatchAll(n, xpath, xpathMatchFlags(dynamic))
			if err != nil {
				return nil, fmt.Errorf("xpath query '%s' for '%s' failed: %s", xpath, argDecl.fqdn, err.Error())
			}
			if reflect.TypeOf(fn).IsVariadic() && len(customFuncDecl.Args) == 1 {
				// Only allow this variable length nodes to args conversion for variadic custom func.
				// and this xpath arg is the **only** argument for this custom func.
				for _, argValueNode := range argValueNodes {
					appendArgValue(argDecl, argValueNode.InnerText())
				}
				break
			}
			// fn is NOT variadic or this xpath arg isn't the only argument for the custom func
			if len(argValueNodes) == 0 {
				// A bit ugly. If the custom func is not variadic or xpath isn't the only arg, and
				// xpath query returned nothing, then use "" empty as the arg value. This is inline
				// with previous logic to reduce regression risk
				appendArgValue(argDecl, "")
				break
			}
			// fn is NOT variadic and xpath query returned at least one value, only use the first one.
			appendArgValue(argDecl, argValueNodes[0].InnerText())
		case KindArray:
			argValue, err := p.ParseNode(n, argDecl)
			if err != nil {
				return nil, err
			}
			if argValue == nil {
				break
			}
			for _, v := range argValue.([]interface{}) {
				appendArgValue(argDecl, v)
			}
		default:
			// Normal case not involving field (so const/external/nested custom_func)
			v, err := p.ParseNode(n, argDecl)
			if err != nil {
				return nil, err
			}
			appendArgValue(argDecl, v)
		}
	}
	return argValues, nil
}

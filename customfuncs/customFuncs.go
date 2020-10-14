package customfuncs

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

// CustomFuncType is the type of a custom function. Has to use interface{} given we support
// non-variadic and variadic functions.
type CustomFuncType = interface{}

// CustomFuncs is a map from custom func names to an actual custom func.
type CustomFuncs = map[string]CustomFuncType

// Merge merges multiple custom func maps into one.
func Merge(funcs ...CustomFuncs) CustomFuncs {
	merged := make(CustomFuncs)
	for _, fs := range funcs {
		for name, f := range fs {
			merged[name] = f
		}
	}
	return merged
}

var builtinPublishedCustomFuncs = map[string]CustomFuncType{
	// keep these custom funcs lexically sorted
	"avg":                     avg,
	"coalesce":                coalesce,
	"concat":                  concat,
	"containsPattern":         containsPattern,
	"copy":                    copyFunc,
	"dateTimeLayoutToRFC3339": dateTimeLayoutToRFC3339,
	"dateTimeToEpoch":         dateTimeToEpoch,
	"dateTimeToRFC3339":       dateTimeToRFC3339,
	"floor":                   floor,
	"ifElse":                  ifElse,
	"isEmpty":                 isEmpty,
	"javascript":              javascript,
	"javascript_with_context": javascriptWithContext,
	"lower":                   lower,
	"splitIntoJsonArray":      splitIntoJsonArray,
	"substring":               substring,
	"sum":                     sum,
	"switch":                  switchFunc,
	"switchByPattern":         switchByPattern,
	"upper":                   upper,
	"uuidv3":                  uuidv3,
}

var builtinHiddenBackCompatCustomFuncs = map[string]CustomFuncType{
	// keep these custom funcs lexically sorted
	"dateTimeToRfc3339":           dateTimeToRFC3339,       // deprecated; use dateTimeToRFC3339.
	"dateTimeWithLayoutToRfc3339": dateTimeLayoutToRFC3339, // deprecated; use dateTimeLayoutToRFC3339.
	"eval":                        eval,                    // deprecated; use javascript.
	"external":                    external,                // deprecated; use "external" decl.
}

// BuiltinCustomFuncs contains all the built-in custom functions.
var BuiltinCustomFuncs = Merge(builtinPublishedCustomFuncs, builtinHiddenBackCompatCustomFuncs)

func coalesce(_ *transformctx.Ctx, strs ...string) (string, error) {
	for _, str := range strs {
		if str != "" {
			return str, nil
		}
	}
	return "", nil
}

func concat(_ *transformctx.Ctx, strs ...string) (string, error) {
	var w strings.Builder
	for _, s := range strs {
		w.WriteString(s)
	}
	return w.String(), nil
}

func containsPattern(_ *transformctx.Ctx, regexPattern string, strs ...string) (string, error) {
	r, err := caches.GetRegex(regexPattern)
	if err != nil {
		return "", err
	}
	for _, str := range strs {
		if r.MatchString(str) {
			return "true", nil
		}
	}
	return "false", nil
}

func copyFunc(_ *transformctx.Ctx, n *idr.Node) (interface{}, error) {
	return idr.J2NodeToInterface(n, true), nil
}

func external(ctx *transformctx.Ctx, name string) (string, error) {
	if v, found := ctx.External(name); found {
		return v, nil
	}
	return "", fmt.Errorf("cannot find external property '%s'", name)
}

func floor(_ *transformctx.Ctx, value, decimalPlaces string) (string, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", fmt.Errorf("unable to parse value '%s' to float64: %s", value, err.Error())
	}
	dp, err := strconv.Atoi(decimalPlaces)
	if err != nil {
		return "", fmt.Errorf("unable to parse decimal place value '%s' to int: %s", decimalPlaces, err.Error())
	}
	if dp < 0 || dp > 100 {
		return "", fmt.Errorf("decimal place value must be an integer with range of [0,100], instead, got %d", dp)
	}
	p10 := math.Pow10(dp)
	return fmt.Sprintf("%v", math.Floor(v*p10)/p10), nil
}

func ifElse(_ *transformctx.Ctx, conditionsAndValues ...string) (string, error) {
	if len(conditionsAndValues)%2 != 1 {
		return "", fmt.Errorf("arg number must be odd, but got: %d", len(conditionsAndValues))
	}
	for i := 0; i < len(conditionsAndValues)/2; i++ {
		condition, err := strconv.ParseBool(conditionsAndValues[2*i])
		if err != nil {
			return "", fmt.Errorf(
				`condition argument must be a boolean string, but got: %s`, conditionsAndValues[2*i])
		}
		if condition {
			return conditionsAndValues[(2*i)+1], nil
		}
	}
	return conditionsAndValues[len(conditionsAndValues)-1], nil
}

func isEmpty(_ *transformctx.Ctx, str string) (string, error) {
	if str == "" {
		return "true", nil
	}
	return "false", nil
}

func lower(_ *transformctx.Ctx, s string) (string, error) {
	return strings.ToLower(s), nil
}

// Splits a `s` into substrings separated by `sep` and return an string array represented in json.
// `trim` indicates whether each of the separated substrings will be space-trimmed or not. if `trim`
// is "", it defaults to "false".
// e.g. str = "a,b, c", sep = ",", trim="", result will be `["a", "b", " c"]`.
// e.g. str = "a,b, c", sep = ",", trim="true", result will be `["a", "b", "c"]`.
func splitIntoJsonArray(_ *transformctx.Ctx, s, sep string, trim string) (string, error) {
	if sep == "" {
		return "", fmt.Errorf("'sep' can't be empty")
	}
	if s == "" {
		return "[]", nil
	}
	toTrim := false
	var err error
	if strs.IsStrNonBlank(trim) {
		toTrim, err = strconv.ParseBool(strings.TrimSpace(trim))
		if err != nil {
			return "", fmt.Errorf(
				`'trim' must be either "" (default to "false"") or "true" or "false". err: %s`, err.Error())
		}
	}
	splits := strings.Split(s, sep)
	if toTrim {
		splits = strs.NoErrMapSlice(splits, func(s string) string {
			return strings.TrimSpace(s)
		})
	}
	// strings.Split always returns a valid non-nil slice (could be empty), thus json marshaling
	// will always succeed.
	b, _ := json.Marshal(splits)
	return string(b), nil
}

func substring(_ *transformctx.Ctx, str, startIndex, lengthStr string) (string, error) {
	start, err := strconv.Atoi(startIndex)
	if err != nil {
		return "", fmt.Errorf("unable to convert start index '%s' into int, err: %s", startIndex, err.Error())
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("unable to convert length '%s' into int, err: %s", lengthStr, err.Error())
	}
	if length < -1 {
		return "", fmt.Errorf("length must be >= -1, but got %d", length)
	}
	// We can/do deal with UTF-8 encoded strings. startIndex and length are all about
	// UTF-8 characters not just bytes.
	runes := []rune(str)
	runeLen := len(runes)
	if start < 0 || start > runeLen {
		return "", fmt.Errorf("start index %d is out of bounds (string length is %d)", start, runeLen)
	}
	if length == -1 {
		length = runeLen - start
	}
	if start+length > runeLen {
		return "", fmt.Errorf(
			"start %d + length %d is out of bounds (string length is %d)", start, length, runeLen)
	}
	return string(runes[start : start+length]), nil
}

func switchFunc(ctx *transformctx.Ctx, expr string, casesReturns ...string) (string, error) {
	if len(casesReturns)%2 != 1 {
		return "", fmt.Errorf("length of 'casesReturns' must be odd, but got: %d", len(casesReturns))
	}
	patternsReturns := make([]string, len(casesReturns))
	for i := 0; i < len(patternsReturns)/2; i++ {
		patternsReturns[2*i] = "^" + regexp.QuoteMeta(casesReturns[2*i]) + "$"
		patternsReturns[2*i+1] = casesReturns[2*i+1]
	}
	patternsReturns[len(casesReturns)-1] = casesReturns[len(casesReturns)-1]
	return switchByPattern(ctx, expr, patternsReturns...)
}

func switchByPattern(_ *transformctx.Ctx, expr string, patternsReturns ...string) (string, error) {
	if len(patternsReturns)%2 != 1 {
		return "", fmt.Errorf(
			"length of 'patternsReturns' must be odd, but got: %d", len(patternsReturns))
	}
	for i := 0; i < len(patternsReturns)/2; i++ {
		re, err := caches.GetRegex(patternsReturns[2*i])
		if err != nil {
			return "", fmt.Errorf(`invalid pattern '%s', err: %s`, patternsReturns[2*i], err.Error())
		}
		if re.MatchString(expr) {
			return patternsReturns[(2*i)+1], nil
		}
	}
	return patternsReturns[len(patternsReturns)-1], nil
}

func upper(_ *transformctx.Ctx, s string) (string, error) {
	return strings.ToUpper(s), nil
}

func uuidv3(_ *transformctx.Ctx, s string) (string, error) {
	return uuid.NewMD5(uuid.Nil, []byte(s)).String(), nil
}

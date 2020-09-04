package customfuncs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/strs"
)

// CustomFuncType is the type of a custom function. Has to use interface{} given we support
// non-variadic and variadic functions.
type CustomFuncType = interface{}

// CustomFuncs is a map from custom func names to an actual custom func.
type CustomFuncs = map[string]CustomFuncType

// BuiltinCustomFuncs contains all the built-in custom functions.
var BuiltinCustomFuncs = map[string]CustomFuncType{
	// keep these custom funcs lexically sorted
	"concat":             concat,
	"external":           external,
	"lower":              lower,
	"splitIntoJsonArray": splitIntoJsonArray,
	"substring":          substring,
	"upper":              upper,
}

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

func concat(_ *transformctx.Ctx, strs ...string) (string, error) {
	var b bytes.Buffer
	for _, s := range strs {
		b.WriteString(s)
	}
	return b.String(), nil
}

func external(ctx *transformctx.Ctx, name string) (string, error) {
	if v, found := ctx.ExternalProperty(name); found {
		return v, nil
	}
	return "", fmt.Errorf("cannot find external property '%s'", name)
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

func upper(_ *transformctx.Ctx, s string) (string, error) {
	return strings.ToUpper(s), nil
}

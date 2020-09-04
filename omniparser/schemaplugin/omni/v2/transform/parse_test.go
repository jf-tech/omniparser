package transform

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/nodes"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/strs"
)

func testParseCtx() *parseCtx {
	ctx := newParseCtx(
		&transformctx.Ctx{
			InputName:          "test-input",
			ExternalProperties: map[string]string{"abc": "efg"},
		},
		customfuncs.Merge(
			customfuncs.CustomFuncs{
				"test_func": func(_ *transformctx.Ctx, args ...string) (string, error) {
					return "test", nil
				},
			},
			customfuncs.BuiltinCustomFuncs))
	// by default disabling transform cache in test because vast majority of
	// test cases don't have their decls' hash computed.
	ctx.disableTransformCache = true
	return ctx
}

func TestResultTypeConversion(t *testing.T) {
	for _, test := range []struct {
		name          string
		value         string
		decl          *Decl
		expectedValue interface{}
		expectedErr   string
	}{
		{
			name:          "result_type not specified",
			value:         "test",
			decl:          &Decl{},
			expectedValue: "test",
			expectedErr:   "",
		},
		{
			name:          "string result_type for empty string",
			value:         "",
			decl:          &Decl{ResultType: testResultType(ResultTypeString)},
			expectedValue: "",
			expectedErr:   "",
		},
		{
			name:          "non-string result_type for empty string",
			value:         "",
			decl:          &Decl{ResultType: testResultType(ResultTypeInt)},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name:          "string result_type for non-empty string",
			value:         "test",
			decl:          &Decl{ResultType: testResultType(ResultTypeString)},
			expectedValue: "test",
			expectedErr:   "",
		},
		{
			name:          "int result_type for non-empty string",
			value:         "123",
			decl:          &Decl{ResultType: testResultType(ResultTypeInt)},
			expectedValue: int64(123),
			expectedErr:   "",
		},
		{
			name:          "failed parsing for int result_type",
			value:         "abc",
			decl:          &Decl{ResultType: testResultType(ResultTypeInt)},
			expectedValue: nil,
			expectedErr:   `strconv.ParseFloat: parsing "abc": invalid syntax`,
		},
		{
			name:          "float result_type for non-empty string",
			value:         "123.45",
			decl:          &Decl{ResultType: testResultType(ResultTypeFloat)},
			expectedValue: 123.45,
			expectedErr:   "",
		},
		{
			name:          "failed parsing for float result_type",
			value:         "abc",
			decl:          &Decl{ResultType: testResultType(ResultTypeFloat)},
			expectedValue: nil,
			expectedErr:   `strconv.ParseFloat: parsing "abc": invalid syntax`,
		},
		{
			name:          "boolean result_type for non-empty string",
			value:         "true",
			decl:          &Decl{ResultType: testResultType(ResultTypeBoolean)},
			expectedValue: true,
			expectedErr:   "",
		},
		{
			name:          "failed parsing for boolean result_type",
			value:         "abc",
			decl:          &Decl{ResultType: testResultType(ResultTypeBoolean)},
			expectedValue: nil,
			expectedErr:   `strconv.ParseBool: parsing "abc": invalid syntax`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			typedValue, err := resultTypeConversion(test.decl, test.value)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Regexp(t, test.expectedErr, err.Error())
			}
			assert.Equal(t, test.expectedValue, typedValue)
		})
	}
}

func TestNormalizeAndSaveValue(t *testing.T) {
	for _, test := range []struct {
		name               string
		decl               *Decl
		value              interface{}
		expectedValue      interface{}
		expectedSaveCalled bool
		expectedErr        string
	}{
		{
			name:               "nil value with KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              nil,
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "nil value with KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              nil,
			expectedValue:      nil,
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "non string value saved",
			decl:               &Decl{},
			value:              123.45,
			expectedValue:      123.45,
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is string and KeepLeadingTrailingSpace false",
			decl:               &Decl{},
			value:              " test  ",
			expectedValue:      "test",
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is empty string and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              "",
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "value is string and KeepLeadingTrailingSpace true",
			decl:               &Decl{KeepLeadingTrailingSpace: true},
			value:              " test  ",
			expectedValue:      " test  ",
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name: "value is string but can't convert to result type",
			decl: &Decl{
				ResultType: testResultType(ResultTypeInt),
				fqdn:       "test_fqdn",
			},
			value:              "abc",
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        `fail to convert value 'abc' to type 'int' on 'test_fqdn', err: strconv.ParseFloat: parsing "abc": invalid syntax`,
		},
		{
			name:               "value is empty slice and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              []interface{}{},
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "value is empty slice and KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              []interface{}{},
			expectedValue:      []interface{}{},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is non-empty slice and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              []interface{}{"string1"},
			expectedValue:      []interface{}{"string1"},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is empty map and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              map[string]interface{}{},
			expectedValue:      nil,
			expectedSaveCalled: false,
			expectedErr:        "",
		},
		{
			name:               "value is empty map and KeepEmptyOrNull true",
			decl:               &Decl{KeepEmptyOrNull: true},
			value:              map[string]interface{}{},
			expectedValue:      map[string]interface{}{},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
		{
			name:               "value is non-empty map and KeepEmptyOrNull false",
			decl:               &Decl{},
			value:              map[string]interface{}{"test_key": "test_value"},
			expectedValue:      map[string]interface{}{"test_key": "test_value"},
			expectedSaveCalled: true,
			expectedErr:        "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			saveCalled := false
			err := normalizeAndSaveValue(test.decl, test.value, func(normalizedValue interface{}) {
				saveCalled = true
				assert.Equal(t, test.expectedValue, normalizedValue)
			})
			assert.Equal(t, test.expectedSaveCalled, saveCalled)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Regexp(t, test.expectedErr, err.Error())
			}
		})
	}
}

func TestComputeXPath(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedErr   string
		expectedXPath string
	}{
		{
			name:          "xpath specified",
			decl:          &Decl{XPath: strs.StrPtr("A/B")},
			expectedErr:   "",
			expectedXPath: "A/B",
		},
		{
			name:          "xpath_dynamic - const",
			decl:          &Decl{XPathDynamic: &Decl{Const: strs.StrPtr("A/C"), kind: KindConst}},
			expectedErr:   "",
			expectedXPath: "A/C",
		},
		{
			name:          "xpath_dynamic - invalid xpath",
			decl:          &Decl{XPathDynamic: &Decl{XPath: strs.StrPtr("<"), kind: KindField, fqdn: "fqdn"}},
			expectedErr:   "xpath query '<' on 'fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
			expectedXPath: "",
		},
		{
			name: "xpath_dynamic - no match",
			decl: &Decl{
				XPathDynamic: &Decl{
					XPath: strs.StrPtr("A/non-existing"),
					kind:  KindField,
					fqdn:  "test_fqdn",
				}},
			expectedErr:   "'test_fqdn' failed to yield a single value: no node matched",
			expectedXPath: "",
		},
		{
			name:          "xpath_dynamic - xpath - success",
			decl:          &Decl{XPathDynamic: &Decl{XPath: strs.StrPtr("C"), kind: KindField}},
			expectedErr:   "",
			expectedXPath: "c",
		},
		{
			name: "xpath_dynamic - custom_func - err",
			decl: &Decl{XPathDynamic: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "substring",
					Args: []*Decl{
						{Const: strs.StrPtr(""), kind: KindConst},
						{Const: strs.StrPtr("123"), kind: KindConst}, // will cause an out of bound error
						{Const: strs.StrPtr("321"), kind: KindConst},
					},
					fqdn: "test_fqdn",
				},
				kind: KindCustomFunc,
			}},
			expectedErr:   `'test_fqdn' failed: start index 123 is out of bounds (string length is 0)`,
			expectedXPath: "",
		},
		{
			name: "xpath_dynamic - custom_func - success",
			decl: &Decl{XPathDynamic: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "concat",
					Args: []*Decl{
						{Const: strs.StrPtr("."), kind: KindConst},
						{Const: strs.StrPtr("/"), kind: KindConst},
						{Const: strs.StrPtr("B"), kind: KindConst},
					},
				},
				kind: KindCustomFunc,
			}},
			expectedErr:   "",
			expectedXPath: "./B",
		},
		{
			name:          "xpath / xpath_dynamic both not specified, default to '.'",
			decl:          &Decl{},
			expectedErr:   "",
			expectedXPath: ".",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			xpath, dynamic, err := testParseCtx().computeXPath(testNode(), test.decl)
			switch {
			case strs.IsStrNonBlank(test.expectedErr):
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Equal(t, "", xpath)
			default:
				assert.NoError(t, err)
				assert.Equal(t, test.expectedXPath, xpath)
				assert.Equal(t, test.decl.XPathDynamic != nil, dynamic)
			}
		})
	}
}

func TestXPathMatchFlags(t *testing.T) {
	dynamic := true
	assert.Equal(t, nodes.DisableXPathCache, xpathMatchFlags(dynamic))
	dynamic = false
	assert.Equal(t, uint(0), xpathMatchFlags(dynamic))
}

func TestParseCtx_ParseNode(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue interface{}
		expectedErr   string
	}{
		{
			name:          "unsupported kind",
			decl:          &Decl{kind: "unsupported", fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "unexpected decl kind 'unsupported' on 'test_fqdn'",
		},
		{
			name:          "const kind",
			decl:          &Decl{Const: strs.StrPtr("test_const"), kind: KindConst},
			expectedValue: "test_const",
			expectedErr:   "",
		},
		{
			name:          "External kind",
			decl:          &Decl{External: strs.StrPtr("abc"), kind: KindExternal},
			expectedValue: "efg",
			expectedErr:   "",
		},
		{
			name:          "field kind",
			decl:          &Decl{XPath: strs.StrPtr("B"), kind: KindField},
			expectedValue: "b",
			expectedErr:   "",
		},
		{
			name:          "field xpath query failure",
			decl:          &Decl{XPath: strs.StrPtr("<"), kind: KindField, fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "object kind",
			decl: &Decl{
				children: []*Decl{{XPath: strs.StrPtr("C"), kind: KindField, fqdn: "test_key"}},
				kind:     KindObject,
			},
			expectedValue: map[string]interface{}{
				"test_key": "c",
			},
			expectedErr: "",
		},
		{
			name: "array kind",
			decl: &Decl{
				children: []*Decl{{XPath: strs.StrPtr("B"), kind: KindField}},
				kind:     KindArray,
			},
			expectedValue: []interface{}{"b"},
			expectedErr:   "",
		},
		{
			name: "custom_func kind",
			decl: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "concat",
					Args: []*Decl{
						{Const: strs.StrPtr("abc"), kind: KindConst, hash: "hash-const"},
						{XPath: strs.StrPtr("B"), kind: KindField, hash: "hash-field"},
						{
							CustomFunc: &CustomFuncDecl{
								Name: "lower",
								Args: []*Decl{
									{Const: strs.StrPtr("A"), kind: KindConst, hash: "hash-const2"},
								},
							},
							kind: KindCustomFunc,
						},
						{Const: strs.StrPtr("A"), kind: KindConst, hash: "hash-const2"},
					},
				},
				kind: KindCustomFunc,
			},
			expectedValue: "abcbaA",
			expectedErr:   "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			ctx := testParseCtx()
			ctx.disableTransformCache = false
			value, err := ctx.parseNode(testNode(), test.decl)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Regexp(t, test.expectedErr, err.Error())
			}
			assert.Equal(t, test.expectedValue, value)
		})
	}
}

func TestParseConst(t *testing.T) {
	value, err := testParseCtx().parseConst(&Decl{Const: strs.StrPtr("test_const")})
	assert.NoError(t, err)
	assert.Equal(t, "test_const", value)
}

func TestParseExternal(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue string
		expectedErr   string
	}{
		{
			name:          "externalProperties found",
			decl:          &Decl{External: strs.StrPtr("abc")},
			expectedValue: "efg",
			expectedErr:   "",
		},
		{
			name:          "externalProperties not found",
			decl:          &Decl{External: strs.StrPtr("efg"), fqdn: "test_fqdn"},
			expectedValue: "",
			expectedErr:   "cannot find external property 'efg' on 'test_fqdn'",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			value, err := testParseCtx().parseExternal(test.decl)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			}
			assert.Equal(t, test.expectedValue, value)
		})
	}
}

func resultTypePtr(t ResultType) *ResultType {
	return &t
}
func TestParseCtx_ParseField(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue interface{}
		expectedErr   string
	}{
		{
			name:          "no query needed",
			decl:          &Decl{}, // by leaving both xpath/xpath_dynamic nil, xpathQueryNeeded returns false.
			expectedValue: "bc",
			expectedErr:   "",
		},
		{
			name:          "computeXPath failed so we default value to nil",
			decl:          &Decl{XPathDynamic: &Decl{External: strs.StrPtr("non-existing"), kind: KindExternal}},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name:          "matched",
			decl:          &Decl{XPath: strs.StrPtr("B"), kind: KindField},
			expectedValue: "b",
			expectedErr:   "",
		},
		{
			name:          "no nodes matched",
			decl:          &Decl{XPath: strs.StrPtr("abc"), kind: KindField},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name:          "more than one node matched",
			decl:          &Decl{XPath: strs.StrPtr("*"), kind: KindField, fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "xpath query '*' on 'test_fqdn' yielded more than one result",
		},
		{
			name:          "invalid xpath",
			decl:          &Decl{XPath: strs.StrPtr("<"), kind: KindField, fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name:          "result_type == object",
			decl:          &Decl{ResultType: resultTypePtr(ResultTypeObject)},
			expectedValue: map[string]interface{}{"B": "b", "C": "c"},
			expectedErr:   "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			value, err := testParseCtx().parseField(testNode(), test.decl)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			}
			assert.Equal(t, test.expectedValue, value)
		})
	}
}

func TestParseCtx_ParseCustomFunc(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue interface{}
		expectedErr   string
	}{
		{
			name: "successful invoking",
			decl: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "concat",
					Args: []*Decl{
						{Const: strs.StrPtr("abc"), kind: KindConst},
						{XPath: strs.StrPtr("B"), kind: KindField},
						{
							CustomFunc: &CustomFuncDecl{
								Name: "lower",
								Args: []*Decl{
									{Const: strs.StrPtr("A"), kind: KindConst},
								},
							},
							kind: KindCustomFunc,
						},
					},
				},
				kind: KindCustomFunc,
			},
			expectedValue: "abcba",
			expectedErr:   "",
		},
		{
			name: "failed invoking",
			decl: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "lower",
					Args: []*Decl{
						{External: strs.StrPtr("non-existing"), kind: KindExternal, fqdn: "test_fqdn"},
					},
					IgnoreErrorAndReturnEmptyStr: false,
				},
				kind: KindCustomFunc,
			},
			expectedValue: nil,
			expectedErr:   "cannot find external property 'non-existing' on 'test_fqdn'",
		},
		{
			name: "xpath matches no node",
			decl: &Decl{
				XPath: strs.StrPtr("NO MATCH"),
				CustomFunc: &CustomFuncDecl{
					Name: "lower",
					Args: []*Decl{
						{External: strs.StrPtr("non-existing"), kind: KindExternal},
					},
					IgnoreErrorAndReturnEmptyStr: false,
				},
				kind: KindCustomFunc,
			},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name: "xpath matches more than one node",
			decl: &Decl{
				XPath: strs.StrPtr("*"),
				CustomFunc: &CustomFuncDecl{
					Name: "lower",
					Args: []*Decl{
						{External: strs.StrPtr("non-existing"), kind: KindExternal},
					},
					IgnoreErrorAndReturnEmptyStr: false,
				},
				kind: KindCustomFunc,
				fqdn: "test_fqdn",
			},
			expectedValue: nil,
			expectedErr:   regexp.QuoteMeta(`xpath query '*' on 'test_fqdn' yielded more than one result`),
		},
		{
			name: "resultType is object",
			decl: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "splitIntoJsonArray",
					Args: []*Decl{
						{Const: strs.StrPtr("a/b/c"), kind: KindConst},
						{Const: strs.StrPtr("/"), kind: KindConst},
						{Const: strs.StrPtr("true"), kind: KindConst},
					},
				},
				kind:       KindCustomFunc,
				ResultType: testResultType(ResultTypeObject),
			},
			expectedValue: []interface{}{"a", "b", "c"},
			expectedErr:   "",
		},
		{
			name: "resultType is object but value is nil",
			decl: &Decl{
				XPath: strs.StrPtr("NO MATCH"),
				CustomFunc: &CustomFuncDecl{
					Name: "lower",
					Args: []*Decl{
						{External: strs.StrPtr("non-existing"), kind: KindExternal},
					},
					IgnoreErrorAndReturnEmptyStr: false,
				},
				kind:       KindCustomFunc,
				ResultType: testResultType(ResultTypeObject),
			},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name: "successful invoking and result_type object",
			decl: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "concat",
					Args: []*Decl{
						{Const: strs.StrPtr("abc"), kind: KindConst},
						{XPath: strs.StrPtr("B"), kind: KindField},
						{
							CustomFunc: &CustomFuncDecl{
								Name: "lower",
								Args: []*Decl{
									{Const: strs.StrPtr("A"), kind: KindConst},
								},
							},
							kind: KindCustomFunc,
						},
					},
				},
				kind:       KindCustomFunc,
				ResultType: testResultType(ResultTypeObject),
			},
			expectedValue: nil,
			expectedErr:   "invalid character 'a' looking for beginning of value",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			value, err := testParseCtx().parseCustomFunc(testNode(), test.decl)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Regexp(t, test.expectedErr, err.Error())
			}
			assert.Equal(t, test.expectedValue, value)
		})
	}
}

func TestParseCtx_ParseObject(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue map[string]interface{}
		expectedErr   string
	}{
		{
			name: "final output",
			decl: &Decl{
				fqdn: "FINAL_OUTPUT",
				kind: KindObject,
				children: []*Decl{
					{
						fqdn:  "FINAL_OUTPUT.test_key",
						kind:  KindField,
						XPath: strs.StrPtr("C"),
					},
				},
			},
			expectedValue: map[string]interface{}{
				"test_key": "c",
			},
			expectedErr: "",
		},
		{
			name: "computeXPath failed",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindObject,
				// this would cause computeXPath fail
				XPathDynamic: &Decl{External: strs.StrPtr("non-existing"), kind: KindExternal},
			},
			expectedValue: nil,
			expectedErr:   "", // no error when nothing matched
		},
		{
			name: "no nodes matched for xpath",
			decl: &Decl{
				fqdn:  "test_fqdn",
				kind:  KindObject,
				XPath: strs.StrPtr("abc"), // unmatched xpath
			},
			expectedValue: nil,
			expectedErr:   "", // no error when nothing matched
		},
		{
			name: "invalid xpath",
			decl: &Decl{
				fqdn:  "test_fqdn",
				kind:  KindObject,
				XPath: strs.StrPtr("<"), // invalid xpath
			},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "failed parsing on child node",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindObject,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  KindField,
						XPath: strs.StrPtr("<"), // invalid xpath syntax.
					},
				},
			},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn.test_key' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "failed normalization",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindObject,
				children: []*Decl{
					{
						fqdn:       "test_fqdn.test_key",
						kind:       KindConst,
						Const:      strs.StrPtr("abc"),
						ResultType: testResultType(ResultTypeInt),
					},
				},
			},
			expectedValue: nil,
			expectedErr:   `fail to convert value 'abc' to type 'int' on 'test_fqdn.test_key', err: strconv.ParseFloat: parsing "abc": invalid syntax`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			value, err := testParseCtx().parseObject(testNode(), test.decl)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			}
			if test.expectedValue == nil {
				assert.Nil(t, value)
			} else {
				assert.Equal(t, test.expectedValue, value)
			}
		})
	}
}

func TestParseCtx_ParseArray(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue []interface{}
		expectedErr   string
	}{
		{
			name: "empty array",
			decl: &Decl{
				fqdn:     "test_fqdn",
				kind:     KindArray,
				children: []*Decl{},
			},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name: "computeXPath failed",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindArray,
				children: []*Decl{
					{
						fqdn: "test_fqdn.test_key",
						kind: KindField,
						// this would cause computeXPath fail
						XPathDynamic: &Decl{External: strs.StrPtr("non-existing"), kind: KindExternal},
					},
				},
			},
			expectedValue: nil, // if computeXPath fails, we'll just skip
			expectedErr:   "",
		},
		{
			name: "invalid xpath in child",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindArray,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  KindField,
						XPath: strs.StrPtr("<"), // invalid xpath syntax.
					},
				},
			},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn.test_key' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "no nodes matched for child xpath",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindArray,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  KindField,
						XPath: strs.StrPtr("abc"),
					},
				},
			},
			expectedValue: nil,
			expectedErr:   "", // no error when nothing matched
		},
		{
			name: "failed parsing child",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindArray,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  KindObject,
						XPath: strs.StrPtr("."),
						children: []*Decl{
							{
								fqdn:       "test_fqdn.test_key.test_key2",
								kind:       KindConst,
								Const:      strs.StrPtr("abc"),
								ResultType: testResultType(ResultTypeInt),
							},
						},
					},
				},
			},
			expectedValue: nil,
			expectedErr:   `fail to convert value 'abc' to type 'int' on 'test_fqdn.test_key.test_key2', err: strconv.ParseFloat: parsing "abc": invalid syntax`,
		},
		{
			name: "failed normalization",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: KindArray,
				children: []*Decl{
					{
						fqdn:       "test_fqdn.test_key",
						kind:       KindConst,
						Const:      strs.StrPtr("abc"),
						ResultType: testResultType(ResultTypeInt),
					},
				},
			},
			expectedValue: nil,
			expectedErr:   `fail to convert value 'abc' to type 'int' on 'test_fqdn.test_key', err: strconv.ParseFloat: parsing "abc": invalid syntax`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			value, err := testParseCtx().parseArray(testNode(), test.decl)
			switch test.expectedErr {
			case "":
				assert.NoError(t, err)
			default:
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			}
			if test.expectedValue == nil {
				assert.Nil(t, value)
			} else {
				assert.Equal(t, test.expectedValue, value)
			}
		})
	}
}

package transform

import (
	"errors"
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/customfuncs"
	v21 "github.com/jf-tech/omniparser/extensions/omniv21/customfuncs"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

func testNode() *idr.Node {
	// A
	//    B
	//    C
	nodeA := idr.CreateNode(idr.ElementNode, "A")
	nodeB := idr.CreateNode(idr.ElementNode, "B")
	textB := idr.CreateNode(idr.TextNode, "b")
	nodeC := idr.CreateNode(idr.ElementNode, "C")
	textC := idr.CreateNode(idr.TextNode, "c")
	idr.AddChild(nodeA, nodeB)
	idr.AddChild(nodeB, textB)
	idr.AddChild(nodeA, nodeC)
	idr.AddChild(nodeC, textC)
	return nodeA
}

func testParseCtx() *parseCtx {
	ctx := NewParseCtx(
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
			customfuncs.CommonCustomFuncs,
			v21.OmniV21CustomFuncs),
		CustomParseFuncs{
			"test_custom_parse_str": func(_ *transformctx.Ctx, _ *idr.Node) (interface{}, error) {
				return "abc", nil
			},
			"test_custom_parse_int": func(_ *transformctx.Ctx, _ *idr.Node) (interface{}, error) {
				return 123, nil
			},
			"test_custom_parse_err": func(_ *transformctx.Ctx, _ *idr.Node) (interface{}, error) {
				return nil, errors.New("test_custom_parse_err")
			},
		})
	// by default disabling transform cache in test because vast majority of
	// test cases don't have their decls' hash computed.
	ctx.disableTransformCache = true
	return ctx
}

func testResultType(typ resultType) *resultType {
	return &typ
}

func TestComputeXPath(t *testing.T) {
	for _, test := range []struct {
		name     string
		decl     *Decl
		err      string
		expected string
	}{
		{
			name:     "xpath specified",
			decl:     &Decl{XPath: strs.StrPtr("A/B")},
			err:      "",
			expected: "A/B",
		},
		{
			name:     "xpath_dynamic - const",
			decl:     &Decl{XPathDynamic: &Decl{Const: strs.StrPtr("A/C"), kind: kindConst}},
			err:      "",
			expected: "A/C",
		},
		{
			name:     "xpath_dynamic - ParseNode failure",
			decl:     &Decl{XPathDynamic: &Decl{XPath: strs.StrPtr("<"), kind: kindField, fqdn: "fqdn"}},
			err:      "xpath query '<' on 'fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
			expected: "",
		},
		{
			name: "xpath_dynamic - ParseNode returns nil",
			decl: &Decl{
				XPathDynamic: &Decl{
					XPath: strs.StrPtr("A/non-existing"),
					kind:  kindField,
					fqdn:  "test_fqdn",
				}},
			err:      "xpath_dynamic on 'test_fqdn' yields empty value",
			expected: "",
		},
		{
			name: "xpath_dynamic - ParseNode returns non-string",
			decl: &Decl{XPathDynamic: &Decl{
				Const:      strs.StrPtr("1"),
				ResultType: testResultType(resultTypeInt),
				kind:       kindConst,
				fqdn:       "fqdn"}},
			err:      "xpath_dynamic on 'fqdn' yields a non-string value '1'",
			expected: "",
		},
		{
			name: "xpath_dynamic - ParseNode returns blank string",
			decl: &Decl{XPathDynamic: &Decl{
				Const:  strs.StrPtr("    "),
				NoTrim: true,
				kind:   kindConst,
				fqdn:   "fqdn"}},
			err:      "xpath_dynamic on 'fqdn' yields empty value",
			expected: "",
		},
		{
			name:     "xpath_dynamic - xpath - success",
			decl:     &Decl{XPathDynamic: &Decl{XPath: strs.StrPtr("C"), kind: kindField}},
			err:      "",
			expected: "c",
		},
		{
			name: "xpath_dynamic - custom_func - err",
			decl: &Decl{XPathDynamic: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "dateTimeToRFC3339",
					Args: []*Decl{
						{Const: strs.StrPtr("not a valid datetime string"), kind: kindConst},
						{Const: strs.StrPtr(""), kind: kindConst},
						{Const: strs.StrPtr(""), kind: kindConst},
					},
					fqdn: "test_fqdn",
				},
				kind: kindCustomFunc,
			}},
			err:      `'test_fqdn' failed: unable to parse 'not a valid datetime string' in any supported date/time format`,
			expected: "",
		},
		{
			name: "xpath_dynamic - custom_func - success",
			decl: &Decl{XPathDynamic: &Decl{
				CustomFunc: &CustomFuncDecl{
					Name: "concat",
					Args: []*Decl{
						{Const: strs.StrPtr("."), kind: kindConst},
						{Const: strs.StrPtr("/"), kind: kindConst},
						{Const: strs.StrPtr("B"), kind: kindConst},
					},
				},
				kind: kindCustomFunc,
			}},
			err:      "",
			expected: "./B",
		},
		{
			name:     "xpath / xpath_dynamic both not specified, default to '.'",
			decl:     &Decl{},
			err:      "",
			expected: ".",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			xpath, dynamic, err := testParseCtx().computeXPath(testNode(), test.decl)
			switch {
			case strs.IsStrNonBlank(test.err):
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", xpath)
			default:
				assert.NoError(t, err)
				assert.Equal(t, test.expected, xpath)
				assert.Equal(t, test.decl.XPathDynamic != nil, dynamic)
			}
		})
	}
}

func TestXPathMatchFlags(t *testing.T) {
	dynamic := true
	assert.Equal(t, idr.DisableXPathCache, xpathMatchFlags(dynamic))
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
			decl:          &Decl{Const: strs.StrPtr("test_const"), kind: kindConst},
			expectedValue: "test_const",
			expectedErr:   "",
		},
		{
			name:          "External kind",
			decl:          &Decl{External: strs.StrPtr("abc"), kind: kindExternal},
			expectedValue: "efg",
			expectedErr:   "",
		},
		{
			name:          "field kind",
			decl:          &Decl{XPath: strs.StrPtr("B"), kind: kindField},
			expectedValue: "b",
			expectedErr:   "",
		},
		{
			name:          "field xpath query failure",
			decl:          &Decl{XPath: strs.StrPtr("<"), kind: kindField, fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "object kind",
			decl: &Decl{
				children: []*Decl{{XPath: strs.StrPtr("C"), kind: kindField, fqdn: "test_key"}},
				kind:     kindObject,
			},
			expectedValue: map[string]interface{}{
				"test_key": "c",
			},
			expectedErr: "",
		},
		{
			name: "array kind",
			decl: &Decl{
				children: []*Decl{{XPath: strs.StrPtr("B"), kind: kindField}},
				kind:     kindArray,
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
						{Const: strs.StrPtr("abc"), kind: kindConst, hash: "hash-const"},
						{XPath: strs.StrPtr("B"), kind: kindField, hash: "hash-field"},
						{
							CustomFunc: &CustomFuncDecl{
								Name: "lower",
								Args: []*Decl{
									{Const: strs.StrPtr("A"), kind: kindConst, hash: "hash-const2"},
								},
							},
							kind: kindCustomFunc,
						},
						{Const: strs.StrPtr("A"), kind: kindConst, hash: "hash-const2"},
					},
				},
				kind: kindCustomFunc,
			},
			expectedValue: "abcbaA",
			expectedErr:   "",
		},
		{
			name: "custom_parse kind",
			decl: &Decl{
				CustomParse: strs.StrPtr("test_custom_parse_str"),
				kind:        kindCustomParse,
			},
			expectedValue: "abc",
			expectedErr:   "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			linkParent(test.decl)
			ctx := testParseCtx()
			ctx.disableTransformCache = false
			value, err := ctx.ParseNode(testNode(), test.decl)
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

func TestParseConst(t *testing.T) {
	value, err := testParseCtx().parseConst(&Decl{Const: strs.StrPtr("test_const")})
	assert.NoError(t, err)
	assert.Equal(t, "test_const", value)
}

func TestParseExternal(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue interface{}
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
			expectedValue: nil,
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
			decl:          &Decl{XPathDynamic: &Decl{External: strs.StrPtr("non-existing"), kind: kindExternal}},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name:          "matched",
			decl:          &Decl{XPath: strs.StrPtr("B"), kind: kindField},
			expectedValue: "b",
			expectedErr:   "",
		},
		{
			name:          "no nodes matched",
			decl:          &Decl{XPath: strs.StrPtr("abc"), kind: kindField},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name:          "more than one node matched",
			decl:          &Decl{XPath: strs.StrPtr("*"), kind: kindField, fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "xpath query '*' on 'test_fqdn' yielded more than one result",
		},
		{
			name:          "invalid xpath",
			decl:          &Decl{XPath: strs.StrPtr("<"), kind: kindField, fqdn: "test_fqdn"},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
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
						{Const: strs.StrPtr("abc"), kind: kindConst},
						{XPath: strs.StrPtr("B"), kind: kindField},
						{
							CustomFunc: &CustomFuncDecl{
								Name: "lower",
								Args: []*Decl{
									{Const: strs.StrPtr("A"), kind: kindConst},
								},
							},
							kind: kindCustomFunc,
						},
					},
				},
				kind: kindCustomFunc,
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
						{External: strs.StrPtr("non-existing"), kind: kindExternal, fqdn: "test_fqdn"},
					},
					IgnoreError: false,
				},
				kind: kindCustomFunc,
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
						{External: strs.StrPtr("non-existing"), kind: kindExternal},
					},
					IgnoreError: false,
				},
				kind: kindCustomFunc,
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
						{External: strs.StrPtr("non-existing"), kind: kindExternal},
					},
					IgnoreError: false,
				},
				kind: kindCustomFunc,
				fqdn: "test_fqdn",
			},
			expectedValue: nil,
			expectedErr:   `xpath query '*' on 'test_fqdn' yielded more than one result`,
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
				assert.Equal(t, test.expectedErr, err.Error())
			}
			assert.Equal(t, test.expectedValue, value)
		})
	}
}

func resultTypePtr(typ resultType) *resultType {
	return &typ
}

func TestParseCtx_ParseCustomParse(t *testing.T) {
	for _, test := range []struct {
		name          string
		decl          *Decl
		expectedValue interface{}
		expectedErr   string
	}{
		{
			name: "successful invoking custom_parse that returns a string",
			decl: &Decl{
				CustomParse: strs.StrPtr("test_custom_parse_str"),
				kind:        kindCustomParse,
			},
			expectedValue: "abc",
			expectedErr:   "",
		},
		{
			name: "successful invoking custom_parse that returns an int",
			decl: &Decl{
				CustomParse: strs.StrPtr("test_custom_parse_int"),
				ResultType:  resultTypePtr(resultTypeInt),
				kind:        kindCustomParse,
			},
			expectedValue: 123,
			expectedErr:   "",
		},
		{
			name: "failed invoking custom_parse",
			decl: &Decl{
				CustomParse: strs.StrPtr("test_custom_parse_err"),
				kind:        kindCustomParse,
			},
			expectedValue: nil,
			expectedErr:   "test_custom_parse_err",
		},
		{
			name: "xpath matches no node",
			decl: &Decl{
				XPath:       strs.StrPtr("NO MATCH"),
				CustomParse: strs.StrPtr("test_custom_parse_str"),
				kind:        kindCustomParse,
			},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name: "xpath matches more than one node",
			decl: &Decl{
				XPath:       strs.StrPtr("*"),
				CustomParse: strs.StrPtr("test_custom_parse_str"),
				kind:        kindCustomParse,
				fqdn:        "test_fqdn",
			},
			expectedValue: nil,
			expectedErr:   `xpath query '*' on 'test_fqdn' yielded more than one result`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := testParseCtx().parseCustomParse(testNode(), test.decl)
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
				kind: kindObject,
				children: []*Decl{
					{
						fqdn:  "FINAL_OUTPUT.test_key",
						kind:  kindField,
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
				kind: kindObject,
				// this would cause computeXPath fail
				XPathDynamic: &Decl{External: strs.StrPtr("non-existing"), kind: kindExternal},
			},
			expectedValue: nil,
			expectedErr:   "", // no error when nothing matched
		},
		{
			name: "no nodes matched for xpath",
			decl: &Decl{
				fqdn:  "test_fqdn",
				kind:  kindObject,
				XPath: strs.StrPtr("abc"), // unmatched xpath
			},
			expectedValue: nil,
			expectedErr:   "", // no error when nothing matched
		},
		{
			name: "invalid xpath",
			decl: &Decl{
				fqdn:  "test_fqdn",
				kind:  kindObject,
				XPath: strs.StrPtr("<"), // invalid xpath
			},
			expectedValue: nil,
			expectedErr:   "xpath query '<' on 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "failed parsing on child node",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: kindObject,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  kindField,
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
				kind: kindObject,
				children: []*Decl{
					{
						fqdn:       "test_fqdn.test_key",
						kind:       kindConst,
						Const:      strs.StrPtr("abc"),
						ResultType: testResultType(resultTypeInt),
					},
				},
			},
			expectedValue: nil,
			expectedErr:   `unable to convert value 'abc' to type 'int' on 'test_fqdn.test_key', err: strconv.ParseInt: parsing "abc": invalid syntax`,
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
				kind:     kindArray,
				children: []*Decl{},
			},
			expectedValue: nil,
			expectedErr:   "",
		},
		{
			name: "computeXPath failed",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: kindArray,
				children: []*Decl{
					{
						fqdn: "test_fqdn.test_key",
						kind: kindField,
						// this would cause computeXPath fail
						XPathDynamic: &Decl{External: strs.StrPtr("non-existing"), kind: kindExternal},
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
				kind: kindArray,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  kindField,
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
				kind: kindArray,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  kindField,
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
				kind: kindArray,
				children: []*Decl{
					{
						fqdn:  "test_fqdn.test_key",
						kind:  kindObject,
						XPath: strs.StrPtr("."),
						children: []*Decl{
							{
								fqdn:       "test_fqdn.test_key.test_key2",
								kind:       kindConst,
								Const:      strs.StrPtr("abc"),
								ResultType: testResultType(resultTypeInt),
							},
						},
					},
				},
			},
			expectedValue: nil,
			expectedErr:   `unable to convert value 'abc' to type 'int' on 'test_fqdn.test_key.test_key2', err: strconv.ParseInt: parsing "abc": invalid syntax`,
		},
		{
			name: "failed normalization",
			decl: &Decl{
				fqdn: "test_fqdn",
				kind: kindArray,
				children: []*Decl{
					{
						fqdn:       "test_fqdn.test_key",
						kind:       kindConst,
						Const:      strs.StrPtr("abc"),
						ResultType: testResultType(resultTypeInt),
					},
				},
			},
			expectedValue: nil,
			expectedErr:   `unable to convert value 'abc' to type 'int' on 'test_fqdn.test_key', err: strconv.ParseInt: parsing "abc": invalid syntax`,
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

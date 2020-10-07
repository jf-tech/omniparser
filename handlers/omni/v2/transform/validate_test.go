package transform

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

func TestValidateTransformDeclarations(t *testing.T) {
	for _, test := range []struct {
		name                      string
		transformDeclarationsJson string
		expectedErr               string
	}{
		{
			name: "success",
			transformDeclarationsJson: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "const": "value1" },
                        "field2": { "xpath_dynamic": { "template": "template12" } },
                        "field3": { "xpath": "E/F/G", "object": {
                            "field4": { "array": [
                                { "const": "value4" },
                                { "xpath": "H/I/J" },
                                { "xpath": "K/L/M", "object": {
                                    "field5": { "xpath": "N/O/P" }
                                }},
                                { "template": "template9" }
                            ]}
                        }},
                        "field6": {
                            "custom_func": {
                                "name": "test_func",
                                "args": [
                                    { "xpath": "Q/R/S" },
                                    { "template": "template12" }
                                ]
                            }
                        },
                        "field_9": { "template": "template9" },
                        "field_10": { "xpath_dynamic": { "const": "X/Y/Z" }, "template": "template10" },
                        "field_11": { "template": "template11" },
                        "field_12": { "template": "template12" },
						"field_13": { "custom_parse": "test_custom_parse" }
                    }},
                    "template9": { "xpath": "1/2/3", "object": {
                        "field9": { "xpath": "4/5/6" }
                    }},
                    "template10": { "object": {
                        "field10": { "const": "value10" }
                    }},
                    "template11": { "array": [
                        { "xpath": "T/U/V" }
                    ]},
                    "template12": { "custom_func": {
                        "name": "test_func",
                        "args": [ { "xpath": "W/X" } ]
                    }}
                }
            }`,
			expectedErr: "",
		},
		{
			name: "failure - xpath and xpath_dynamic specified at the same time",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "xpath": "A/B/C", "xpath_dynamic": { "const": "E/F/G" } }
                    }}
                }
            }`,
			expectedErr: "'FINAL_OUTPUT.field1' cannot set both 'xpath' and 'xpath_dynamic' at the same time",
		},
		{
			name: "failure - xpath_dynamic validate fails: custom_func non-existing",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "xpath_dynamic": { "custom_func": {
                            "name": "non-existing",
                            "args": []
                        }}}
                    }}
                }
            }`,
			expectedErr: "unknown custom_func 'non-existing' on 'FINAL_OUTPUT.field1.xpath_dynamic'",
		},
		{
			name: "failure - xpath_dynamic result_type not string",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "xpath_dynamic": { "const": "123", "result_type": "int" } }
                    }}
                }
            }`,
			expectedErr: "expected 'result_type' 'string' for 'FINAL_OUTPUT.field1.xpath_dynamic', but got 'int'",
		},
		{
			name: "failure - xpath_dynamic kind not primitive",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "xpath_dynamic": { "template": "template1" } }
                    }},
                    "template1": { "result_type": "string", "object": {
                        "field2": { "const": "123" },
                        "field3" : { "external": "something" }
                    }}
                }
            }`,
			expectedErr: "expected primitive decl kind for 'FINAL_OUTPUT.field1.xpath_dynamic', but got 'object'",
		},
		{
			name: "failure - object template invalid",
			transformDeclarationsJson: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_12": { "template": "template12" }
                    }}
                }
            }`,
			expectedErr: "'FINAL_OUTPUT.field_12' contains non-existing template reference 'template12'",
		},
		{
			name: "failure - array template invalid",
			transformDeclarationsJson: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field4": { "array": [
                            { "template": "template12" }
                        ]}
                    }}
                }
            }`,
			expectedErr: "'FINAL_OUTPUT.field4.elem\\[1\\]' contains non-existing template reference 'template12'",
		},
		{
			name: "failure - custom_func arg decl validation failure",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": {
                        "name": "test_func",
                        "args": [ { "template": "huh" } ]
                    }}
                }
            }`,
			expectedErr: "'FINAL_OUTPUT.custom_func\\(test_func\\).arg\\[1\\]' contains non-existing template reference 'huh'",
		},
		{
			name: "failure - custom_func arg decl result_type not string or array",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": {
                        "name": "test_func",
                        "args": [ { "const": "true", "result_type": "boolean" } ]
                    }}
                }
            }`,
			expectedErr: "expected 'result_type' 'string' or 'array' for 'FINAL_OUTPUT.custom_func\\(test_func\\).arg\\[1\\]', but got 'boolean'",
		},
		{
			name: "failure - custom_func arg decl kind not primitive",
			transformDeclarationsJson: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": {
                        "name": "test_func",
                        "args": [ { "result_type": "string", "object": {} } ]
                    }}
                }
            }`,
			expectedErr: "expected primitive decl or array kind for 'FINAL_OUTPUT.custom_func\\(test_func\\).arg\\[1\\]', but got 'object'",
		},
		{
			name: "failure - circular template ref",
			transformDeclarationsJson: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_1": { "template": "template1" }
                    }},
                    "template1": { "object": {
                        "field_2": { "template": "template2" }
                    }},
                    "template2": { "object": {
                        "field_3": { "template": "template3" }
                    }},
                    "template3": { "object": {
                        "field_circular": { "template": "template1" }
                    }}
                }
            }`,
			expectedErr: "template circular dependency detected on 'FINAL_OUTPUT.field_1.field_2.field_3.field_circular': 'FINAL_OUTPUT'->'template1'->'template2'->'template3'->'template1'",
		},
		{
			name: "failure - xpath conflict for template reference",
			transformDeclarationsJson: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_1": { "xpath": "abc", "template": "template1" }
                    }},
                    "template1": { "xpath_dynamic": { "const": "efg" }, "object": {
                        "field_2": { "template": "template2" }
                    }}
                }
            }`,
			expectedErr: "cannot specify 'xpath' or 'xpath_dynamic' on both 'FINAL_OUTPUT.field_1' and the template 'template1' it references",
		},
		{
			name: "failure - unknown custom_parse",
			transformDeclarationsJson: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_1": { "xpath": "abc", "custom_parse": "non-existing" }
                    }}
                }
            }`,
			expectedErr: "unknown custom_parse 'non-existing' on 'FINAL_OUTPUT.field_1'",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			finalOutputDecl, err := ValidateTransformDeclarations(
				[]byte(test.transformDeclarationsJson),
				customfuncs.CustomFuncs{
					"test_func": func() {},
				},
				CustomParseFuncs{
					"test_custom_parse": func(_ *transformctx.Ctx, _ *idr.Node) (interface{}, error) {
						return nil, nil
					},
				})
			switch {
			case strs.IsStrNonBlank(test.expectedErr):
				assert.Error(t, err)
				assert.Regexp(t, test.expectedErr, err.Error())
				assert.Nil(t, finalOutputDecl)
			default:
				assert.NoError(t, err)
				cupaloy.SnapshotT(t, jsons.BPM(finalOutputDecl))
			}
		})
	}
}

func TestDetectKind(t *testing.T) {
	for _, test := range []struct {
		name         string
		decl         *Decl
		expectedKind Kind
	}{
		{
			name:         "const",
			decl:         &Decl{Const: strs.StrPtr("test")},
			expectedKind: KindConst,
		},
		{
			name:         "external",
			decl:         &Decl{External: strs.StrPtr("test")},
			expectedKind: KindExternal,
		},
		{
			name:         "custom func",
			decl:         &Decl{CustomFunc: &CustomFuncDecl{Name: "test"}},
			expectedKind: KindCustomFunc,
		},
		{
			name:         "object with empty map",
			decl:         &Decl{XPath: strs.StrPtr("test"), Object: map[string]*Decl{}},
			expectedKind: KindObject,
		},
		{
			name: "object with non-empty map",
			decl: &Decl{
				XPathDynamic: &Decl{},
				Object:       map[string]*Decl{"a": {Const: strs.StrPtr("test")}},
			},
			expectedKind: KindObject,
		},
		{
			name: "array",
			decl: &Decl{
				Array: []*Decl{{Const: strs.StrPtr("test")}},
			},
			expectedKind: KindArray,
		},
		{
			name:         "template",
			decl:         &Decl{XPath: strs.StrPtr("test"), Template: strs.StrPtr("test")},
			expectedKind: KindTemplate,
		},
		{
			name:         "field with xpath",
			decl:         &Decl{XPath: strs.StrPtr("test")},
			expectedKind: KindField,
		},
		{
			name:         "field with xpath_dynamic",
			decl:         &Decl{XPathDynamic: &Decl{}},
			expectedKind: KindField,
		},
		{
			name:         "unknown",
			decl:         &Decl{},
			expectedKind: KindUnknown,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			actualKind := detectKind(test.decl)
			assert.Equal(t, test.expectedKind, actualKind)
		})
	}
}

func TestComputeDeclHash(t *testing.T) {
	decl1 := &Decl{
		Object: map[string]*Decl{
			"field3": {Const: strs.StrPtr("const"), kind: KindConst, fqdn: "root.field3", hash: "h3"},
			"field1": {External: strs.StrPtr("external"), kind: KindExternal, fqdn: "root.field1", hash: "h1"},
			"field2": {Template: strs.StrPtr("template"), kind: KindTemplate, fqdn: "root.field2", hash: "h2"},
		},
		kind: KindObject,
		fqdn: "root",
		hash: "h0",
	}
	decl1.children = []*Decl{decl1.Object["field3"], decl1.Object["field1"], decl1.Object["field2"]}
	decl1.Object["field1"].parent = decl1
	decl1.Object["field2"].parent = decl1
	decl1.Object["field3"].parent = decl1

	assert.Equal(t, "root", decl1.fqdn)

	declHashes := map[string]string{}
	h0 := computeDeclHash(decl1, declHashes)
	assert.Equal(t, 1, len(declHashes))

	decl1Copy := decl1.deepCopy()
	assert.Equal(t, "", decl1Copy.fqdn)

	h0prime := computeDeclHash(decl1Copy, declHashes)
	assert.Equal(t, 1, len(declHashes))
	assert.Equal(t, h0, h0prime)

	assert.NotEqual(t, jsons.BPM(decl1), jsons.BPM(decl1Copy))
}

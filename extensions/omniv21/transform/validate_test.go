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
		name     string
		declJSON string
		err      string
	}{
		{
			name: "success",
			declJSON: ` {
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
						"$field_13 with space. and other non-alphanumeric chars": { "custom_parse": "test_custom_parse" }
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
			err: "",
		},
		{
			name: "failure - xpath and xpath_dynamic specified at the same time",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "xpath": "A/B/C", "xpath_dynamic": { "const": "E/F/G" } }
                    }}
                }
            }`,
			err: "'FINAL_OUTPUT.field1' cannot set both 'xpath' and 'xpath_dynamic' at the same time",
		},
		{
			name: "failure - xpath_dynamic validate fails: custom_func non-existing",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field1": { "xpath_dynamic": { "custom_func": {
                            "name": "non-existing",
                            "args": []
                        }}}
                    }}
                }
            }`,
			err: "unknown custom_func 'non-existing' on 'FINAL_OUTPUT.field1.xpath_dynamic'",
		},
		{
			name: "failure - object template invalid",
			declJSON: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_12": { "template": "template12" }
                    }}
                }
            }`,
			err: "'FINAL_OUTPUT.field_12' contains non-existing template reference 'template12'",
		},
		{
			name: "failure - array template invalid",
			declJSON: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field4": { "array": [
                            { "template": "template12" }
                        ]}
                    }}
                }
            }`,
			err: "'FINAL_OUTPUT.field4.elem[1]' contains non-existing template reference 'template12'",
		},
		{
			name: "failure - custom_func arg decl validation failure",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": {
                        "name": "test_func",
                        "args": [ { "template": "huh" } ]
                    }}
                }
            }`,
			err: "'FINAL_OUTPUT.custom_func(test_func).arg[1]' contains non-existing template reference 'huh'",
		},
		{
			name: "failure - custom_func not a func",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": { "name": "invalid_func_not_a_func" }}
                }
            }`,
			err: "custom_func 'invalid_func_not_a_func' is not a function",
		},
		{
			name: "failure - custom_func missing ctx",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": { "name": "invalid_func_missing_ctx" }}
                }
            }`,
			err: "custom_func 'invalid_func_missing_ctx' missing required ctx argument",
		},
		{
			name: "failure - custom_func missing 2 return values",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": { "name": "invalid_func_missing_return" }}
                }
            }`,
			err: "custom_func 'invalid_func_missing_return' must have 2 return values, instead got 0",
		},
		{
			name: "failure - custom_func missing error return value",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": { "name": "invalid_func_no_err_return" }}
                }
            }`,
			err: "custom_func 'invalid_func_no_err_return' 2nd return value must be of error type, instead got int",
		},
		{
			name: "failure - custom_func arg decl validation failure",
			declJSON: `{
                "transform_declarations": {
                    "FINAL_OUTPUT": { "custom_func": {
                        "name": "test_func",
                        "args": [ { "template": "huh" } ]
                    }}
                }
            }`,
			err: "'FINAL_OUTPUT.custom_func(test_func).arg[1]' contains non-existing template reference 'huh'",
		},
		{
			name: "failure - circular template ref",
			declJSON: ` {
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
			err: "template circular dependency detected on 'FINAL_OUTPUT.field_1.field_2.field_3.field_circular': 'FINAL_OUTPUT'->'template1'->'template2'->'template3'->'template1'",
		},
		{
			name: "failure - xpath conflict for template reference",
			declJSON: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_1": { "xpath": "abc", "template": "template1" }
                    }},
                    "template1": { "xpath_dynamic": { "const": "efg" }, "object": {
                        "field_2": { "template": "template2" }
                    }}
                }
            }`,
			err: "cannot specify 'xpath' or 'xpath_dynamic' on both 'FINAL_OUTPUT.field_1' and the template 'template1' it references",
		},
		{
			name: "failure - unknown custom_parse",
			declJSON: ` {
                "transform_declarations": {
                    "FINAL_OUTPUT": { "object": {
                        "field_1": { "xpath": "abc", "custom_parse": "non-existing" }
                    }}
                }
            }`,
			err: "unknown custom_parse 'non-existing' on 'FINAL_OUTPUT.field_1'",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			finalOutputDecl, err := ValidateTransformDeclarations(
				[]byte(test.declJSON),
				customfuncs.CustomFuncs{
					"test_func":                   func(*transformctx.Ctx) (interface{}, error) { return nil, nil },
					"invalid_func_not_a_func":     "not a func",
					"invalid_func_missing_ctx":    func() {},
					"invalid_func_missing_return": func(*transformctx.Ctx) {},
					"invalid_func_no_err_return":  func(*transformctx.Ctx) (int, int) { return 0, 0 },
				},
				CustomParseFuncs{
					"test_custom_parse": func(_ *transformctx.Ctx, _ *idr.Node) (interface{}, error) {
						return nil, nil
					},
				})
			switch {
			case strs.IsStrNonBlank(test.err):
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, finalOutputDecl)
			default:
				assert.NoError(t, err)
				cupaloy.SnapshotT(t, jsons.BPM(finalOutputDecl))
			}
		})
	}
}

func TestComputeDeclHash(t *testing.T) {
	decl1 := &Decl{
		Object: map[string]*Decl{
			"field3": {Const: strs.StrPtr("const"), kind: kindConst, fqdn: "root.field3", hash: "h3"},
			"field1": {External: strs.StrPtr("external"), kind: kindExternal, fqdn: "root.field1", hash: "h1"},
			"field2": {Template: strs.StrPtr("template"), kind: kindTemplate, fqdn: "root.field2", hash: "h2"},
		},
		kind: kindObject,
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

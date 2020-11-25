package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"
)

func TestMarshalDecl(t *testing.T) {
	var decls []Decl
	err := json.Unmarshal([]byte(`[
            {
                "const": "123",
                "type": "int"
            },
            {
                "external": "abc",
                "no_trim": true,
                "keep_empty_or_null": true,
                "type": "string"
            },
            {
                "xpath": "xyz",
                "no_trim": true,
                "keep_empty_or_null": true
            },
            {
                "xpath_dynamic": { "const": "true", "type": "boolean" },
                "template": "t"
            },
            {
                "xpath": "abc",
                "object": {
                    "field1": { "external": "123" }
                }
            },
            {
                "array": [
                    { "const": "123", "type": "float" }
                ]
            },
            {
                "custom_func": {
                    "name": "upper",
                    "args": [
                        { "const": "  abc  ", "no_trim": true }
                    ],
                    "ignore_error": true
                }
            }
        ]`), &decls)
	assert.NoError(t, err)
	cupaloy.SnapshotT(t, jsons.BPM(decls))
}

func TestMarshalDeclWithParentAndChildren(t *testing.T) {
	var decl Decl
	err := json.Unmarshal([]byte(`{
            "xpath": "abc",
            "object": {
                "field1": { "external": "123" }
            }
        }`), &decl)
	assert.NoError(t, err)
	decl.fqdn = "root"
	decl.kind = kindObject
	decl.children = []*Decl{decl.Object["field1"]}
	childDecl := decl.Object["field1"]
	childDecl.fqdn = "root.field1"
	childDecl.kind = kindExternal
	childDecl.parent = &decl
	cupaloy.SnapshotT(t, jsons.BPM(decl))
}

func TestResolveKind(t *testing.T) {
	for _, test := range []struct {
		name         string
		decl         *Decl
		expectedKind kind
	}{
		{
			name:         "const",
			decl:         &Decl{Const: strs.StrPtr("test")},
			expectedKind: kindConst,
		},
		{
			name:         "external",
			decl:         &Decl{External: strs.StrPtr("test")},
			expectedKind: kindExternal,
		},
		{
			name:         "custom func",
			decl:         &Decl{CustomFunc: &CustomFuncDecl{Name: "test"}},
			expectedKind: kindCustomFunc,
		},
		{
			name:         "object with empty map",
			decl:         &Decl{XPath: strs.StrPtr("test"), Object: map[string]*Decl{}},
			expectedKind: kindObject,
		},
		{
			name: "object with non-empty map",
			decl: &Decl{
				XPathDynamic: &Decl{},
				Object:       map[string]*Decl{"a": {Const: strs.StrPtr("test")}},
			},
			expectedKind: kindObject,
		},
		{
			name: "array",
			decl: &Decl{
				Array: []*Decl{{Const: strs.StrPtr("test")}},
			},
			expectedKind: kindArray,
		},
		{
			name:         "template",
			decl:         &Decl{XPath: strs.StrPtr("test"), Template: strs.StrPtr("test")},
			expectedKind: kindTemplate,
		},
		{
			name:         "field with xpath",
			decl:         &Decl{XPath: strs.StrPtr("test")},
			expectedKind: kindField,
		},
		{
			name:         "field with xpath_dynamic",
			decl:         &Decl{XPathDynamic: &Decl{}},
			expectedKind: kindField,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.decl.resolveKind()
			assert.Equal(t, test.expectedKind, test.decl.kind)
		})
	}
}

func TestIsXPathSet(t *testing.T) {
	assert.True(t, (&Decl{XPath: strs.StrPtr("A/B/C")}).isXPathSet())
	assert.True(t, (&Decl{XPathDynamic: &Decl{}}).isXPathSet())
	assert.False(t, (&Decl{}).isXPathSet())
}

func verifyDeclDeepCopy(t *testing.T, d1, d2 *Decl) {
	if d1 == nil && d2 == nil {
		return
	}
	verifyPtrsInDeepCopy := func(p1, p2 interface{}) {
		// both are nil, that's fine.
		if reflect.ValueOf(p1).IsNil() && reflect.ValueOf(p2).IsNil() {
			return
		}
		// both are not nil, then make sure they point to different memory addresses.
		if !reflect.ValueOf(p1).IsNil() && !reflect.ValueOf(p2).IsNil() {
			assert.True(t, fmt.Sprintf("%p", p1) != fmt.Sprintf("%p", p2))
			return
		}
		// If one is nil the other isn't, something wrong with the deep copy
		assert.FailNow(t, "p1 (%p) != p2 (%p)", p1, p2)
	}
	verifyPtrsInDeepCopy(d1, d2)
	// content is the same
	d1json := jsons.BPM(d1)
	d2json := jsons.BPM(d2)
	assert.Equal(t, d1json, d2json)

	// Just doing verifyPtrsInDeepCopy on d1/d2 isn't enough, because it's possible
	// that d1 and d2 are two different copies of a Decl, but inside, corresponding ptrs
	// from d1/d2 are pointing to some same memory address. We need to this manually for
	// all ptr elements, recursively.
	verifyPtrsInDeepCopy(d1.Const, d2.Const)
	verifyPtrsInDeepCopy(d1.External, d2.External)
	verifyPtrsInDeepCopy(d1.XPath, d2.XPath)

	verifyDeclDeepCopy(t, d1.XPathDynamic, d2.XPathDynamic)

	verifyPtrsInDeepCopy(d1.CustomFunc, d2.CustomFunc)
	if d1.CustomFunc != nil {
		for i := range d1.CustomFunc.Args {
			verifyDeclDeepCopy(t, d1.CustomFunc.Args[i], d2.CustomFunc.Args[i])
		}
	}
	verifyPtrsInDeepCopy(d1.CustomParse, d2.CustomParse)

	verifyPtrsInDeepCopy(d1.Template, d2.Template)

	verifyPtrsInDeepCopy(d1.Object, d2.Object)
	for name := range d1.Object {
		verifyDeclDeepCopy(t, d1.Object[name], d2.Object[name])
	}

	verifyPtrsInDeepCopy(d1.Array, d2.Array)
	for i := range d1.Array {
		verifyDeclDeepCopy(t, d1.Array[i], d2.Array[i])
	}

	verifyPtrsInDeepCopy(d1.ResultType, d2.ResultType)
}

func TestDeclDeepCopy(t *testing.T) {
	declJson := `{ "xpath": "value0", "object": {
        "field1": { "const": "value1", "type": "boolean" },
        "field2": { "external": "value2" },
        "field3": { "xpath": "value3" },
        "field4": { "xpath_dynamic": { "const": "value4" } },
        "field5": { "custom_func": {
            "name": "func5",
            "args": [
                { "const": "arg51" },
                { "external": "arg52" },
                { "xpath": "arg53" },
                { "xpath_dynamic": { "const": "arg54" } },
                { "custom_func": { "name": "arg55", "args": [] } },
                { "template": "arg56" }
            ]
        }},
        "field6": { "template": "value6", "type": "int" },
        "field7": { "xpath_dynamic": { "const": "value7" }, "object": {
            "field71": { "const": "value71" },
            "field72": { "keep_empty_or_null": true, "object": {
                "field721": { "const": "value721", "type": "float" }
            }}
        }},
        "field8": { "array": [
            { "const": "field81", "type": "string", "no_trim": true },
            { "template": "field82" },
            { "object": {
                "field831": { "const": "value831" }
            }}
        ]},
		"field9": { "xpath": "value9", "custom_parse": "cp9" }
    }}`
	var src Decl
	assert.NoError(t, json.Unmarshal([]byte(declJson), &src))
	dst := src.deepCopy()
	verifyDeclDeepCopy(t, &src, dst)
}

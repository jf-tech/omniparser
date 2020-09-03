package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/testlib"
)

func TestMarshalDecl(t *testing.T) {
	var decls []Decl
	err := json.Unmarshal([]byte(`[
            {
                "const": "123",
                "result_type": "int"
            },
            {
                "external": "abc",
                "keep_leading_trailing_space": true,
                "keep_empty_or_null": true,
                "result_type": "string"
            },
            {
                "xpath": "xyz",
                "keep_leading_trailing_space": true,
                "keep_empty_or_null": true
            },
            {
                "xpath_dynamic": { "const": "true", "result_type": "boolean" },
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
                    { "const": "123", "result_type": "float" }
                ]
            },
            {
                "custom_func": {
                    "name": "upper",
                    "args": [
                        { "const": "  abc  ", "keep_leading_trailing_space": true }
                    ],
                    "IgnoreErrorAndReturnEmptyStr": true
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
	decl.kind = KindObject
	decl.children = []*Decl{decl.Object["field1"]}
	childDecl := decl.Object["field1"]
	childDecl.fqdn = "root.field1"
	childDecl.kind = KindExternal
	childDecl.parent = &decl
	cupaloy.SnapshotT(t, jsons.BPM(decl))
}

func testResultType(rt ResultType) *ResultType {
	return &rt
}

func TestDeclResultType(t *testing.T) {
	assert.Equal(t, ResultTypeUnknown, (&Decl{}).resultType())
	assert.Equal(t, ResultTypeUnknown, (&Decl{kind: KindTemplate}).resultType())
	assert.Equal(t, ResultTypeString, (&Decl{kind: KindConst}).resultType())
	assert.Equal(t, ResultTypeString, (&Decl{kind: KindExternal}).resultType())
	assert.Equal(t, ResultTypeString, (&Decl{kind: KindField}).resultType())
	assert.Equal(t, ResultTypeString, (&Decl{kind: KindCustomFunc}).resultType())
	assert.Equal(t, ResultTypeObject, (&Decl{kind: KindObject}).resultType())
	assert.Equal(t, ResultTypeArray, (&Decl{kind: KindArray}).resultType())
	assert.Equal(t, ResultTypeBoolean, (&Decl{ResultType: testResultType(ResultTypeBoolean)}).resultType())
	assert.Equal(t, ResultTypeString, (&Decl{ResultType: testResultType(ResultTypeString)}).resultType())
	assert.Equal(t, ResultTypeInt, (&Decl{ResultType: testResultType(ResultTypeInt)}).resultType())
	assert.Equal(t, ResultTypeFloat, (&Decl{ResultType: testResultType(ResultTypeFloat)}).resultType())
}

func TestIsPrimitiveKind(t *testing.T) {
	assert.True(t, (&Decl{kind: KindConst}).isPrimitiveKind())
	assert.True(t, (&Decl{kind: KindExternal}).isPrimitiveKind())
	assert.True(t, (&Decl{kind: KindField}).isPrimitiveKind())
	assert.True(t, (&Decl{kind: KindCustomFunc}).isPrimitiveKind())

	assert.False(t, (&Decl{kind: KindObject}).isPrimitiveKind())
	assert.False(t, (&Decl{kind: KindArray}).isPrimitiveKind())
	assert.False(t, (&Decl{kind: KindTemplate}).isPrimitiveKind())
}

func TestIsXPathSet(t *testing.T) {
	assert.True(t, (&Decl{XPath: testlib.StrPtr("A/B/C")}).isXPathSet())
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

	verifyPtrsInDeepCopy(d1.Template, d2.Template)

	verifyPtrsInDeepCopy(d1.Object, d2.Object)
	for name, _ := range d1.Object {
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
        "field1": { "const": "value1", "result_type": "boolean" },
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
        "field6": { "template": "value6", "result_type": "int" },
        "field7": { "xpath_dynamic": { "const": "value7" }, "object": {
            "field71": { "const": "value71" },
            "field72": { "keep_empty_or_null": true, "object": {
                "field721": { "const": "value721", "result_type": "float" }
            }}
        }},
        "field8": { "array": [
            { "const": "field81", "result_type": "string", "keep_leading_trailing_space": true },
            { "template": "field82" },
            { "object": {
                "field831": { "const": "value831" }
            }}
        ]}
    }}`
	var src Decl
	assert.NoError(t, json.Unmarshal([]byte(declJson), &src))
	dst := src.deepCopy()
	verifyDeclDeepCopy(t, &src, dst)
}

package transform

import (
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"
)

func testNode() *node.Node {
	// A
	//    B
	//    C
	nodeA := &node.Node{Type: node.ElementNode, Data: "A"}
	nodeB := &node.Node{Type: node.ElementNode, Data: "B"}
	textB := &node.Node{Type: node.TextNode, Data: "b"}
	nodeC := &node.Node{Type: node.ElementNode, Data: "C"}
	textC := &node.Node{Type: node.TextNode, Data: "c"}
	node.AddChild(nodeA, nodeB)
	node.AddChild(nodeB, textB)
	node.AddChild(nodeA, nodeC)
	node.AddChild(nodeC, textC)
	return nodeA
}

func TestInvokeCustomFunc_Success(t *testing.T) {
	result, err := testParseCtx().invokeCustomFunc(
		testNode(),
		&CustomFuncDecl{
			Name: "concat",
			Args: []*Decl{
				{Const: strs.StrPtr("["), kind: KindConst},
				// multiple values returned and only the first one is used.
				// note the only time when multiple values are used is when an xpath arg is the only
				// arg to a variadic custom func.
				{XPath: strs.StrPtr("*"), kind: KindField},
				{Const: strs.StrPtr("'"), kind: KindConst},
				{
					CustomFunc: &CustomFuncDecl{
						Name: "upper",
						Args: []*Decl{
							// this xpath going up too far, xquery failed and empty string is return/used.
							{XPath: strs.StrPtr("../../Huh"), kind: KindField},
						},
					},
					kind: KindCustomFunc,
				},
				{
					CustomFunc: &CustomFuncDecl{
						Name: "external",
						Args: []*Decl{
							// this would cause 'external' custom func to fail, but
							// IgnoreErrorAndReturnEmptyStr would come in for rescue.
							{Const: strs.StrPtr("non-existing"), kind: KindConst},
						},
						IgnoreErrorAndReturnEmptyStr: true,
						fqdn:                         "test_fqdn",
					},
					kind: KindCustomFunc,
				},
				{Const: strs.StrPtr("'"), kind: KindConst},
				{
					CustomFunc: &CustomFuncDecl{
						Name: "concat",
						Args: []*Decl{
							{XPath: strs.StrPtr("*"), kind: KindField}, // multiple values returned and used.
						},
					},
					kind: KindCustomFunc,
				},
				{Const: strs.StrPtr("-"), kind: KindConst},
				{External: strs.StrPtr("abc"), kind: KindExternal},
				{Const: strs.StrPtr("-"), kind: KindConst},
				{
					CustomFunc: &CustomFuncDecl{
						Name: "javascript_with_context",
						Args: []*Decl{
							{Const: strs.StrPtr("var n=JSON.parse(_node); '['+n.B+'/'+n.C+']'"), kind: KindConst},
						},
					},
					kind: KindCustomFunc,
				},
			},
		})
	assert.NoError(t, err)
	assert.Equal(t, "[b''bc-efg-[b/c]", result)
}

func TestInvokeCustomFunc_MultipleValuesFromXPathUsed(t *testing.T) {
	result, err := testParseCtx().invokeCustomFunc(
		testNode(),
		&CustomFuncDecl{
			Name: "concat",
			Args: []*Decl{
				// multiple values returned and all values are used.
				// note the only time when multiple values are used is when an xpath arg is the only
				// arg to a variadic custom func.
				{XPath: strs.StrPtr("*"), kind: KindField},
			},
		})
	assert.NoError(t, err)
	assert.Equal(t, "bc", result)
}

func TestInvokeCustomFunc_ArrayArgSuccess(t *testing.T) {
	decl := &CustomFuncDecl{
		Name: "concat",
		Args: []*Decl{
			// This array arg will return 'B' and 'C'
			{kind: KindArray, Array: []*Decl{
				{kind: KindCustomFunc, XPath: strs.StrPtr("*"), CustomFunc: &CustomFuncDecl{
					Name: "upper",
					Args: []*Decl{
						{XPath: strs.StrPtr("."), kind: KindField},
					},
				}},
			}},
			// This array arg will nothing
			{kind: KindArray, Array: []*Decl{
				{kind: KindCustomFunc, XPath: strs.StrPtr("non-existing"), CustomFunc: &CustomFuncDecl{
					Name: "upper",
					Args: []*Decl{
						{XPath: strs.StrPtr("."), kind: KindField},
					},
				}},
			}},
		},
	}
	decl.Args[0].children = append(decl.Args[0].children, decl.Args[0].Array[0])
	decl.Args[0].children[0].parent = decl.Args[0]
	result, err := testParseCtx().invokeCustomFunc(testNode(), decl)
	assert.NoError(t, err)
	assert.Equal(t, "BC", result)
}

func TestInvokeCustomFunc_ArrayArgFailure_InvalidXPath(t *testing.T) {
	decl := &CustomFuncDecl{
		Name: "concat",
		Args: []*Decl{
			{kind: KindArray, Array: []*Decl{
				{kind: KindCustomFunc, XPath: strs.StrPtr("<"), CustomFunc: &CustomFuncDecl{
					Name: "upper",
					Args: []*Decl{
						{XPath: strs.StrPtr("."), kind: KindField},
					},
				}},
			}},
		},
	}
	decl.Args[0].children = append(decl.Args[0].children, decl.Args[0].Array[0])
	decl.Args[0].children[0].parent = decl.Args[0]
	_, err := testParseCtx().invokeCustomFunc(testNode(), decl)
	assert.Error(t, err)
	assert.Equal(t,
		"xpath query '<' on '' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		err.Error())
}

func TestParseCtx_InvokeCustomFuncFailure(t *testing.T) {
	for _, test := range []struct {
		name           string
		customFuncDecl *CustomFuncDecl
		expectedErr    string
	}{
		{
			name: "external property not found",
			customFuncDecl: &CustomFuncDecl{
				Name: "upper",
				Args: []*Decl{
					{External: strs.StrPtr("efg"), kind: KindExternal, fqdn: "test_fqdn"},
				},
			},
			expectedErr: "cannot find external property 'efg' on 'test_fqdn'",
		},
		{
			name: "failed custom func call",
			customFuncDecl: &CustomFuncDecl{
				Name: "external",
				Args: []*Decl{
					{Const: strs.StrPtr("non-existing"), kind: KindConst},
				},
				fqdn: "test_fqdn",
			},
			expectedErr: "'test_fqdn' failed: cannot find external property 'non-existing'",
		},
		{
			name: "compute xpath failure",
			customFuncDecl: &CustomFuncDecl{
				Name: "concat",
				Args: []*Decl{
					{
						XPathDynamic: &Decl{
							External: strs.StrPtr("non-existing"),
							kind:     KindExternal,
							fqdn:     "test_fqdn",
						},
						kind: KindField,
					},
				},
			},
			expectedErr: "cannot find external property 'non-existing' on 'test_fqdn'",
		},
		{
			name: "failed to match node",
			customFuncDecl: &CustomFuncDecl{
				Name: "concat",
				Args: []*Decl{
					{Const: strs.StrPtr("abc"), kind: KindConst},
					// xpath is syntactically invalid.
					{XPath: strs.StrPtr("<"), kind: KindField, fqdn: "test_fqdn"},
					{Const: strs.StrPtr("abc"), kind: KindConst},
				},
			},
			expectedErr: "xpath query '<' for 'test_fqdn' failed: xpath '<' compilation failed: expression must evaluate to a node-set",
		},
		{
			name: "nested custom func failure",
			customFuncDecl: &CustomFuncDecl{
				Name: "concat",
				Args: []*Decl{
					{Const: strs.StrPtr("abc"), kind: KindConst},
					{Const: strs.StrPtr("efg"), kind: KindConst},
					{
						CustomFunc: &CustomFuncDecl{
							Name: "external",
							Args: []*Decl{
								{Const: strs.StrPtr("non-existing"), kind: KindConst}, // Invalid
							},
							fqdn: "test_fqdn",
						},
						kind: KindCustomFunc,
					},
				},
			},
			expectedErr: "'test_fqdn' failed: cannot find external property 'non-existing'",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := testParseCtx().invokeCustomFunc(testNode(), test.customFuncDecl)
			assert.Error(t, err)
			assert.Regexp(t, test.expectedErr, err.Error())
			assert.Equal(t, "", result)
		})
	}
}

package transform

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

func TestInvokeCustomFunc(t *testing.T) {
	for _, test := range []struct {
		name     string
		n        *idr.Node
		decl     *CustomFuncDecl
		err      string
		expected interface{}
	}{
		{
			name: "arg parsing error",
			n:    testNode(),
			decl: &CustomFuncDecl{
				Name: "concat",
				Args: []*Decl{
					{
						External: strs.StrPtr("non-existing"),
						kind:     kindExternal,
						fqdn:     "test-arg-fqdn",
					},
				},
				fqdn: "test-fqdn",
			},
			err:      `cannot find external property 'non-existing' on 'test-arg-fqdn'`,
			expected: nil,
		},
		{
			name: "custom_func call failure",
			n:    testNode(),
			decl: &CustomFuncDecl{
				Name: "javascript",
				Args: []*Decl{
					{
						Const: strs.StrPtr("var;"),
						kind:  kindConst,
						fqdn:  "test-arg-fqdn",
					},
				},
				fqdn: "test-fqdn",
			},
			err:      `'test-fqdn' failed: invalid javascript: SyntaxError: (anonymous): Line 1:4 Unexpected token ; (and 1 more errors)`,
			expected: nil,
		},
		{
			name: "custom_func call failure but ignored",
			n:    testNode(),
			decl: &CustomFuncDecl{
				Name: "concat",
				Args: []*Decl{
					{
						Const: strs.StrPtr("a/"),
						kind:  kindConst,
						fqdn:  "test-arg1-fqdn",
					},
					{
						CustomFunc: &CustomFuncDecl{
							Name: "javascript_with_context",
							Args: []*Decl{
								{
									Const: strs.StrPtr("var;"),
									kind:  kindConst,
									fqdn:  "test-arg2-js-arg-fqdn",
								},
								{
									Const: strs.StrPtr("not used arg 1"),
									kind:  kindConst,
								},
								{
									Const: strs.StrPtr("not used arg 1"),
									kind:  kindConst,
								},
							},
							IgnoreError: true,
							fqdn:        "test-arg2-js-fqdn",
						},
						ResultType: testResultType(resultTypeString),
						kind:       kindCustomFunc,
						fqdn:       "test-arg2-fqdn",
					},
					{
						Const: strs.StrPtr("/b"),
						kind:  kindConst,
						fqdn:  "test-arg3-fqdn",
					},
				},
				fqdn: "a//b",
			},
			err:      ``,
			expected: "a//b",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, err := testParseCtx().invokeCustomFunc(test.n, test.decl)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, r)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, r)
			}
		})
	}
}

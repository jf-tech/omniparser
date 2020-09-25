package transform

import (
	"errors"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/transformctx"
)

func TestInvokeCustomParse(t *testing.T) {
	for _, test := range []struct {
		name        string
		customParse CustomParseFuncType
		n           *node.Node
		expected    interface{}
		expectedErr string
	}{
		{
			name: "success",
			customParse: func(ctx *transformctx.Ctx, n *node.Node) (interface{}, error) {
				assert.Equal(t, "test-input", ctx.InputName)
				assert.Equal(t, "123", n.InnerText())
				return "321", nil
			},
			n:           &node.Node{Type: node.TextNode, Data: "123"},
			expected:    "321",
			expectedErr: "",
		},
		{
			name: "failure",
			customParse: func(ctx *transformctx.Ctx, n *node.Node) (interface{}, error) {
				assert.Equal(t, "test-input", ctx.InputName)
				assert.Equal(t, "123", n.InnerText())
				return nil, errors.New("test failure")
			},
			n:           &node.Node{Type: node.TextNode, Data: "123"},
			expected:    nil,
			expectedErr: "test failure",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := testParseCtx().invokeCustomParse(test.customParse, test.n)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

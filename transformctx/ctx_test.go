package transformctx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtx_External(t *testing.T) {
	for _, test := range []struct {
		name          string
		props         map[string]string
		lookup        string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "nil",
			props:         nil,
			lookup:        "xyz",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "empty",
			props:         map[string]string{},
			lookup:        "xyz",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "not found",
			props:         map[string]string{"123": "123"},
			lookup:        "xyz",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "found",
			props:         map[string]string{"xyz": "abc"},
			lookup:        "xyz",
			expectedValue: "abc",
			expectedFound: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			v, found := (&Ctx{ExternalProperties: test.props}).External(test.lookup)
			assert.Equal(t, test.expectedValue, v)
			assert.Equal(t, test.expectedFound, found)
		})
	}
}

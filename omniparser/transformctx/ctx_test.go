package transformctx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtx_ExternalProperty(t *testing.T) {
	for _, test := range []struct {
		name               string
		externalProperties map[string]string
		propNameToLookUp   string
		expectedValue      string
		expectedFound      bool
	}{
		{
			name:               "externalProperties nil",
			externalProperties: nil,
			propNameToLookUp:   "abc",
			expectedValue:      "",
			expectedFound:      false,
		},
		{
			name:               "externalProperties empty",
			externalProperties: map[string]string{},
			propNameToLookUp:   "efg",
			expectedValue:      "",
			expectedFound:      false,
		},
		{
			name:               "can't find prop",
			externalProperties: map[string]string{"abc": "abc"},
			propNameToLookUp:   "efg",
			expectedValue:      "",
			expectedFound:      false,
		},
		{
			name:               "found",
			externalProperties: map[string]string{"abc": "123"},
			propNameToLookUp:   "abc",
			expectedValue:      "123",
			expectedFound:      true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := &Ctx{ExternalProperties: test.externalProperties}
			v, found := ctx.ExternalProperty(test.propNameToLookUp)
			assert.Equal(t, test.expectedValue, v)
			assert.Equal(t, test.expectedFound, found)
		})
	}
}

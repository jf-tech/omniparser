package omniv2

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/errs"
)

func TestParseSchema(t *testing.T) {
	p, err := ParseSchema(nil)
	assert.Equal(t, errs.ErrSchemaNotSupported, err)
	assert.Nil(t, p)
}

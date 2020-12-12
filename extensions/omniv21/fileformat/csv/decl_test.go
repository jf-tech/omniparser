package csv

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"
)

func TestColumnName(t *testing.T) {
	assert.Equal(t, "name", Column{Name: "name"}.name())
	assert.Equal(t, "alias", Column{Name: "name", Alias: strs.StrPtr("alias")}.name())
}

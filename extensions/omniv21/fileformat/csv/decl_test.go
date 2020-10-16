package csv

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"
)

func TestColumnName(t *testing.T) {
	assert.Equal(t, "name", column{Name: "name"}.name())
	assert.Equal(t, "alias", column{Name: "name", Alias: strs.StrPtr("alias")}.name())
}

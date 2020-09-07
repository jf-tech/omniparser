package testlib

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTempFileWithContent_Success(t *testing.T) {
	tmpEmpty := CreateTempFileWithContent(t, "", t.Name(), "")
	defer func() { assert.NoError(t, os.Remove(tmpEmpty.Name())) }()
	actual, err := ioutil.ReadFile(tmpEmpty.Name())
	assert.NoError(t, err)
	assert.Equal(t, "", string(actual))

	tmpSuccess := CreateTempFileWithContent(t, "", t.Name(), "success")
	defer func() { assert.NoError(t, os.Remove(tmpSuccess.Name())) }()
	actual, err = ioutil.ReadFile(tmpSuccess.Name())
	assert.NoError(t, err)
	assert.Equal(t, "success", string(actual))
}

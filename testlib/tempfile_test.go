package testlib

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTempFileWithContent_BadDir(t *testing.T) {
	_, err := CreateTempFileWithContent("some_dir_does_not_exist", t.Name(), "success")
	assert.Error(t, err)
}

func TestCreateTempFileWithContent_Success(t *testing.T) {
	tmp, err := CreateTempFileWithContent("", t.Name(), "success")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(tmp.Name())) }()
	actual, err := ioutil.ReadFile(tmp.Name())
	assert.NoError(t, err)
	assert.Equal(t, "success", string(actual))
}

func TestCreateTempFileWithContent_EmptyContent(t *testing.T) {
	tmp, err := CreateTempFileWithContent("", t.Name(), "")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(tmp.Name())) }()
	actual, err := ioutil.ReadFile(tmp.Name())
	assert.NoError(t, err)
	assert.Equal(t, "", string(actual))
}

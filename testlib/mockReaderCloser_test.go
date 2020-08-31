package testlib

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMockReadCloser(t *testing.T) {
	reader := NewMockReadCloser("test failure", nil)
	_, err := ioutil.ReadAll(reader)
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())
	err = reader.Close()
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())

	content := "this is a test"
	reader = NewMockReadCloser("", []byte(content))
	actualContent, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, content, string(actualContent))
	err = reader.Close()
	assert.NoError(t, err)
}

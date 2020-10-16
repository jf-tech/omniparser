package json

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/idr"
)

func TestIsErrNodeReadingFailed(t *testing.T) {
	assert.True(t, IsErrNodeReadingFailed(ErrNodeReadingFailed("test")))
	assert.Equal(t, "test", ErrNodeReadingFailed("test").Error())
	assert.False(t, IsErrNodeReadingFailed(errors.New("test")))
}

func TestReader_Read_Success(t *testing.T) {
	r, err := NewReader(
		"test-input",
		strings.NewReader(`
			[
				{
					"id": 123,
					"name": "john",
					"age": 40
				},
				{
					"id": 456,
					"name": "jane",
					"age": 50
				},
				{
					"id": 789,
					"name": "smith",
					"age": 20
				}
			]`),
		"/*[age>30]")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.NoError(t, err)
	name, err := idr.MatchSingle(n, "name")
	assert.NoError(t, err)
	assert.Equal(t, "john", name.InnerText())
	// intentionally not calling r.Release(n) to verify that the
	// stream node is freed up by a subsequent Read() call.

	n, err = r.Read()
	assert.NoError(t, err)
	name, err = idr.MatchSingle(n, "name")
	assert.NoError(t, err)
	assert.Equal(t, "jane", name.InnerText())
	r.Release(n)

	n, err = r.Read()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestReader_Read_InvalidJSON(t *testing.T) {
	r, err := NewReader("test-input", strings.NewReader("{\n}\n}"), "/A/B[. != 'c']")
	assert.NoError(t, err)

	n, err := r.Read()
	assert.Error(t, err)
	assert.True(t, IsErrNodeReadingFailed(err))
	assert.Equal(t,
		`input 'test-input' before/near line 3: invalid character '}' looking for beginning of value`,
		err.Error())
	assert.Nil(t, n)
}

func TestReader_FmtErr(t *testing.T) {
	r, err := NewReader("test-input", strings.NewReader(""), "/A/B")
	assert.NoError(t, err)
	err = r.FmtErr("golang is %s", "fun")
	assert.Error(t, err)
	assert.Equal(t, `input 'test-input' before/near line 1: golang is fun`, err.Error())
}

func TestReader_IsContinuableError(t *testing.T) {
	r, err := NewReader("test", strings.NewReader(""), "/A/B")
	assert.NoError(t, err)
	assert.False(t, r.IsContinuableError(io.EOF))
	assert.False(t, r.IsContinuableError(ErrNodeReadingFailed("failure")))
	assert.True(t, r.IsContinuableError(errs.ErrTransformFailed("failure")))
	assert.True(t, r.IsContinuableError(errors.New("failure")))
}

func TestNewReader_InvalidXPath(t *testing.T) {
	r, err := NewReader("test-input", strings.NewReader(""), "[not-valid")
	assert.Error(t, err)
	assert.Equal(t,
		`invalid xpath '[not-valid', err: expression must evaluate to a node-set`,
		err.Error())
	assert.Nil(t, r)
}

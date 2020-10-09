package omniparser

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/errs"
)

type testReadCall struct {
	record []byte
	err    error
}

type testIngester struct {
	readCalled      int
	readCalls       []testReadCall
	continuableErrs map[error]bool
}

func (g *testIngester) Read() ([]byte, error) {
	if g.readCalled >= len(g.readCalls) {
		panic(fmt.Sprintf("Read() called %d time(s), but not enough mock entries setup", g.readCalled))
	}
	r := g.readCalls[g.readCalled]
	g.readCalled++
	return r.record, r.err
}

func (g *testIngester) IsContinuableError(err error) bool {
	_, found := g.continuableErrs[err]
	return found
}

func (g *testIngester) FmtErr(format string, args ...interface{}) error {
	return errors.New("ctx formatted: " + fmt.Sprintf(format, args...))
}

func TestTransform_EndWithEOF(t *testing.T) {
	continuableErr1 := errors.New("continuable error 1")
	tfm := &transform{
		ingester: &testIngester{
			readCalls: []testReadCall{
				{record: []byte("1st good read")},
				{err: continuableErr1},
				{record: []byte("2nd good read")},
				{err: io.EOF},
			},
			continuableErrs: map[error]bool{continuableErr1: true},
		},
	}
	assert.True(t, tfm.Next())
	record, err := tfm.Read()
	assert.NoError(t, err)
	assert.Equal(t, "1st good read", string(record))

	assert.True(t, tfm.Next())
	record, err = tfm.Read()
	assert.Error(t, err)
	assert.True(t, errs.IsErrTransformFailed(err))
	assert.Equal(t, continuableErr1.Error(), err.Error())
	assert.Nil(t, record)

	assert.True(t, tfm.Next())
	record, err = tfm.Read()
	assert.NoError(t, err)
	assert.Equal(t, "2nd good read", string(record))

	assert.False(t, tfm.Next())
	record, err = tfm.Read()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, record)

	// Verifying when EOF is reached, repeatedly calling Next will still get you EOF.
	assert.False(t, tfm.Next())
	record, err = tfm.Read()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, record)
}

func TestTransform_EndWithNonContinuableError(t *testing.T) {
	tfm := &transform{
		ingester: &testIngester{
			readCalls: []testReadCall{
				{record: []byte("1st good read")},
				{err: errors.New("fatal error")},
			},
		},
	}
	assert.True(t, tfm.Next())
	record, err := tfm.Read()
	assert.NoError(t, err)
	assert.Equal(t, "1st good read", string(record))

	assert.True(t, tfm.Next())
	record, err = tfm.Read()
	assert.Error(t, err)
	assert.False(t, errs.IsErrTransformFailed(err))
	assert.Equal(t, "fatal error", err.Error())
	assert.Nil(t, record)

	// Verifying when fatal error occurred, repeatedly calling Next/Read will still get you the same err
	assert.False(t, tfm.Next())
	record, err = tfm.Read()
	assert.Error(t, err)
	assert.Equal(t, "fatal error", err.Error())
	assert.Nil(t, record)
}

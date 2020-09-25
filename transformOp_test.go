package omniparser

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/errs"
)

type testReadCall struct {
	record []byte
	err    error
}

type testInputProcessor struct {
	readCalled      int
	readCalls       []testReadCall
	continuableErrs map[error]bool
}

func (p *testInputProcessor) Read() ([]byte, error) {
	if p.readCalled >= len(p.readCalls) {
		panic(fmt.Sprintf("Read() called %d time(s), but not enough mock entries setup", p.readCalled))
	}
	r := p.readCalls[p.readCalled]
	p.readCalled++
	return r.record, r.err
}

func (p *testInputProcessor) IsContinuableError(err error) bool {
	_, found := p.continuableErrs[err]
	return found
}

func (p *testInputProcessor) FmtErr(format string, args ...interface{}) error {
	return errors.New("ctx formatted: " + fmt.Sprintf(format, args...))
}

func TestTransformOp_EndWithEOF(t *testing.T) {
	continuableErr1 := errors.New("continuable error 1")
	op := &transformOp{
		inputProcessor: &testInputProcessor{
			readCalls: []testReadCall{
				{record: []byte("1st good read")},
				{err: continuableErr1},
				{record: []byte("2nd good read")},
				{err: errs.ErrEOF},
			},
			continuableErrs: map[error]bool{continuableErr1: true},
		},
	}
	assert.True(t, op.Next())
	record, err := op.Read()
	assert.NoError(t, err)
	assert.Equal(t, "1st good read", string(record))

	assert.True(t, op.Next())
	record, err = op.Read()
	assert.Error(t, err)
	assert.True(t, errs.IsErrTransformFailed(err))
	assert.Equal(t, continuableErr1.Error(), err.Error())
	assert.Nil(t, record)

	assert.True(t, op.Next())
	record, err = op.Read()
	assert.NoError(t, err)
	assert.Equal(t, "2nd good read", string(record))

	assert.False(t, op.Next())
	record, err = op.Read()
	assert.Error(t, err)
	assert.Equal(t, errs.ErrEOF, err)
	assert.Nil(t, record)

	// Verifying when EOF is reached, repeatedly calling Next will still get you EOF.
	assert.False(t, op.Next())
	record, err = op.Read()
	assert.Error(t, err)
	assert.Equal(t, errs.ErrEOF, err)
	assert.Nil(t, record)
}

func TestTransformOp_EndWithNonContinuableError(t *testing.T) {
	op := &transformOp{
		inputProcessor: &testInputProcessor{
			readCalls: []testReadCall{
				{record: []byte("1st good read")},
				{err: errors.New("fatal error")},
			},
		},
	}
	assert.True(t, op.Next())
	record, err := op.Read()
	assert.NoError(t, err)
	assert.Equal(t, "1st good read", string(record))

	assert.True(t, op.Next())
	record, err = op.Read()
	assert.Error(t, err)
	assert.False(t, errs.IsErrTransformFailed(err))
	assert.Equal(t, "fatal error", err.Error())
	assert.Nil(t, record)

	// Verifying when fatal error occurred, repeatedly calling Next/Read will still get you the same err
	assert.False(t, op.Next())
	record, err = op.Read()
	assert.Error(t, err)
	assert.Equal(t, "fatal error", err.Error())
	assert.Nil(t, record)
}

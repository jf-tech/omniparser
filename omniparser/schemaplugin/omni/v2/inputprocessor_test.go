package omniv2

import (
	"errors"
	"fmt"
	"testing"

	node "github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
)

var testContinuableErr = errors.New("continuable error")

type testReader struct {
	result []*node.Node
	err    []error
}

func (r *testReader) Read() (*node.Node, error) {
	if len(r.result) == 0 {
		return nil, errs.ErrEOF
	}
	result := r.result[0]
	err := r.err[0]
	r.result = r.result[1:]
	r.err = r.err[1:]
	return result, err
}

func (r *testReader) IsContinuableError(err error) bool { return err == testContinuableErr }

func (r *testReader) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("ctx: "+format, args...)
}

func TestInputProcessor_Read_ReadFailure(t *testing.T) {
	p := &inputProcessor{
		reader: &testReader{result: []*node.Node{nil}, err: []error{errors.New("test failure")}},
	}
	b, err := p.Read()
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())
	assert.Nil(t, b)
}

func TestInputProcessor_Read_ParseNodeFailure(t *testing.T) {
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		[]byte(` {
			"transform_declarations": {
				"FINAL_OUTPUT": { "const": "abc", "result_type": "int" }
			}
		}`), nil, nil)
	assert.NoError(t, err)
	p := &inputProcessor{
		finalOutputDecl: finalOutputDecl,
		reader:          &testReader{result: []*node.Node{nil}, err: []error{nil}},
	}
	b, err := p.Read()
	assert.Error(t, err)
	assert.True(t, errs.IsErrTransformFailed(err))
	assert.True(t, p.IsContinuableError(err))
	assert.Equal(t,
		`ctx: fail to transform. err: fail to convert value 'abc' to type 'int' on 'FINAL_OUTPUT', err: strconv.ParseFloat: parsing "abc": invalid syntax`,
		err.Error())
	assert.Nil(t, b)
}

func TestInputProcessor_Read_Success(t *testing.T) {
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		[]byte(` {
			"transform_declarations": {
				"FINAL_OUTPUT": { "const": "123", "result_type": "int" }
			}
		}`), nil, nil)
	assert.NoError(t, err)
	p := &inputProcessor{
		finalOutputDecl: finalOutputDecl,
		reader:          &testReader{result: []*node.Node{nil}, err: []error{nil}},
	}
	b, err := p.Read()
	assert.NoError(t, err)
	assert.Equal(t, "123", string(b))
}

func TestIsContinuableError(t *testing.T) {
	p := &inputProcessor{reader: &testReader{}}
	assert.False(t, p.IsContinuableError(errors.New("test failure")))
	assert.True(t, p.IsContinuableError(testContinuableErr))
}

func TestFmtErr(t *testing.T) {
	p := &inputProcessor{reader: &testReader{}}
	assert.Equal(t, "ctx: some 1 fruit", p.FmtErr("some %d %s", 1, "fruit").Error())
}

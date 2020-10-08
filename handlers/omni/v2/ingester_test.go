package omniv2

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
	"github.com/jf-tech/omniparser/idr"
)

var errContinuableInTest = errors.New("continuable error")

type testReader struct {
	result []*idr.Node
	err    []error
}

func (r *testReader) Read() (*idr.Node, error) {
	if len(r.result) == 0 {
		return nil, errs.ErrEOF
	}
	result := r.result[0]
	err := r.err[0]
	r.result = r.result[1:]
	r.err = r.err[1:]
	return result, err
}

func (r *testReader) IsContinuableError(err error) bool { return err == errContinuableInTest }

func (r *testReader) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("ctx: "+format, args...)
}

func TestIngester_Read_ReadFailure(t *testing.T) {
	g := &ingester{
		reader: &testReader{result: []*idr.Node{nil}, err: []error{errors.New("test failure")}},
	}
	b, err := g.Read()
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())
	assert.Nil(t, b)
}

func TestIngester_Read_ParseNodeFailure(t *testing.T) {
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		[]byte(` {
			"transform_declarations": {
				"FINAL_OUTPUT": { "const": "abc", "result_type": "int" }
			}
		}`), nil, nil)
	assert.NoError(t, err)
	g := &ingester{
		finalOutputDecl: finalOutputDecl,
		reader:          &testReader{result: []*idr.Node{nil}, err: []error{nil}},
	}
	b, err := g.Read()
	assert.Error(t, err)
	assert.True(t, errs.IsErrTransformFailed(err))
	assert.True(t, g.IsContinuableError(err))
	assert.Equal(t,
		`ctx: fail to transform. err: fail to convert value 'abc' to type 'int' on 'FINAL_OUTPUT', err: strconv.ParseFloat: parsing "abc": invalid syntax`,
		err.Error())
	assert.Nil(t, b)
}

func TestIngester_Read_Success(t *testing.T) {
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		[]byte(` {
			"transform_declarations": {
				"FINAL_OUTPUT": { "const": "123", "result_type": "int" }
			}
		}`), nil, nil)
	assert.NoError(t, err)
	g := &ingester{
		finalOutputDecl: finalOutputDecl,
		reader:          &testReader{result: []*idr.Node{nil}, err: []error{nil}},
	}
	b, err := g.Read()
	assert.NoError(t, err)
	assert.Equal(t, "123", string(b))
}

func TestIsContinuableError(t *testing.T) {
	g := &ingester{reader: &testReader{}}
	assert.False(t, g.IsContinuableError(errors.New("test failure")))
	assert.True(t, g.IsContinuableError(errContinuableInTest))
}

func TestFmtErr(t *testing.T) {
	g := &ingester{reader: &testReader{}}
	assert.Equal(t, "ctx: some 1 fruit", g.FmtErr("some %d %s", 1, "fruit").Error())
}

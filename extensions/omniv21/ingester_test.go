package omniv21

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/idr"
)

var errContinuableInTest = errors.New("continuable error")
var ingesterTestNode = idr.CreateNode(idr.ElementNode, "test")

type testReader struct {
	result        []*idr.Node
	err           []error
	releaseCalled int
}

func (r *testReader) Read() (*idr.Node, error) {
	if len(r.result) == 0 {
		return nil, io.EOF
	}
	result := r.result[0]
	err := r.err[0]
	r.result = r.result[1:]
	r.err = r.err[1:]
	return result, err
}

func (r *testReader) Release(_ *idr.Node) { r.releaseCalled++ }

func (r *testReader) IsContinuableError(err error) bool { return err == errContinuableInTest }

func (r *testReader) FmtErr(format string, args ...interface{}) error {
	return fmt.Errorf("ctx: "+format, args...)
}

func TestIngester_Read_ReadFailure(t *testing.T) {
	g := &ingester{
		reader: &testReader{result: []*idr.Node{nil}, err: []error{errors.New("test failure")}},
	}
	raw, b, err := g.Read()
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())
	assert.Nil(t, raw)
	assert.Nil(t, b)
	assert.Equal(t, 0, g.reader.(*testReader).releaseCalled)
}

func TestIngester_Read_ParseNodeFailure(t *testing.T) {
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		[]byte(` {
			"transform_declarations": {
				"FINAL_OUTPUT": { "const": "abc", "type": "int" }
			}
		}`), nil, nil)
	assert.NoError(t, err)
	g := &ingester{
		finalOutputDecl: finalOutputDecl,
		reader:          &testReader{result: []*idr.Node{ingesterTestNode}, err: []error{nil}},
	}
	raw, b, err := g.Read()
	assert.Error(t, err)
	assert.True(t, errs.IsErrTransformFailed(err))
	assert.True(t, g.IsContinuableError(err))
	assert.Equal(t,
		`ctx: fail to transform. err: unable to convert value 'abc' to type 'int' on 'FINAL_OUTPUT', err: strconv.ParseInt: parsing "abc": invalid syntax`,
		err.Error())
	assert.Nil(t, raw)
	assert.Nil(t, b)
	assert.Equal(t, 0, g.reader.(*testReader).releaseCalled)
}

func TestIngester_Read_Success(t *testing.T) {
	finalOutputDecl, err := transform.ValidateTransformDeclarations(
		[]byte(` {
			"transform_declarations": {
				"FINAL_OUTPUT": { "const": "123", "type": "int" }
			}
		}`), nil, nil)
	assert.NoError(t, err)
	g := &ingester{
		finalOutputDecl: finalOutputDecl,
		reader:          &testReader{result: []*idr.Node{ingesterTestNode}, err: []error{nil}},
	}
	raw, b, err := g.Read()
	assert.NoError(t, err)
	assert.Equal(t, "41665284-dab9-300d-b647-7ace9cb514b4", raw.Checksum())
	assert.Equal(t, "{}", idr.JSONify2(raw.Raw().(*idr.Node)))
	assert.Equal(t, "123", string(b))
	assert.Equal(t, 0, g.reader.(*testReader).releaseCalled)
	raw, b, err = g.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, raw)
	assert.Nil(t, b)
	assert.Equal(t, 1, g.reader.(*testReader).releaseCalled)
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

package omniv2

import (
	"encoding/json"
	"errors"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/errs"
	omniv2fileformat "github.com/jf-tech/omniparser/handlers/omni/v2/fileformat"
	"github.com/jf-tech/omniparser/handlers/omni/v2/transform"
	"github.com/jf-tech/omniparser/transformctx"
)

type ingester struct {
	finalOutputDecl  *transform.Decl
	customFuncs      customfuncs.CustomFuncs
	customParseFuncs transform.CustomParseFuncs
	ctx              *transformctx.Ctx
	reader           omniv2fileformat.FormatReader
}

func (g *ingester) Read() ([]byte, error) {
	node, err := g.reader.Read()
	if err != nil {
		// Read() supposed to have already done CtxAwareErr error wrapping. So directly return.
		return nil, err
	}
	result, err := transform.NewParseCtx(
		g.ctx, g.customFuncs, g.customParseFuncs).ParseNode(node, g.finalOutputDecl)
	if err != nil {
		// ParseNode() error not CtxAwareErr wrapped, so wrap it.
		// Note errs.ErrorTransformFailed is a continuable error.
		return nil, errs.ErrTransformFailed(g.fmtErrStr("fail to transform. err: %s", err.Error()))
	}
	return json.Marshal(result)
}

func (g *ingester) IsContinuableError(err error) bool {
	return errs.IsErrTransformFailed(err) || g.reader.IsContinuableError(err)
}

func (g *ingester) FmtErr(format string, args ...interface{}) error {
	return errors.New(g.fmtErrStr(format, args...))
}

func (g *ingester) fmtErrStr(format string, args ...interface{}) string {
	return g.reader.FmtErr(format, args...).Error()
}

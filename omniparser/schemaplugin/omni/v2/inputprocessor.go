package omniv2

import (
	"encoding/json"
	"errors"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

type inputProcessor struct {
	finalOutputDecl *transform.Decl
	customFuncs     customfuncs.CustomFuncs
	ctx             *transformctx.Ctx
	reader          InputReader
}

func (p *inputProcessor) Read() ([]byte, error) {
	node, err := p.reader.Read()
	if err != nil {
		// Read() supposed to have already done CtxAwareErr error wrapping. So directly return.
		return nil, err
	}
	result, err := transform.NewParseCtx(p.ctx, p.customFuncs).ParseNode(node, p.finalOutputDecl)
	if err != nil {
		// ParseNode() error not CtxAwareErr wrapped, so wrap it.
		// Note errs.ErrorTransformFailed is a continuable error.
		return nil, errs.ErrTransformFailed(p.fmtErrStr("fail to transform. err: %s", err.Error()))
	}
	return json.Marshal(result)
}

func (p *inputProcessor) IsContinuableError(err error) bool {
	return errs.IsErrTransformFailed(err) || p.reader.IsContinuableError(err)
}

func (p *inputProcessor) FmtErr(format string, args ...interface{}) error {
	return errors.New(p.fmtErrStr(format, args...))
}

func (p *inputProcessor) fmtErrStr(format string, args ...interface{}) string {
	return p.reader.FmtErr(format, args...).Error()
}

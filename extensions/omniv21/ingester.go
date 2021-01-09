package omniv21

import (
	"encoding/json"
	"errors"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/transformctx"
)

// RawRecord contains the raw data ingested in from the input stream in the form of an IDR tree.
// Note callers outside this package should absolutely make **NO** modifications to the content of
// RawRecord. Treat it like read-only.
type RawRecord struct {
	Node *idr.Node
}

// UUIDv3 returns a stable MD5(v3) hash of the RawRecord.
func (rr *RawRecord) UUIDv3() string {
	hash, _ := customfuncs.UUIDv3(nil, idr.JSONify2(rr.Node))
	return hash
}

type ingester struct {
	finalOutputDecl  *transform.Decl
	customFuncs      customfuncs.CustomFuncs
	customParseFuncs transform.CustomParseFuncs // Deprecated.
	ctx              *transformctx.Ctx
	reader           fileformat.FormatReader
	rawRecord        RawRecord
}

// Read ingests a raw record from the input stream, transforms it according the given schema and return
// the raw record, transformed JSON bytes.
func (g *ingester) Read() (interface{}, []byte, error) {
	if g.rawRecord.Node != nil {
		g.reader.Release(g.rawRecord.Node)
		g.rawRecord.Node = nil
	}
	n, err := g.reader.Read()
	if n != nil {
		g.rawRecord.Node = n
	}
	if err != nil {
		// Read() supposed to have already done CtxAwareErr error wrapping. So directly return.
		return nil, nil, err
	}
	result, err := transform.NewParseCtx(g.ctx, g.customFuncs, g.customParseFuncs).ParseNode(n, g.finalOutputDecl)
	if err != nil {
		// ParseNode() error not CtxAwareErr wrapped, so wrap it.
		// Note errs.ErrorTransformFailed is a continuable error.
		return nil, nil, errs.ErrTransformFailed(g.fmtErrStr("fail to transform. err: %s", err.Error()))
	}
	transformed, err := json.Marshal(result)
	return &g.rawRecord, transformed, err
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

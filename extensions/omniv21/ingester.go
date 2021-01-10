package omniv21

import (
	"encoding/json"
	"errors"

	"github.com/jf-tech/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/fileformat"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/idr"
	"github.com/jf-tech/omniparser/schemahandler"
	"github.com/jf-tech/omniparser/transformctx"
)

type rawRecord struct {
	node *idr.Node
}

func (rr *rawRecord) Raw() interface{} {
	return rr.node
}

// Checksum returns a stable MD5(v3) hash of the rawRecord.
func (rr *rawRecord) Checksum() string {
	hash, _ := customfuncs.UUIDv3(nil, idr.JSONify2(rr.node))
	return hash
}

type ingester struct {
	finalOutputDecl  *transform.Decl
	customFuncs      customfuncs.CustomFuncs
	customParseFuncs transform.CustomParseFuncs // Deprecated.
	ctx              *transformctx.Ctx
	reader           fileformat.FormatReader
	rawRecord        rawRecord
}

// Read ingests a raw record from the input stream, transforms it according the given schema and return
// the raw record, transformed JSON bytes.
func (g *ingester) Read() (schemahandler.RawRecord, []byte, error) {
	if g.rawRecord.node != nil {
		g.reader.Release(g.rawRecord.node)
		g.rawRecord.node = nil
	}
	n, err := g.reader.Read()
	if n != nil {
		g.rawRecord.node = n
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

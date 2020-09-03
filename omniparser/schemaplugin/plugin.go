package schemaplugin

import (
	"io"

	"github.com/jf-tech/omniparser/omniparser/customfuncs"
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

type ParseSchemaCtx struct {
	Name        string
	Header      Header
	Content     []byte
	CustomFuncs customfuncs.CustomFuncs
}

// ParseSchemaFunc is a type of a func that checks if a given schema is supported by
// its associated plugin, and, if yes, parses the schema content, creates and initializes
// a new instance of its associated plugin.
// If the given schema is not supported, errs.ErrSchemaNotSupported should be returned.
// Any other error returned will cause omniparser to fail entirely.
// Note, any non errs.ErrSchemaNotSupported error returned here should be errs.CtxAwareErr
// formatted (i.e. error should contain schema name and if possible error line number).
type ParseSchemaFunc func(ctx *ParseSchemaCtx) (Plugin, error)

// Plugin is an interface representing a schema plugin responsible for processing input
// stream based on its given schema.
type Plugin interface {
	// GetInputProcessor returns an InputProcessor for an input stream.
	// Omniparser will not call GetInputProcessor unless ParseSchema has returned supported.
	// Omniparser calls GetInputProcessor when client supplies an input stream and is ready
	// for the parser to process/transform the input.
	GetInputProcessor(ctx *transformctx.Ctx, input io.Reader) (InputProcessor, error)
}

// InputProcessor is an interface responsible for a given input stream
// parsing and processing.
type InputProcessor interface {
	// Read is called repeatedly during the processing of an input stream. Each call it should return
	// one result object, called record. It's entirely up to the implementation of this interface/method
	// to decide whether internally it does all the processing all at once (such as in the very first call
	// of `Read()`) and only hands out one record object at a time, OR, processes and returns one record
	// for each call. However, the overall design principle of omniparser is to have streaming processing
	// capability so memory won't be a constraint when dealing with large input file. All built-in plugin
	// and processors are done this way.
	Read() ([]byte, error)

	// IsContinuableError is called to determine if the error returned by Read is fatal or not. After Read
	// is called, the result record or error will be returned to caller. After caller consumes record or
	// error, omniparser needs to decide whether to continue the transform operation or not, based on
	// whether the last err is "continuable" or not.
	IsContinuableError(error) bool

	// CtxAwareErr interface is embedded to provide omniparser and custom functions a way to provide
	// context aware (such as input file name + line number) error formatting.
	errs.CtxAwareErr
}

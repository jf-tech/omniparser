package schemaplugin

import (
	"errors"
	"io"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

// ErrSchemaNotSupported indicates a schema is not supported by a plugin.
var ErrSchemaNotSupported = errors.New("schema not supported")

// SchemaParserFunc is a type of a func that checks if a given schema is supported by
// its associated plugin, and, if yes, parses the schema content, creates and initializes
// a new instance of its associated plugin.
// If the given schema is not supported, ErrSchemaNotSupported should be returned.
// Any other error returned will cause omniparser to fail entirely.
// Note, any non ErrSchemaNotSupported error returned here should be errs.CtxAwareErr
// formatted (i.e. error should contain schema name and if possible error line number).
type SchemaParserFunc func(name string, header Header, content []byte) (Plugin, error)

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
}

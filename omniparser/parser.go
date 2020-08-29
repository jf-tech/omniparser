package omniparser

import (
	"io"

	"github.com/jf-tech/omniparser/omniparser/schemaPlugin"
	"github.com/jf-tech/omniparser/omniparser/transformCtx"
)

// Parser is an interface that represents an instance of omniparser.
// One instance of Parser is associated with one and only one schema.
// The instance of Parser can be reused for parsing and transforming
// multiple input files/streams, as long as they are all intended for the
// same schema.
// Each parsing/transform, however, needs a separate instance of
// TransformOp. TransformOp must not be shared and reused across different
// input files/streams.
// While the same instance of Parser can be shared across multiple threads,
// TransformOp is not multi-thread safe. All operations on it must be done
// within the same go routine.
type Parser interface {
	GetTransformOp(name string, input io.Reader, ctx *transformCtx.Ctx) (TransformOp, error)
	SchemaHeader() schemaPlugin.Header
	SchemaRawContent() string
}

package fileformat

import (
	"io"

	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/extensions/omniv21/transform"
	"github.com/jf-tech/omniparser/header"
	"github.com/jf-tech/omniparser/idr"
)

// FileFormat defines a specific file format.
type FileFormat interface {
	// ValidateSchema validates the file format specific portion of the schema and returns
	// any needed runtime data that can be used in later constructed FormatReader. If this
	// file format isn't a match for the one specified by 'format' argument, it must return
	// errs.ErrSchemaNotSupported. All other errs must be ctx aware formatted.
	ValidateSchema(format string, schemaContent []byte, finalOutputDecl *transform.Decl) (interface{}, error)

	// CreateFormatReader creates an FormatReader which reads records of input data for this file format.
	CreateFormatReader(
		schemaHeader header.Header,
		inputName string,
		input io.Reader,
		formatRuntime interface{}) (FormatReader, error)
}

// FormatReader is an interface for reading a specific input format in omni schema handler. We'll have
// a number of format specific readers. The omni schema handler will use these readers for loading input
// stream content before doing the xpath/node based parsing.
type FormatReader interface {
	// Read returns a *Node and its subtree that will eventually be parsed and transformed into an
	// output record. If EOF has been reached, io.EOF must be returned.
	Read() (*idr.Node, error)
	// Release gives the reader a chance to free resources of the *Node and its subtree that it returned
	// to caller in a previous Read() call.
	Release(*idr.Node)
	// IsContinuableError determines whether an FormatReader returned error is continuable or not.
	// For certain errors (like EOF or corruption) there is no point to keep on trying; while others
	// can be safely ignored.
	IsContinuableError(err error) bool
	// CtxAwareErr allows FormatReader to format an error by providing context information (such as input
	// file name and (approx.) error location, such as line number)
	errs.CtxAwareErr
}

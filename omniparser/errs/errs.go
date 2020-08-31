package errs

import (
	"errors"
	"io"
)

// ErrSchemaNotSupported indicates a schema is not supported by a plugin.
var ErrSchemaNotSupported = errors.New("schema not supported")

// ErrEOF indicates the end of input stream has reached.
var ErrEOF = io.EOF

// ErrTransformFailed indicates a particular record transform has failed. In general
// this isn't fatal, and processing can continue.
type ErrTransformFailed string

// Error implements the error interface
func (e ErrTransformFailed) Error() string { return string(e) }

// IsErrTransformFailed tells if an error is of ErrTransformFailed.
func IsErrTransformFailed(err error) bool {
	switch err.(type) {
	case ErrTransformFailed:
		return true
	default:
		return false
	}
}

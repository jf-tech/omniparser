package omniparser

import (
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/schemahandler"
)

// Transform is an interface that represents one input stream ingestion and transform
// operation. An instance of a Transform must not be shared and reused among different
// input streams. An instance of a Transform must not be used across multiple goroutines.
type Transform interface {
	// Read returns a JSON byte slice representing one ingested and transformed record.
	// io.EOF should be returned when input stream is completely consumed and future calls
	// to Read should always return io.EOF.
	// errs.ErrTransformFailed should be returned when a record ingestion and transformation
	// failed and such failure isn't considered fatal. Future calls to Read will attempt
	// new record ingestion and transformations.
	// Any other error returned is considered fatal and future calls to Read will always
	// return the same error.
	// Note if returned error isn't nil, then returned []byte will be nil.
	Read() ([]byte, error)
}

type transform struct {
	ingester schemahandler.Ingester
	lastErr  error
}

// Read returns a JSON byte slice representing one ingested and transformed record.
// io.EOF should be returned when input stream is completely consumed and future calls
// to Read should always return io.EOF.
// errs.ErrTransformFailed should be returned when a record ingestion and transformation
// failed and such failure isn't considered fatal. Future calls to Read will attempt
// new record ingestion and transformations.
// Any other error returned is considered fatal and future calls to Read will always
// return the same error.
// Note if returned error isn't nil, then returned []byte will be nil.
func (o *transform) Read() ([]byte, error) {
	// errs.ErrTransformFailed is a generic wrapping error around all handlers' ingesters'
	// **continuable** errors (so client side doesn't have to deal with myriad of different
	// types of benign continuable errors. All other errors: non-continuable errors or io.EOF
	// should cause the operation to cease.
	if o.lastErr != nil && !errs.IsErrTransformFailed(o.lastErr) {
		return nil, o.lastErr
	}
	record, err := o.ingester.Read()
	if err != nil {
		if o.ingester.IsContinuableError(err) {
			// If ingester error is continuable, wrap it into a standard generic ErrTransformFailed
			// so caller has an easier time to deal with it. If fatal error, then leave it raw to the
			// caller so they can decide what it is and how to proceed.
			err = errs.ErrTransformFailed(err.Error())
		}
		record = nil
	}
	o.lastErr = err
	return record, err
}

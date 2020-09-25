package omniparser

import (
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/handlers"
)

// Transform is an interface that represents one input stream ingestion and transform
// operation. An instance of a Transform must not be shared and reused among different
// input streams. An instance of a Transform must not be used across multiple goroutines.
type Transform interface {
	// Next indicates whether the ingestion/transform is completed or not.
	Next() bool
	// Read returns a JSON byte slice representing one ingested and transformed record.
	Read() ([]byte, error)
}

type transform struct {
	ingester  handlers.Ingester
	curErr    error
	curRecord []byte
}

// Next calls the underlying schema handler's ingester to do the ingestion and transformation, saves
// the resulting record and/or error, and returns whether the entire operation is completed or not.
func (o *transform) Next() bool {
	// ErrTransformFailed is a generic wrapping error around all handlers' ingesters'
	// **continuable** errors (so client side doesn't have to deal with myriad of different
	// types of benign continuable errors. All other errors: non-continuable errors or ErrEOF
	// should cause the operation to cease.
	if o.curErr != nil && !errs.IsErrTransformFailed(o.curErr) {
		return false
	}

	for {
		record, err := o.ingester.Read()
		switch {
		case err == nil:
			o.curRecord = record
			o.curErr = nil
			return true
		case err == errs.ErrEOF:
			o.curErr = err
			o.curRecord = nil
			// No more processing needed.
			return false
		default:
			o.curErr = err
			// If ingester error is continuable, wrap it into a standard generic ErrTransformFailed
			// so caller has an easier time to deal with it. If fatal error, then leave it raw to the
			// caller so they can decide what it is and how to proceed.
			if o.ingester.IsContinuableError(err) {
				o.curErr = errs.ErrTransformFailed(err.Error())
			}
			o.curRecord = nil
			return true
		}
	}
}

// Read returns the current ingested and transformed record and/or the current error.
func (o *transform) Read() ([]byte, error) {
	return o.curRecord, o.curErr
}

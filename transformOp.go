package omniparser

import (
	"github.com/jf-tech/omniparser/errs"
	"github.com/jf-tech/omniparser/schemaplugin"
)

// TransformOp is an interface that represents one input stream parsing/transform operation.
// Instance of TransformOp must not be shared and reused among different input streams.
// Instance of TransformOp must not be used across multiple goroutines.
type TransformOp interface {
	// Next indicates whether the parsing/transform operation is completed or not.
	Next() bool
	// Read returns a JSON byte slice representing one parsing/transform record.
	Read() ([]byte, error)
}

type transformOp struct {
	inputProcessor schemaplugin.InputProcessor
	curErr         error
	curRecord      []byte
}

// Next gives the associated InputProcessor to do input processor and returns whether the entire
// operation is done or not.
func (o *transformOp) Next() bool {
	// ErrTransformFailed is a generic wrapping error around all plugins' InputProcessors'
	// **continuable** errors (so client side doesn't have to deal with myriad of different
	// types of "benign" continuable errors. All other errors: non-continuable errors or ErrEOF
	// should cause the operation to cease.
	if o.curErr != nil && !errs.IsErrTransformFailed(o.curErr) {
		return false
	}

	for {
		record, err := o.inputProcessor.Read()
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
			// Only wrap the plugin error into ErrorTransformFailed when it's continuable error; if it's not,
			// we do want to pop the plugin specific raw non-continuable error up so client can choose to
			// make some sense out of why the operation has ceased and follow up accordingly such as cleanup,
			// error report, stats, etc)
			if o.inputProcessor.IsContinuableError(err) {
				o.curErr = errs.ErrTransformFailed(err.Error())
			}
			o.curRecord = nil
			// return true so client can call Read() to get the error.
			return true
		}
	}
}

// Read returns the current transformed record or the current error.
func (o *transformOp) Read() ([]byte, error) {
	return o.curRecord, o.curErr
}

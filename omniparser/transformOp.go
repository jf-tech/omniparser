package omniparser

// TransformOp is an interface that represents one input stream parsing/transform operation.
// Instance of TransformOp must not be shared and reused among different input streams.
// Instance of TransformOp should not be used across multiple goroutines.
type TransformOp interface {
	// Next indicates whether the parsing/transform operation is completed or not.
	Next() bool
	// Read returns a JSON byte slice representing one parsing/transform result.
	Read() ([]byte, error)
	// Parser returns the Parser from which this TransformOp is created.
	Parser() Parser
}

package errs

// CtxAwareErr formats and creates an error is context aware: e.g. during schema parsing and
// validation, implementation of this interface can give us errors prefixed with schema name
// and line number. During input ingestion and transforming, it can give us errors prefixed
// with input stream (file) name and line number, etc.
type CtxAwareErr interface {
	FmtErr(format string, args ...interface{}) error
}

package testlib

import (
	"bytes"
	"errors"
	"io"
)

type alwaysFailReadCloser struct{ err error }

// Read always returns the designated error.
func (a alwaysFailReadCloser) Read([]byte) (int, error) { return 0, a.err }

// Close always returns the designated error.
func (a alwaysFailReadCloser) Close() error { return a.err }

type bytesReadCloser struct{ underlying io.Reader }

func (b bytesReadCloser) Read(p []byte) (n int, err error) { return b.underlying.Read(p) }
func (bytesReadCloser) Close() error                       { return nil }

func NewMockReadCloser(failureMsg string, content []byte) io.ReadCloser {
	if failureMsg != "" {
		return alwaysFailReadCloser{errors.New(failureMsg)}
	}
	return bytesReadCloser{bytes.NewReader(content)}
}

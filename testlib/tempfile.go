package testlib

import (
	"io/ioutil"
	"os"

	"github.com/jf-tech/omniparser/strs"
)

// For dir/pattern params check https://golang.org/pkg/io/ioutil/#TempFile
// Caller is responsible for calling os.Remove on the returned file.
func CreateTempFileWithContent(dir, pattern, content string) (*os.File, error) {
	f, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		return nil, err
	}

	// If content is empty, no need to write
	if !strs.IsStrNonBlank(content) {
		_ = f.Close()
		return f, nil
	}

	if _, err := f.Write([]byte(content)); err != nil {
		_ = os.Remove(f.Name())
		return nil, err
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return nil, err
	}

	return f, nil
}

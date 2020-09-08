package testlib

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/strs"
)

// CreateTempFileWithContent creates a temp file with desired content in it
// for testing. For dir/pattern params check https://golang.org/pkg/io/ioutil/#TempFile
// Caller is responsible for calling os.Remove on the returned file.
func CreateTempFileWithContent(t *testing.T, dir, pattern, content string) *os.File {
	f, err := ioutil.TempFile(dir, pattern)
	assert.NoError(t, err)
	// If content is empty, no need to write
	if !strs.IsStrNonBlank(content) {
		_ = f.Close()
		return f
	}
	_, err = f.Write([]byte(content))
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)
	return f
}

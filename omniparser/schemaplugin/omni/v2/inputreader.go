package omniv2

import (
	node "github.com/antchfx/xmlquery"

	"github.com/jf-tech/omniparser/omniparser/errs"
)

// InputReader is an interface for reading input stream in omni plugin. We'll have a number of file
// format specific readers. The omni plugin will use these readers for loading input stream content
// before doing the xpath/node based parsing.
type InputReader interface {
	// Read returns a *Node and its subtree that will eventually be parsed and transformed into an
	// output record.
	Read() (*node.Node, error)
	// IsContinuableError determines whether an InputReader returned error is continuable or not.
	// For certain errors (like EOF or corruption) there is no point to keep on trying; while others
	// can be safely ignored.
	IsContinuableError(err error) bool
	// InputReader must be able to format an error by providing context information (such as input
	// file name and (approx.) error location, such as line number)
	errs.CtxAwareErr
}

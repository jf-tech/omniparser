package flatfile

import (
	"github.com/jf-tech/omniparser/idr"
)

// RecReader defines an interface for a flat file specific format to read data and match
// a RecDecl.
type RecReader interface {
	// MoreUnprocessedData indicates whether there is any unprocessed data left in the input
	// stream. Possible return values:
	// - (true, nil): more data is available.
	// - (false, nil): no more data is available.
	// - (_, err): some and most likely fatal IO error has occurred.
	// Implementation notes:
	// - If some data is read in and io.EOF is encountered, (true, nil) should be returned.
	// - If no data is read in and io.EOF is encountered, (false, nil) should be returned.
	// - Under no circumstances, io.EOF should be returned.
	// - Once a call to this method returns (false, nil), all future calls to it should always
	//   return (false, nil).
	MoreUnprocessedData() (more bool, err error)

	// ReadAndMatch matches the passed-in *non-group* RecDecl to unprocessed data and creates a
	// corresponding IDR node if data matches and createIDR flag is turned on.
	// Implementation notes:
	// - If io.EOF is encountered while there is still unmatched thus unprocessed data,
	//   io.EOF shouldn't be returned.
	// - If a non io.EOF error encountered during IO, return (false, nil, err).
	ReadAndMatch(decl RecDecl, createIDR bool) (matched bool, node *idr.Node, err error)
}

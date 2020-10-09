package omniv2csv

import (
	"github.com/jf-tech/go-corelib/strs"
)

type column struct {
	Name string `json:"name"`
	// If the CSV column 'name' contains characters (such as space, or special letters) that are
	// not suitable for *idr.Node construction and xpath query, this gives schema writer an
	// alternate way to name/label the column. Optional.
	Alias *string `json:"alias"`
}

func (c column) name() string {
	return strs.StrPtrOrElse(c.Alias, c.Name)
}

type fileDecl struct {
	Delimiter           string   `json:"delimiter"`
	ReplaceDoubleQuotes bool     `json:"replace_double_quotes"`
	HeaderRowIndex      *int     `json:"header_row_index"`
	DataRowIndex        int      `json:"data_row_index"`
	Columns             []column `json:"columns"`
}

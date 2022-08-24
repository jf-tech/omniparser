package flatfile

// RecDecl defines a flat file record. It is meant to be a common facade for all the
// actual unit of processing from each format, such as envelope from fixed length,
// segment from EDI, and record from csv, etc.
type RecDecl interface {
	DeclName() string // to avoid collision, since most decl has Name as a field.
	Target() bool
	Group() bool
	MinOccurs() int
	MaxOccurs() int
	ChildDecls() []RecDecl
}

const (
	rootName = "#root"
)

type rootDecl struct {
	children []RecDecl
}

func (d rootDecl) DeclName() string      { return rootName }
func (d rootDecl) Target() bool          { return false }
func (d rootDecl) Group() bool           { return true }
func (d rootDecl) MinOccurs() int        { return 1 }
func (d rootDecl) MaxOccurs() int        { return 1 }
func (d rootDecl) ChildDecls() []RecDecl { return d.children }

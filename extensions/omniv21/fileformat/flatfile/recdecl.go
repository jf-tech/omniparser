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

// Design note: flatfile.fixedlength, flatfile.csv, etc all have similar structs that contain name,
// is_target, type, min, max, etc and implementation of RecDecl interface all very similar. So why
// not just change RecDecl into a struct and embed that struct into format specific decl structs?
// We chose not to do that and rather take a small hit of code duplication in preference to
// flexibility. Note RecDecl's ChildDecls need to return child decls and each format specific decl
// is different, so that's the first incompatibility here that ChildDecls cannot be even be included
// in the struct if we go down the RecDecl being struct route. Yes, generics can do that but we
// don't want to move omniparser 1.14 dependency up all the way to 1.18 simply because of this.
// Second, depending on formats, the default values for min/max are different: csv/fixed-length
// min/max default to 0/-1, but for EDI min/max default to 1/1. Given these incompatibility and
// loss of flexibility, we chose to stick with the RecDecl interface route and have each format
// somewhat duplicate a small amount of trivial code.

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

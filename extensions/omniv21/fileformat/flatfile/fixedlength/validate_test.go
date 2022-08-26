package fixedlength

import (
	"testing"

	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestValidateFileDecl_AutoTargetFirstEnvelope(t *testing.T) {
	decl := &FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A"},
		},
	}
	assert.False(t, decl.Envelopes[0].Target())
	err := (&validateCtx{}).validateFileDecl(decl)
	assert.NoError(t, err)
	assert.True(t, decl.Envelopes[0].Target())
}

func TestValidateFileDecl_InvalidHeaderRegexp(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", Header: strs.StrPtr("[invalid")},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"envelope/envelope_group 'A' has an invalid 'header' regexp '[invalid': error parsing regexp: missing closing ]: `[invalid`",
		err.Error())
}

func TestValidateFileDecl_InvalidFooterRegexp(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", Footer: strs.StrPtr("[invalid")},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"envelope/envelope_group 'A' has an invalid 'footer' regexp '[invalid': error parsing regexp: missing closing ]: `[invalid`",
		err.Error())
}

func TestValidateFileDecl_GroupHasColumns(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{
				Name:     "A",
				Type:     strs.StrPtr(typeGroup),
				Columns:  []*ColumnDecl{{}},
				Children: []*EnvelopeDecl{{}},
			},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `envelope_group 'A' must not have any columns`, err.Error())
}

func TestValidateFileDecl_GroupHasNoChildren(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", Type: strs.StrPtr(typeGroup), IsTarget: true},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		`envelope_group 'A' must have at least one child envelope/envelope_group`, err.Error())
}

func TestValidateFileDecl_TwoIsTarget(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", IsTarget: true},
			{Name: "B", Type: strs.StrPtr(typeGroup), Children: []*EnvelopeDecl{
				{Name: "C", IsTarget: true},
			}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		`a second envelope/envelope_group ('B/C') with 'is_target' = true is not allowed`,
		err.Error())
}

func TestValidateFileDecl_MinGreaterThanMax(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", Children: []*EnvelopeDecl{
				{Name: "B", Min: testlib.IntPtr(2), Max: testlib.IntPtr(1)}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `envelope/envelope_group 'A/B' has 'min' value 2 > 'max' value 1`, err.Error())
}

func TestValidateFileDecl_ColumnLineIndexAndLinePatternSameTime(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", Columns: []*ColumnDecl{
				{Name: "c", LineIndex: testlib.IntPtr(2), LinePattern: strs.StrPtr(".")}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"envelope 'A' column 'c' cannot have both `line_index` and `line_pattern` specified at the same time",
		err.Error())
}

func TestValidateFileDecl_InvalidColumnLinePattern(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Envelopes: []*EnvelopeDecl{
			{Name: "A", Columns: []*ColumnDecl{
				{Name: "c", LinePattern: strs.StrPtr("[invalid")}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"envelope 'A' column 'c' has an invalid 'line_pattern' regexp '[invalid': error parsing regexp: missing closing ]: `[invalid`",
		err.Error())
}

func TestValidateFileDecl_Success(t *testing.T) {
	col1 := &ColumnDecl{Name: "c1", LineIndex: testlib.IntPtr(1)}
	col2 := &ColumnDecl{Name: "c2"}
	col3 := &ColumnDecl{Name: "c3", LinePattern: strs.StrPtr("^C$")}
	fd := &FileDecl{
		Envelopes: []*EnvelopeDecl{
			{
				Name:   "A",
				Header: strs.StrPtr("^A_BEGIN$"),
				Footer: strs.StrPtr("^A_END$"),
				Children: []*EnvelopeDecl{
					{
						Name: "B", IsTarget: true,
						Columns: []*ColumnDecl{col1, col2, col3},
					},
				},
			},
		},
	}
	err := (&validateCtx{}).validateFileDecl(fd)
	assert.NoError(t, err)
	assert.Equal(t, "A", fd.Envelopes[0].fqdn)
	assert.True(t, fd.Envelopes[0].matchHeader([]byte("A_BEGIN")))
	assert.True(t, fd.Envelopes[0].matchFooter([]byte("A_END")))
	assert.Equal(t, 1, len(fd.Envelopes[0].childRecDecls))
	assert.Same(t, fd.Envelopes[0].Children[0], fd.Envelopes[0].childRecDecls[0].(*EnvelopeDecl))
	assert.Equal(t, "A/B", fd.Envelopes[0].Children[0].fqdn)
	assert.Equal(t, []*ColumnDecl{col1, col2, col3}, fd.Envelopes[0].Children[0].Columns)
	assert.True(t, fd.Envelopes[0].Children[0].Columns[0].lineMatch(0, []byte("C")))
	assert.True(t, fd.Envelopes[0].Children[0].Columns[2].lineMatch(0, []byte("C")))
}

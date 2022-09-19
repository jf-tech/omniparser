package csv

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"
)

func TestValidateFileDecl_AutoTargetFirstRecord(t *testing.T) {
	decl := &FileDecl{
		Records: []*RecordDecl{
			{Name: "A"},
			{Name: "B"},
		},
	}
	assert.False(t, decl.Records[0].Target())
	assert.False(t, decl.Records[1].Target())
	err := (&validateCtx{}).validateFileDecl(decl)
	assert.NoError(t, err)
	assert.True(t, decl.Records[0].Target())
	assert.False(t, decl.Records[1].Target())
}

func TestValidateFileDecl_InvalidHeaderRegexp(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", Header: strs.StrPtr("[invalid")},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"record/record_group 'A' has an invalid 'header' regexp '[invalid': error parsing regexp: missing closing ]: `[invalid`",
		err.Error())
}

func TestValidateFileDecl_InvalidFooterRegexp(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", Footer: strs.StrPtr("[invalid")},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"record/record_group 'A' has an invalid 'footer' regexp '[invalid': error parsing regexp: missing closing ]: `[invalid`",
		err.Error())
}

func TestValidateFileDecl_GroupHasColumns(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{
				Name:     "A",
				Type:     strs.StrPtr(typeGroup),
				Columns:  []*ColumnDecl{{}},
				Children: []*RecordDecl{{}},
			},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `record_group 'A' must not have any columns`, err.Error())
}

func TestValidateFileDecl_GroupHasNoChildren(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", Type: strs.StrPtr(typeGroup), IsTarget: true},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		`record_group 'A' must have at least one child record/record_group`, err.Error())
}

func TestValidateFileDecl_TwoIsTarget(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", IsTarget: true},
			{Name: "B", Type: strs.StrPtr(typeGroup), Children: []*RecordDecl{
				{Name: "C", IsTarget: true},
			}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		`a second record/record_group ('B/C') with 'is_target' = true is not allowed`,
		err.Error())
}

func TestValidateFileDecl_MinGreaterThanMax(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", Children: []*RecordDecl{
				{Name: "B", Min: testlib.IntPtr(2), Max: testlib.IntPtr(1)}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, `record/record_group 'A/B' has 'min' value 2 > 'max' value 1`, err.Error())
}

func TestValidateFileDecl_ColumnLineIndexAndLinePatternSameTime(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", Columns: []*ColumnDecl{
				{Name: "c", LineIndex: testlib.IntPtr(2), LinePattern: strs.StrPtr(".")}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"record 'A' column 'c' cannot have both `line_index` and `line_pattern` specified at the same time",
		err.Error())
}

func TestValidateFileDecl_InvalidColumnLinePattern(t *testing.T) {
	err := (&validateCtx{}).validateFileDecl(&FileDecl{
		Records: []*RecordDecl{
			{Name: "A", Columns: []*ColumnDecl{
				{Name: "c", LinePattern: strs.StrPtr("[invalid")}}},
		},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"record 'A' column 'c' has an invalid 'line_pattern' regexp '[invalid': error parsing regexp: missing closing ]: `[invalid`",
		err.Error())
}

func TestValidateFileDecl_Success(t *testing.T) {
	col1 := &ColumnDecl{Name: "c1", LineIndex: testlib.IntPtr(1)}
	col2 := &ColumnDecl{Name: "c2", Index: testlib.IntPtr(3)}
	col3 := &ColumnDecl{Name: "c3", LinePattern: strs.StrPtr("^C$")}
	fd := &FileDecl{
		Records: []*RecordDecl{
			{
				Name:   "A",
				Header: strs.StrPtr("^A_BEGIN$"),
				Footer: strs.StrPtr("^A_END$"),
				Children: []*RecordDecl{
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
	assert.Equal(t, "A", fd.Records[0].fqdn)
	assert.True(t,
		fd.Records[0].matchHeader(&line{recordStart: 0, recordNum: 1}, []string{"A_BEGIN"}, ","))
	assert.True(t,
		fd.Records[0].matchFooter(&line{recordStart: 0, recordNum: 1}, []string{"A_END"}, ","))
	assert.Equal(t, 1, len(fd.Records[0].childRecDecls))
	assert.Same(t, fd.Records[0].Children[0], fd.Records[0].childRecDecls[0].(*RecordDecl))
	assert.Equal(t, "A/B", fd.Records[0].Children[0].fqdn)
	cupaloy.SnapshotT(t, fd.Records[0].Children[0].Columns)
}

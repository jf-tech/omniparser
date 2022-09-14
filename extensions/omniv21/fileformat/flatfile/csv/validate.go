package csv

import (
	"fmt"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"
)

type validateCtx struct {
	seenTarget bool
}

func (ctx *validateCtx) validateFileDecl(fileDecl *FileDecl) error {
	for _, decl := range fileDecl.Records {
		if err := ctx.validateRecordDecl(decl.Name, decl); err != nil {
			return err
		}
	}
	if !ctx.seenTarget && len(fileDecl.Records) > 0 {
		// for easy of use and convenience, if no is_target=true record is specified, then
		// the first one will be automatically designated as target record.
		fileDecl.Records[0].IsTarget = true
	}
	return nil
}

func (ctx *validateCtx) validateRecordDecl(fqdn string, decl *RecordDecl) (err error) {
	decl.fqdn = fqdn
	if decl.Header != nil {
		if decl.headerRegexp, err = caches.GetRegex(*decl.Header); err != nil {
			return fmt.Errorf(
				"record/record_group '%s' has an invalid 'header' regexp '%s': %s",
				fqdn, *decl.Header, err.Error())
		}
	}
	if decl.Footer != nil {
		if decl.footerRegexp, err = caches.GetRegex(*decl.Footer); err != nil {
			return fmt.Errorf(
				"record/record_group '%s' has an invalid 'footer' regexp '%s': %s",
				fqdn, *decl.Footer, err.Error())
		}
	}
	if decl.Group() {
		if len(decl.Columns) > 0 {
			return fmt.Errorf("record_group '%s' must not have any columns", fqdn)
		}
		if len(decl.Children) <= 0 {
			return fmt.Errorf(
				"record_group '%s' must have at least one child record/record_group", fqdn)
		}
	}
	if decl.Target() {
		if ctx.seenTarget {
			return fmt.Errorf(
				"a second record/record_group ('%s') with 'is_target' = true is not allowed",
				fqdn)
		}
		ctx.seenTarget = true
	}
	if decl.MinOccurs() > decl.MaxOccurs() {
		return fmt.Errorf("record/record_group '%s' has 'min' value %d > 'max' value %d",
			fqdn, decl.MinOccurs(), decl.MaxOccurs())
	}
	for i, col := range decl.Columns {
		prevCol := (*ColumnDecl)(nil)
		if i > 0 {
			prevCol = decl.Columns[i-1]
		}
		if err = ctx.validateColumnDecl(fqdn, prevCol, col); err != nil {
			return err
		}
	}
	for _, c := range decl.Children {
		if err = ctx.validateRecordDecl(strs.BuildFQDN2("/", fqdn, c.Name), c); err != nil {
			return err
		}
	}
	decl.childRecDecls = toFlatFileRecDecls(decl.Children)
	return nil
}

func intPtr(v int) *int {
	return &v
}

func (ctx *validateCtx) validateColumnDecl(fqdn string, prevDecl, decl *ColumnDecl) (err error) {
	// If column.index not specified, then we'll use the previous column's index value + 1 unless
	// the column is the first column, then 1 will be used.
	// if column.index is explicitly specified, it will be honored.
	if decl.Index == nil {
		if prevDecl == nil {
			decl.Index = intPtr(1)
		} else {
			decl.Index = intPtr(*prevDecl.Index + 1)
		}
	}
	if decl.LineIndex != nil && decl.LinePattern != nil {
		return fmt.Errorf(
			"record '%s' column '%s' cannot have both `line_index` and `line_pattern` specified at the same time",
			fqdn, decl.Name)
	}
	if decl.LinePattern != nil {
		if decl.linePatternRegexp, err = caches.GetRegex(*decl.LinePattern); err != nil {
			return fmt.Errorf(
				"record '%s' column '%s' has an invalid 'line_pattern' regexp '%s': %s",
				fqdn, decl.Name, *decl.LinePattern, err.Error())
		}
	}
	return nil
}

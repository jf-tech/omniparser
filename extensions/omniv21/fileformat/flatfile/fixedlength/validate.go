package fixedlength

import (
	"fmt"

	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/strs"
)

type validateCtx struct {
	seenTarget bool
}

func (ctx *validateCtx) validateFileDecl(fileDecl *FileDecl) error {
	for _, envelopeDecl := range fileDecl.Envelopes {
		if err := ctx.validateEnvelopeDecl(envelopeDecl.Name, envelopeDecl); err != nil {
			return err
		}
	}
	if !ctx.seenTarget && len(fileDecl.Envelopes) > 0 {
		// for easy of use and convenience, if no is_target=true envelope is specified, then
		// the first one will be automatically designated as target envelope.
		fileDecl.Envelopes[0].IsTarget = true
	}
	return nil
}

func (ctx *validateCtx) validateEnvelopeDecl(fqdn string, envelopeDecl *EnvelopeDecl) (err error) {
	envelopeDecl.fqdn = fqdn
	if envelopeDecl.Header != nil {
		if envelopeDecl.headerRegexp, err = caches.GetRegex(*envelopeDecl.Header); err != nil {
			return fmt.Errorf(
				"envelope/envelope_group '%s' has an invalid 'header' regexp '%s': %s",
				fqdn, *envelopeDecl.Header, err.Error())
		}
	}
	if envelopeDecl.Footer != nil {
		if envelopeDecl.footerRegexp, err = caches.GetRegex(*envelopeDecl.Footer); err != nil {
			return fmt.Errorf(
				"envelope/envelope_group '%s' has an invalid 'footer' regexp '%s': %s",
				fqdn, *envelopeDecl.Footer, err.Error())
		}
	}
	if envelopeDecl.Group() {
		if len(envelopeDecl.Columns) > 0 {
			return fmt.Errorf("envelope_group '%s' must not have any columns", fqdn)
		}
		if len(envelopeDecl.Children) <= 0 {
			return fmt.Errorf(
				"envelope_group '%s' must have at least one child envelope/envelope_group", fqdn)
		}
	}
	if envelopeDecl.Target() {
		if ctx.seenTarget {
			return fmt.Errorf(
				"a second envelope/envelope_group ('%s') with 'is_target' = true is not allowed",
				fqdn)
		}
		ctx.seenTarget = true
	}
	if envelopeDecl.MinOccurs() > envelopeDecl.MaxOccurs() {
		return fmt.Errorf("envelope/envelope_group '%s' has 'min' value %d > 'max' value %d",
			fqdn, envelopeDecl.MinOccurs(), envelopeDecl.MaxOccurs())
	}
	for _, colDecl := range envelopeDecl.Columns {
		if err = ctx.validateColumnDecl(fqdn, colDecl); err != nil {
			return err
		}
	}
	for _, c := range envelopeDecl.Children {
		if err = ctx.validateEnvelopeDecl(strs.BuildFQDN2("/", fqdn, c.Name), c); err != nil {
			return err
		}
	}
	envelopeDecl.childRecDecls = toFlatFileRecDecls(envelopeDecl.Children)
	return nil
}

func (ctx *validateCtx) validateColumnDecl(fqdn string, colDecl *ColumnDecl) (err error) {
	if colDecl.LineIndex != nil && colDecl.LinePattern != nil {
		return fmt.Errorf(
			"envelope '%s' column '%s' cannot have both `line_index` and `line_pattern` specified at the same time",
			fqdn, colDecl.Name)
	}
	if colDecl.LinePattern != nil {
		if colDecl.linePatternRegexp, err = caches.GetRegex(*colDecl.LinePattern); err != nil {
			return fmt.Errorf(
				"envelope '%s' column '%s' has an invalid 'line_pattern' regexp '%s': %s",
				fqdn, colDecl.Name, *colDecl.LinePattern, err.Error())
		}
	}
	return nil
}

package edi

import (
	"errors"
	"fmt"

	"github.com/jf-tech/go-corelib/strs"
)

type ediValidateCtx struct {
	seenTarget bool
}

func (ctx *ediValidateCtx) validateFileDecl(fileDecl *FileDecl) error {
	for _, segDecl := range fileDecl.SegDecls {
		if err := ctx.validateSegDecl(segDecl.Name, segDecl); err != nil {
			return err
		}
	}
	if !ctx.seenTarget {
		return errors.New("missing segment/segment_group with 'is_target' = true")
	}
	return nil
}

func (ctx *ediValidateCtx) validateSegDecl(segFQDN string, segDecl *SegDecl) error {
	segDecl.fqdn = segFQDN
	if segDecl.minOccurs() > segDecl.maxOccurs() {
		return fmt.Errorf(
			"segment '%s' has 'min' value %d > 'max' value %d", segFQDN, segDecl.minOccurs(), segDecl.maxOccurs())
	}
	if segDecl.IsTarget {
		if ctx.seenTarget {
			return fmt.Errorf("a second segment/group ('%s') with 'is_target' = true is not allowed", segFQDN)
		}
		ctx.seenTarget = true
	}
	if segDecl.isGroup() && len(segDecl.Children) <= 0 {
		return fmt.Errorf("segment_group '%s' must have at least one child segment/segment_group", segFQDN)
	}
	for _, child := range segDecl.Children {
		err := ctx.validateSegDecl(strs.BuildFQDN2(fqdnDelim, segFQDN, child.Name), child)
		if err != nil {
			return err
		}
	}
	return nil
}

package edi

import (
	"github.com/jf-tech/go-corelib/maths"
	"github.com/jf-tech/go-corelib/strs"
)

const (
	segTypeSeg   = "segment"
	segTypeGroup = "segment_group"
)

const (
	fqdnDelim = "/"
)

type elem struct {
	Name           string `json:"name,omitempty"`
	Index          int    `json:"index,omitempty"`
	ComponentIndex *int   `json:"component_index,omitempty"`
	EmptyIfMissing bool   `json:"empty_if_missing,omitempty"`
}

type segDecl struct {
	Name     string     `json:"name,omitempty"`
	Type     *string    `json:"type,omitempty"`
	IsTarget bool       `json:"is_target,omitempty"`
	Min      *int       `json:"min,omitempty"`
	Max      *int       `json:"max,omitempty"`
	Elems    []elem     `json:"elements,omitempty"`
	Children []*segDecl `json:"child_segments,omitempty"`
}

func (d *segDecl) isGroup() bool {
	return d.Type != nil && *d.Type == segTypeGroup
}

func (d *segDecl) minOccurs() int {
	switch d.Min {
	case nil:
		return 1
	default:
		return *d.Min
	}
}

func (d *segDecl) maxOccurs() int {
	switch {
	case d.Max == nil:
		return 1
	case *d.Max < 0:
		return maths.MaxIntValue
	default:
		return *d.Max
	}
}

type segDeclRuntime struct {
	decl           *segDecl
	children       []*segDeclRuntime
	parent         *segDeclRuntime
	containsTarget bool // whether this seg itself or any of its children has is_target == true
	occurred       int
	fqdn           string
}

const (
	rootSegName = "#root"
)

func newSegDeclRuntimeTree(segDecls ...*segDecl) *segDeclRuntime {
	return newSegDeclRuntime(
		nil,
		&segDecl{
			Name:     rootSegName,
			Type:     strs.StrPtr(segTypeGroup),
			Children: segDecls,
		})
}

func newSegDeclRuntime(parent *segDeclRuntime, segDecl *segDecl) *segDeclRuntime {
	rt := segDeclRuntime{decl: segDecl, parent: parent}
	if parent != nil {
		// we don't set fqdn on root.
		if parent.fqdn != "" {
			rt.fqdn = strs.BuildFQDN2(fqdnDelim, parent.fqdn, segDecl.Name)
		} else {
			// we don't include rootSegName as part of fqdn.
			rt.fqdn = segDecl.Name
		}
	}
	containsTarget := segDecl.IsTarget
	for _, child := range segDecl.Children {
		childRt := newSegDeclRuntime(&rt, child)
		containsTarget = containsTarget || childRt.containsTarget
		rt.children = append(rt.children, childRt)
	}
	rt.containsTarget = containsTarget
	return &rt
}

func (rt *segDeclRuntime) matchSegName(segName string) bool {
	switch rt.decl.isGroup() {
	case true:
		// Group (or so called loop) itself doesn't have a segment name in EDI file (we do assign a
		// name to it for xpath query reference, but that name isn't a segment name per se). A
		// group/loop's first non-group child, recursively if necessary, can be used as the group's
		// identifying segment name, per EDI standard. Meaning if a group's first non-group child's
		// segment exists in an EDI file, then this group has an instance in the file. If the first
		// non-group child's segment isn't found, then we (the standard) assume the group is skipped.
		// The explanation can be found:
		//  - https://www.gxs.co.uk/wp-content/uploads/tutorial_ansi.pdf (around page 9), quote
		//    "...loop is optional, but if any segment in the loop is used, the first segment
		//    within the loop becomes mandatory..."
		//  - https://github.com/smooks/smooks-edi-cartridge/blob/54f97e89156114e13e1acd3b3c46fe9a4234918c/edi-sax/src/main/java/org/smooks/edi/edisax/model/internal/SegmentGroup.java#L68
		return len(rt.children) > 0 && rt.children[0].matchSegName(segName)
	default:
		return rt.decl.Name == segName
	}
}

func (rt *segDeclRuntime) resetChildrenOccurred() {
	for _, child := range rt.children {
		child.occurred = 0
		child.resetChildrenOccurred()
	}
}

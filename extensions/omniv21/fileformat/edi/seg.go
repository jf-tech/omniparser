package edi

import (
	"github.com/jf-tech/go-corelib/maths"
)

// variable/func naming guide:
//
// full name      | short name
// -----------------------------------
// segment        | seg
// component      | comp
// element        | elem
// node           | n
// reader         | r
// delimiter      | delim
// current        | cur
// number         | num
// declaration    | decl
// character      | char

const (
	segTypeSeg   = "segment"
	segTypeGroup = "segment_group"
)

const (
	fqdnDelim   = "/"
	rootSegName = "#root"
)

// Elem describes an element inside an EDI segment.
type Elem struct {
	Name           string  `json:"name,omitempty"`
	Index          int     `json:"index,omitempty"`
	CompIndex      *int    `json:"component_index,omitempty"`
	EmptyIfMissing bool    `json:"empty_if_missing,omitempty"` // Deprecated, use Default
	Default        *string `json:"default,omitempty"`
}

func (e Elem) compIndex() int {
	if e.CompIndex == nil {
		return 1
	}
	return *e.CompIndex
}

// SegDecl describes an EDI segment declaration/settings.
type SegDecl struct {
	Name     string     `json:"name,omitempty"`
	Type     *string    `json:"type,omitempty"`
	IsTarget bool       `json:"is_target,omitempty"`
	Min      *int       `json:"min,omitempty"`
	Max      *int       `json:"max,omitempty"`
	Elems    []Elem     `json:"elements,omitempty"`
	Children []*SegDecl `json:"child_segments,omitempty"`
	fqdn     string     // internal computed field
}

func (d *SegDecl) isGroup() bool {
	return d.Type != nil && *d.Type == segTypeGroup
}

func (d *SegDecl) minOccurs() int {
	switch d.Min {
	case nil:
		// for majority cases, segments have min=1, max=1, so default nil to 1
		return 1
	default:
		return *d.Min
	}
}

func (d *SegDecl) maxOccurs() int {
	switch {
	case d.Max == nil:
		// for majority cases, segments have min=1, max=1, so default nil to 1
		return 1
	case *d.Max < 0:
		// typically, schema writer uses -1 to indicate infinite; practically max int is good enough :)
		return maths.MaxIntValue
	default:
		return *d.Max
	}
}

func (d *SegDecl) matchSegName(segName string) bool {
	switch d.isGroup() {
	case true:
		// Group (or so-called loop) itself doesn't have a segment name in EDI file (we do assign a
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
		return len(d.Children) > 0 && d.Children[0].matchSegName(segName)
	default:
		return d.Name == segName
	}
}

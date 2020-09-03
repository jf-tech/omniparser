package transform

import (
	"encoding/json"

	"github.com/jf-tech/omniparser/strs"
)

type Kind string

const (
	KindConst      Kind = "const"
	KindExternal   Kind = "external"
	KindField      Kind = "field"
	KindObject     Kind = "object"
	KindArray      Kind = "array"
	KindCustomFunc Kind = "custom_func"
	KindTemplate   Kind = "template"
)

type ResultType string

const (
	ResultTypeUnknown ResultType = "unknown"
	ResultTypeInt     ResultType = "int"
	ResultTypeFloat   ResultType = "float"
	ResultTypeBoolean ResultType = "boolean"
	ResultTypeString  ResultType = "string"
	ResultTypeObject  ResultType = "object"
	ResultTypeArray   ResultType = "array"
)

const (
	FinalOutput = "FINAL_OUTPUT"
)

type CustomFuncDecl struct {
	Name                         string  `json:"name,omitempty"`
	Args                         []*Decl `json:"args,omitempty"`
	IgnoreErrorAndReturnEmptyStr bool    `json:"ignore_error_and_return_empty_str,omitempty"`
	fqdn                         string  // internal; never unmarshaled from a schema.
}

func (d CustomFuncDecl) MarshalJSON() ([]byte, error) {
	type Alias CustomFuncDecl
	return json.Marshal(&struct {
		Alias
		Fqdn string `json:"fqdn,omitempty"` // Marshal into JSON for test snapshots.
	}{
		Alias: Alias(d),
		Fqdn:  d.fqdn,
	})
}

// Note only deep-copy all the public fields, those internal computed fields are not copied.
func (d *CustomFuncDecl) deepCopy() *CustomFuncDecl {
	dest := &CustomFuncDecl{}
	dest.Name = d.Name
	dest.Args = nil
	for _, argDecl := range d.Args {
		dest.Args = append(dest.Args, argDecl.deepCopy())
	}
	dest.IgnoreErrorAndReturnEmptyStr = d.IgnoreErrorAndReturnEmptyStr
	return dest
}

// This is the struct will be unmarshaled from `transform_declarations` section of an omni schema.
type Decl struct {
	// Applicable for KindConst.
	Const *string `json:"const,omitempty"`
	// Applicable for KindExternal
	External *string `json:"external,omitempty"`
	// Applicable for KindField, KindObject, KindTemplate, KindCustomFunc
	XPath *string `json:"xpath,omitempty"`
	// Applicable for KindField, KindObject, KindTemplate, KindCustomFunc
	XPathDynamic *Decl `json:"xpath_dynamic,omitempty"`
	// Applicable for KindCustomFunc.
	CustomFunc *CustomFuncDecl `json:"custom_func,omitempty"`
	// Applicable for KindTemplate.
	Template *string `json:"template,omitempty"`
	// Applicable for KindObject.
	Object map[string]*Decl `json:"object,omitempty"`
	// Applicable for KindArray.
	Array []*Decl `json:"array,omitempty"`
	// Applicable for KindConst, KindExternal, KindField or KindCustomFunc.
	ResultType               *ResultType `json:"result_type,omitempty"`
	KeepLeadingTrailingSpace bool        `json:"keep_leading_trailing_space,omitempty"`
	KeepEmptyOrNull          bool        `json:"keep_empty_or_null,omitempty"`

	// Internal runtime fields that are not unmarshaled from a schema.
	fqdn     string
	kind     Kind
	hash     string
	children []*Decl
	parent   *Decl
}

func (d Decl) MarshalJSON() ([]byte, error) {
	emptyToNil := func(s string) string {
		return strs.FirstNonBlank(s, "(nil)")
	}
	type Alias Decl
	return json.Marshal(&struct {
		Alias
		FQDN string `json:"fqdn,omitempty"`
		Kind string `json:"kind,omitempty"`
		// skip hash as it is generated from uuid and would otherwise cause unit test snapshot failures
		Children []string `json:"children,omitempty"`
		Parent   string   `json:"parent,omitempty"`
	}{
		Alias: Alias(d),
		FQDN:  emptyToNil(d.fqdn),
		Kind:  emptyToNil(string(d.kind)),
		Children: func() []string {
			var fqdns []string
			for _, child := range d.children {
				fqdns = append(fqdns, emptyToNil(child.fqdn))
			}
			return fqdns
		}(),
		Parent: func() string {
			if d.parent != nil {
				return emptyToNil(d.parent.fqdn)
			}
			return emptyToNil("")
		}(),
	})
}

func (d *Decl) resultType() ResultType {
	switch d.ResultType {
	case nil:
		switch d.kind {
		case KindConst, KindExternal, KindField, KindCustomFunc:
			return ResultTypeString
		case KindObject:
			return ResultTypeObject
		case KindArray:
			return ResultTypeArray
		default:
			return ResultTypeUnknown
		}
	default:
		return *d.ResultType
	}
}

func (d *Decl) isPrimitiveKind() bool {
	switch d.kind {
	case KindConst, KindExternal, KindField, KindCustomFunc:
		// Don't put KindTemplate here because we don't know what actual kind the template
		// will resolve into: a template can resolve into a const/field/external/etc or it
		// can resolve into an array or object, so better be safe.
		return true
	default:
		return false
	}
}

func (d *Decl) isXPathSet() bool {
	return d.XPath != nil || d.XPathDynamic != nil
}

// Note only deep-copy all the public fields, those internal computed fields are not copied.
func (d *Decl) deepCopy() *Decl {
	dest := &Decl{}
	dest.Const = strs.CopyStrPtr(d.Const)
	dest.External = strs.CopyStrPtr(d.External)
	dest.XPath = strs.CopyStrPtr(d.XPath)
	if d.XPathDynamic != nil {
		dest.XPathDynamic = d.XPathDynamic.deepCopy()
	}
	if d.CustomFunc != nil {
		dest.CustomFunc = d.CustomFunc.deepCopy()
	}
	dest.Template = strs.CopyStrPtr(d.Template)
	if len(d.Object) > 0 {
		dest.Object = map[string]*Decl{}
		for childName, childDecl := range d.Object {
			dest.Object[childName] = childDecl.deepCopy()
		}
	}
	for _, childDecl := range d.Array {
		dest.Array = append(dest.Array, childDecl.deepCopy())
	}
	if d.ResultType != nil {
		rt := *d.ResultType
		dest.ResultType = &rt
	}
	dest.KeepLeadingTrailingSpace = d.KeepLeadingTrailingSpace
	dest.KeepEmptyOrNull = d.KeepEmptyOrNull
	return dest
}

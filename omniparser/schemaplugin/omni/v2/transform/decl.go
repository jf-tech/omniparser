package transform

import (
	"encoding/json"

	"github.com/jf-tech/omniparser/strs"
)

// Kind specifies the types of omni schema's input elements.
type Kind string

const (
	KindUnknown    Kind = "unknown"
	KindConst      Kind = "const"
	KindExternal   Kind = "external"
	KindField      Kind = "field"
	KindObject     Kind = "object"
	KindArray      Kind = "array"
	KindCustomFunc Kind = "custom_func"
	KindTemplate   Kind = "template"
)

// ResultType specifies the types of omni schema's output elements.
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
	// FinalOutput is the special name of a Decl that is designated for the output
	// for an omni schema.
	FinalOutput = "FINAL_OUTPUT"
)

// CustomFuncDecl is the decl for a "custom_func".
type CustomFuncDecl struct {
	Name                         string  `json:"name,omitempty"`
	Args                         []*Decl `json:"args,omitempty"`
	IgnoreErrorAndReturnEmptyStr bool    `json:"ignore_error_and_return_empty_str,omitempty"`
	fqdn                         string  // internal; never unmarshaled from a schema.
}

// MarshalJSON is the custom JSON marshaler for CustomFuncDecl.
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

// Decl is the type for omni schema's `transform_declarations` declarations.
type Decl struct {
	// Const indicates the input element is a cost.
	Const *string `json:"const,omitempty"`
	// External indicates the input element is from an external property.
	External *string `json:"external,omitempty"`
	// XPath specifies an xpath for an input element.
	XPath *string `json:"xpath,omitempty"`
	// XPathDynamic specifies a dynamically constructed xpath for an input element.
	XPathDynamic *Decl `json:"xpath_dynamic,omitempty"`
	// CustomFunc specifies the input element is a custom function.
	CustomFunc *CustomFuncDecl `json:"custom_func,omitempty"`
	// Template specifies the input element is a template.
	Template *string `json:"template,omitempty"`
	// Object specifies the input element is an object.
	Object map[string]*Decl `json:"object,omitempty"`
	// Array specifies the input element is an array.
	Array []*Decl `json:"array,omitempty"`
	// ResultType specifies the desired output type of an element.
	ResultType *ResultType `json:"result_type,omitempty"`
	// KeepLeadingTrailingSpace specifies space trimming in string value of the output element.
	KeepLeadingTrailingSpace bool `json:"keep_leading_trailing_space,omitempty"`
	// KeepEmptyOrNull specifies whether or not keep an empty/null output or not.
	KeepEmptyOrNull bool `json:"keep_empty_or_null,omitempty"`

	// Internal runtime fields that are not unmarshaled from a schema.
	fqdn     string
	kind     Kind
	hash     string
	children []*Decl
	parent   *Decl
}

// MarshalJSON is the custom JSON marshaler for Decl.
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

// Note only deep-copy all the public fields, those internal computed fields MUST not be copied:
// see explanation in validate.go's computeDeclHash().
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

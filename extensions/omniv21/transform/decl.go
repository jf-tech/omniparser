package transform

import (
	"encoding/json"

	"github.com/jf-tech/go-corelib/strs"
)

// Kind specifies the types of omni schema's input elements.
// Note in an actual schema, there is no such a field 'kind'. Kind
// is inferred.
type Kind string

const (
	// KindConst and rest are kinds for all transform decls.
	KindConst       Kind = "const"
	KindExternal    Kind = "external"
	KindField       Kind = "field"
	KindObject      Kind = "object"
	KindArray       Kind = "array"
	KindCustomFunc  Kind = "custom_func"
	KindCustomParse Kind = "custom_parse"
	KindTemplate    Kind = "template"
)

// ResultType specifies the types of omni schema's output elements.
// It corresponds to schema's 'type' field.
type ResultType string

const (
	// ResultTypeInt and rest are the possible ResultType values.
	ResultTypeInt     ResultType = "int"
	ResultTypeFloat   ResultType = "float"
	ResultTypeBoolean ResultType = "boolean"
	ResultTypeString  ResultType = "string"
)

const (
	// FinalOutput is the special name of a Decl that is designated for the output
	// for an omni schema.
	FinalOutput = "FINAL_OUTPUT"
)

// CustomFuncDecl is the decl for a "custom_func".
type CustomFuncDecl struct {
	Name        string  `json:"name,omitempty"`
	Args        []*Decl `json:"args,omitempty"`
	IgnoreError bool    `json:"ignore_error,omitempty"`
	fqdn        string  // internal; never unmarshaled from a schema.
}

// MarshalJSON is the custom JSON marshaler for CustomFuncDecl.
func (d CustomFuncDecl) MarshalJSON() ([]byte, error) {
	type Alias CustomFuncDecl
	return json.Marshal(&struct {
		Alias
		FQDN string `json:"fqdn,omitempty"` // Marshal into JSON for test snapshots.
	}{
		Alias: Alias(d),
		FQDN:  d.fqdn,
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
	dest.IgnoreError = d.IgnoreError
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
	// CustomParse specifies the input element is to be custom parsed.
	CustomParse *string `json:"custom_parse,omitempty"`
	// Template specifies the input element is a template.
	Template *string `json:"template,omitempty"`
	// Object specifies the input element is an object.
	Object map[string]*Decl `json:"object,omitempty"`
	// Array specifies the input element is an array.
	Array []*Decl `json:"array,omitempty"`
	// ResultType specifies the desired output type of an element.
	ResultType *ResultType `json:"type,omitempty"`
	// NoTrim specifies space trimming in string value of the output element.
	NoTrim bool `json:"no_trim,omitempty"`
	// KeepEmptyOrNull specifies whether or not keep an empty/null output or not.
	KeepEmptyOrNull bool `json:"keep_empty_or_null,omitempty"`

	// Internal fields are computed at schema loading time.
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

func (d *Decl) resolveKind() {
	switch {
	case d.Const != nil:
		d.kind = KindConst
	case d.External != nil:
		d.kind = KindExternal
	case d.CustomFunc != nil:
		d.kind = KindCustomFunc
	case d.CustomParse != nil:
		d.kind = KindCustomParse
	case d.Object != nil:
		d.kind = KindObject
	case d.Array != nil:
		d.kind = KindArray
	case d.Template != nil:
		d.kind = KindTemplate
	default:
		d.kind = KindField
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
	dest.CustomParse = strs.CopyStrPtr(d.CustomParse)
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
	dest.NoTrim = d.NoTrim
	dest.KeepEmptyOrNull = d.KeepEmptyOrNull
	return dest
}

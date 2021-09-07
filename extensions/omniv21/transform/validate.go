package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jf-tech/go-corelib/strs"

	"github.com/jf-tech/omniparser/customfuncs"
)

type validateCtx struct {
	Decls            map[string]*Decl `json:"transform_declarations"`
	customFuncs      customfuncs.CustomFuncs
	customParseFuncs CustomParseFuncs // Deprecated.
	declHashes       map[string]string
}

// ValidateTransformDeclarations validates `transform_declarations` section of an omni schema and returns
// the `FINAL_OUTPUT` corresponding Decl.
func ValidateTransformDeclarations(
	schemaContent []byte, customFuncs customfuncs.CustomFuncs, customParseFuncs CustomParseFuncs) (*Decl, error) {

	var ctx validateCtx
	// We did json schema validation earlier, so this unmarshal guarantees to succeed.
	_ = json.Unmarshal(schemaContent, &ctx)
	ctx.customFuncs = customFuncs
	ctx.customParseFuncs = customParseFuncs
	ctx.declHashes = map[string]string{}

	// We did json schema validation earlier, so "FINAL_OUTPUT" must exist.
	finalOutputDecl, err := ctx.validateDecl(finalOutput, ctx.Decls[finalOutput], []string{finalOutput})
	if err != nil {
		return nil, err
	}
	linkParent(finalOutputDecl)
	return finalOutputDecl, nil
}

// In order to detect circular template references (e.g. template A has a reference to template B which
// has a reference to C and C has one back to A), we need to keep a template reference stack, starting
// from the root template 'FINAL_OUTPUT'. Everytime we see a template, we push its name onto the stack.
// and check if it has appeared before or not.
func (ctx *validateCtx) validateDecl(fqdn string, decl *Decl, templateRefStack []string) (*Decl, error) {
	err := ctx.validateXPath(fqdn, decl, templateRefStack)
	if err != nil {
		return nil, err
	}
	decl.fqdn = fqdn
	decl.resolveKind()
	switch decl.kind {
	case kindObject:
		err := ctx.validateObject(fqdn, decl, templateRefStack)
		if err != nil {
			return nil, err
		}
	case kindArray:
		err := ctx.validateArray(fqdn, decl, templateRefStack)
		if err != nil {
			return nil, err
		}
	case kindCustomFunc:
		err := ctx.validateCustomFunc(fqdn, decl, templateRefStack)
		if err != nil {
			return nil, err
		}
	case kindCustomParse:
		err := ctx.validateCustomParse(fqdn, decl)
		if err != nil {
			return nil, err
		}
	case kindTemplate:
		decl, err = ctx.validateTemplate(fqdn, decl, templateRefStack)
		if err != nil {
			return nil, err
		}
	}
	decl.hash = computeDeclHash(decl, ctx.declHashes)
	return decl, nil
}

func (ctx *validateCtx) validateXPath(fqdn string, decl *Decl, templateRefStack []string) error {
	if decl.XPath != nil && decl.XPathDynamic != nil {
		return fmt.Errorf("'%s' cannot set both 'xpath' and 'xpath_dynamic' at the same time", fqdn)
	}
	// unlike `xpath` which is a constant string, `xpath_dynamic` value comes from the computation of
	// regular decl, and it can be of a const/field/custom_func/template/external, so we need to parse
	// and validate the decl as well.
	if decl.XPathDynamic != nil {
		var err error
		decl.XPathDynamic, err = ctx.validateDecl(
			strs.BuildFQDN(fqdn, "xpath_dynamic"), decl.XPathDynamic, templateRefStack)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *validateCtx) validateObject(fqdn string, decl *Decl, templateRefStack []string) error {
	for childName, childDecl := range decl.Object {
		childDecl, err := ctx.validateDecl(
			// childName can contain '.' or '%', it needs to be escaped.
			strs.BuildFQDN(fqdn, strs.BuildFQDNWithEsc(childName)), childDecl, templateRefStack)
		if err != nil {
			return err
		}
		decl.Object[childName] = childDecl
		decl.children = append(decl.children, childDecl)
	}
	// Sort the `children` array for unit test snapshot stability.
	// Given this schema parsing/loading is usually done infrequently, the sorting here shouldn't
	// incur too much latency penalty for production code path.
	if len(decl.children) > 0 {
		sort.Slice(decl.children, func(i, j int) bool { return decl.children[i].fqdn < decl.children[j].fqdn })
	}
	return nil
}

func (ctx *validateCtx) validateArray(fqdn string, decl *Decl, templateRefStack []string) error {
	for i, childDecl := range decl.Array {
		childDecl, err := ctx.validateDecl(
			strs.BuildFQDN(fqdn, fmt.Sprintf("elem[%d]", i+1)), childDecl, templateRefStack)
		if err != nil {
			return err
		}
		decl.Array[i] = childDecl
		decl.children = append(decl.children, childDecl)
	}
	// sort the `children` array for unit test snapshot stability.
	if len(decl.children) > 0 {
		sort.Slice(decl.children, func(i, j int) bool { return decl.children[i].fqdn < decl.children[j].fqdn })
	}
	return nil
}

func (ctx *validateCtx) validateCustomFunc(fqdn string, decl *Decl, templateRefStack []string) error {
	fn, found := ctx.customFuncs[decl.CustomFunc.Name]
	if !found {
		return fmt.Errorf("unknown custom_func '%s' on '%s'", decl.CustomFunc.Name, fqdn)
	}
	if reflect.ValueOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("custom_func '%s' is not a function", decl.CustomFunc.Name)
	}
	fnType := reflect.TypeOf(fn)
	if fnType.NumIn() < 1 {
		return fmt.Errorf("custom_func '%s' missing required ctx argument", decl.CustomFunc.Name)
	}
	if fnType.NumOut() != 2 {
		return fmt.Errorf(
			"custom_func '%s' must have 2 return values, instead got %d", decl.CustomFunc.Name, fnType.NumOut())
	}
	if !fnType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("custom_func '%s' 2nd return value must be of error type, instead got %s",
			decl.CustomFunc.Name, fnType.Out(1))
	}
	decl.CustomFunc.fqdn = strs.BuildFQDN(fqdn, fmt.Sprintf("custom_func(%s)", decl.CustomFunc.Name))
	for i := 0; i < len(decl.CustomFunc.Args); i++ {
		argDecl, err := ctx.validateDecl(
			strs.BuildFQDN(decl.CustomFunc.fqdn, fmt.Sprintf("arg[%d]", i+1)),
			decl.CustomFunc.Args[i],
			templateRefStack)
		if err != nil {
			return err
		}
		decl.CustomFunc.Args[i] = argDecl
		decl.children = append(decl.children, argDecl)
	}
	return nil
}

func (ctx *validateCtx) validateCustomParse(fqdn string, decl *Decl) error {
	if _, found := ctx.customParseFuncs[*decl.CustomParse]; !found {
		return fmt.Errorf("unknown custom_parse '%s' on '%s'", *decl.CustomParse, fqdn)
	}
	return nil
}

func (ctx *validateCtx) validateTemplate(fqdn string, decl *Decl, templateRefStack []string) (*Decl, error) {
	templateName := *decl.Template
	templateDecl, found := ctx.Decls[templateName]
	if !found {
		return nil, fmt.Errorf(
			"'%s' contains non-existing template reference '%s'", fqdn, templateName)
	}

	// need to make a copy otherwise slice is passed by reference and append might alter
	// the slice in place.
	templateRefStack = append(strs.CopySlice(templateRefStack), templateName)
	if strs.HasDup(templateRefStack) {
		return nil, fmt.Errorf("template circular dependency detected on '%s': %s",
			fqdn, strings.Join(
				strs.NoErrMapSlice(templateRefStack, func(s string) string { return "'" + s + "'" }),
				"->"))
	}

	// Make a copy in case the template is referenced in multiple places.
	declNew := templateDecl.deepCopy()
	// between the template site and the template itself, there can only be one decl with xpath/xpath_dynamic set.
	if declNew.isXPathSet() && decl.isXPathSet() {
		return nil, fmt.Errorf(
			"cannot specify 'xpath' or 'xpath_dynamic' on both '%s' and the template '%s' it references",
			fqdn, templateName)
	}
	if decl.isXPathSet() {
		declNew.XPath = decl.XPath
		declNew.XPathDynamic = decl.XPathDynamic
	}

	return ctx.validateDecl(fqdn, declNew, templateRefStack)
}

func computeDeclHash(decl *Decl, declHashes map[string]string) string {
	// We'd like to create a stable encoding of a decl then we can use it to lookup
	// in declHashes. If we find an existing entry, then use that entry's hash id as
	// the decl's hash. If we don't then create a new hash id for it and save into
	// declHashes.
	//
	// The key and difficulty is to have a STABLE encoding of a decl. Remember in
	// golang, the order of enumerating a map is non-deterministic, that makes the
	// problem somewhat hard. Luckily, we have json.Marshal, which according to golang
	// sorts the map's keys. So using json.Marshal gives us a stable encoding of a
	// struct.
	//
	// However, here is a problem. Decl has all the regular exported fields, as well
	// as those unexported runtime computation fields (such as `fqdn`, `kind`, even
	// `hash` itself). Now we're facing a dilemma:
	// - if we stick with the standard/built-in json marshaler for Decl, then those
	//   "hidden" computation fields won't be marshaled and that's bad for our unit
	//   test snapshots, in which we really want to see and ensure those computation
	//   fields' correctness.
	// - but if we do a custom json marshaler to include those computation fields,
	//   then here when we try to marshal a Decl into a string, that string would
	//   include those "hidden" fields. Guess what? By including `fqdn` in a Decl's
	//   encoding would make any two Decls unique no matter how identical they look
	//   alike, thus defeating the purpose of this decl hash computation.
	//
	// In golang we can't really switch json marshaler easily for a given type at
	// runtime (otherwise it would be easy: in unit tests, use custom json marshaler
	// to include those hidden fields; and in production code, use standard json
	// marshaler to ignore them). Solution:
	// - define the custom json marshaler for Decl to include those hidden files
	//   so unit tests are happy.
	// - here we first use deepCopy() to make a copy of the input Decl. Note
	//   deepCopy() only copies the public/exported fields.
	// - then use json marshaler to encode the new Decl copy into a stable json str.
	b, _ := json.Marshal(decl.deepCopy())
	declJson := string(b)
	if hash, found := declHashes[declJson]; found {
		return hash
	}
	declHash := uuid.New().String()
	declHashes[declJson] = declHash
	return declHash
}

func linkParent(decl *Decl) {
	for _, child := range decl.children {
		child.parent = decl
		linkParent(child)
	}
}

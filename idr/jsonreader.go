package idr

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/antchfx/xpath"
	"github.com/jf-tech/go-corelib/caches"
	"github.com/jf-tech/go-corelib/ios"
)

// JSONStreamReader is a streaming JSON to *Node reader.
type JSONStreamReader struct {
	r                          *ios.LineCountingReader
	d                          *json.Decoder
	xpathExpr, xpathFilterExpr *xpath.Expr
	root, cur, stream          *Node
}

// streamCandidateCheck checks if sp.cur is a potential stream candidate.
// Note even if sp.cur is marked, it only means it could potentially be the
// stream candidate, not guaranteed. E.g., say have a JSON doc looks like this:
// {
//    "x": {
//       "a": 5
//    },
//    "y": {
//       "a": 2
//    }
// }
// And our stream xpath is "/*/a[. < 4]".
// When the reader first encounters "x/a" node, it has the potential to be the
// stream candidate since it matches "/*/a". But we haven't finished processing
// the entire node, thus no way we know if the value of "a" would match the filter
// or not. So we have to mark it as a potential stream candidate and will let
// wrapUpCurAndTargetCheck to do the final check when the entire node of "/x/a"
// is ingested and processed, in which case, "/x/a" will be not be considered
// as stream target, but later "/x/b" will be.
func (sp *JSONStreamReader) streamCandidateCheck() {
	if sp.xpathExpr != nil && sp.stream == nil && MatchAny(sp.root, sp.xpathExpr) {
		sp.stream = sp.cur
	}
}

// wrapUpCurAndTargetCheck wraps sp.cur node processing and also checks if the sp.cur is the stream
// candidate and if it is, then does a final check: a stream candidate is the target if:
// - If it has finished processing (sp.cur == sp.stream)
// - Either we don't have a stream filter xpath or the stream filter xpath matches.
func (sp *JSONStreamReader) wrapUpCurAndTargetCheck() *Node {
	cur := sp.cur
	// No matter what outcome the wrapUpCurAndTargetCheck() is, the current node is done, and
	// we need to adjust sp.cur to its parent.
	sp.cur = sp.cur.Parent
	// Only do stream target check if the finished cur node is the stream candidate
	if cur != sp.stream {
		return nil
	}
	if sp.xpathFilterExpr == nil || MatchAny(sp.root, sp.xpathFilterExpr) {
		return sp.stream
	}
	// This means while the sp.stream was marked as a stream candidate by the initial
	// sp.streamCandidateCheck call, but now we've completed the construction of this
	// node fully and discovered sp.xpathFilterExpr can't be satisfied, so this
	// sp.stream isn't a target. To prevent future mismatch for other stream candidate,
	// we need to remove it from Node tree completely. And reset sp.stream.
	RemoveAndReleaseTree(sp.stream)
	sp.stream = nil
	return nil
}

func (sp *JSONStreamReader) addElementChild(data string, jtype JSONType) {
	child := CreateJSONNode(ElementNode, data, jtype)
	AddChild(sp.cur, child)
	sp.cur = child
}

func (sp *JSONStreamReader) addTextChild(tok interface{}) {
	var data string
	var jtype JSONType
	switch v := tok.(type) {
	case float64:
		data = strconv.FormatFloat(v, 'f', -1, 64)
		jtype = JSONValueNum
	case bool:
		data = strconv.FormatBool(v)
		jtype = JSONValueBool
	case nil:
		data = ""
		jtype = JSONValueNull
	default:
		data = v.(string)
		jtype = JSONValueStr
	}
	child := CreateJSONNode(TextNode, data, jtype)
	AddChild(sp.cur, child)
	// Since the child being added is a value node, there won't be anything else
	// added below it, so no need to advance sp.cur to child.
}

func (sp *JSONStreamReader) parseDelim(tok json.Delim) *Node {
	switch tok {
	case '{':
		switch {
		// arr check needs to be before prop or root check because it's entirely
		// possible sp.cur is a property or root whose value is array type - in that
		// case we want the arr case to be hit first.
		case IsJSONArr(sp.cur):
			// if we see "{" inside an "[]", we create an anonymous object element node
			// to host it.
			sp.addElementChild("", JSONObj)
			sp.streamCandidateCheck()
		case IsJSONProp(sp.cur):
			// a "{" follows a property name, indicate this property's value is an
			// object. Note we don't need to streamCandidateCheck here because we've
			// already done the check when the property itself is processed.
			sp.cur.FormatSpecific = JSONTypeOf(sp.cur) | JSONObj
		case IsJSONRoot(sp.cur):
			// if we see "{" directly on root, make the root node an obj type container
			// and do stream candidate check.
			sp.cur.FormatSpecific = JSONTypeOf(sp.cur) | JSONObj
			sp.streamCandidateCheck()
		}
	case '[':
		switch {
		// Similarly, case order matters.
		case IsJSONArr(sp.cur):
			// if we see "[" inside an "[]" or directly on root, we create an anonymous
			// arr element node to host it.
			sp.addElementChild("", JSONArr)
			sp.streamCandidateCheck()
		case IsJSONProp(sp.cur):
			// Again, similarly we don't do streamCandidateCheck here since the check is already
			// done when the property node is created.
			sp.cur.FormatSpecific = JSONTypeOf(sp.cur) | JSONArr
		case IsJSONRoot(sp.cur):
			// arr directly on root.
			sp.cur.FormatSpecific = JSONTypeOf(sp.cur) | JSONArr
			sp.streamCandidateCheck()
		}
	case '}', ']':
		ret := sp.wrapUpCurAndTargetCheck()
		if ret != nil {
			return ret
		}
	}
	return nil
}

func (sp *JSONStreamReader) parseVal(tok json.Token) *Node {
	switch {
	// Note case order matters, because cur type could be prop|obj or root|obj, in those
	// cases, we want IsJSONObj case to be hit first.
	case IsJSONObj(sp.cur):
		sp.addElementChild(tok.(string), JSONProp)
		sp.streamCandidateCheck()
	// Similarly, we want arr check before prop check.
	case IsJSONArr(sp.cur):
		// if parent is an array or root, so we're adding a value directly to
		// the array or root, by creating an anonymous element node, then the
		// value as text node underneath it.
		sp.addElementChild("", JSONProp)
		sp.streamCandidateCheck()
		sp.addTextChild(tok)
		ret := sp.wrapUpCurAndTargetCheck()
		if ret != nil {
			return ret
		}
	case IsJSONProp(sp.cur):
		sp.addTextChild(tok)
		ret := sp.wrapUpCurAndTargetCheck()
		if ret != nil {
			return ret
		}
	case IsJSONRoot(sp.cur):
		// A value is directly setting on root. We need to do both stream candidate check
		// and target check.
		sp.streamCandidateCheck()
		sp.addTextChild(tok)
		ret := sp.wrapUpCurAndTargetCheck()
		if ret != nil {
			return ret
		}
	}
	return nil
}

func (sp *JSONStreamReader) parse() (*Node, error) {
	for {
		tok, err := sp.d.Token()
		if err != nil {
			// including io.EOF
			return nil, err
		}
		switch tok := tok.(type) {
		case json.Delim:
			if ret := sp.parseDelim(tok); ret != nil {
				return ret, nil
			}
		case string, float64, bool, nil:
			if ret := sp.parseVal(tok); ret != nil {
				return ret, nil
			}
		}
	}
}

// Read returns a *Node that matches the xpath streaming criteria.
func (sp *JSONStreamReader) Read() (*Node, error) {
	// Because this is a streaming read, we need to release/remove last
	// stream node from the node tree to free up memory. If Release() is
	// called after Read() call, then sp.stream is already cleaned up;
	// adding this piece of code here just in case Release() isn't called.
	if sp.stream != nil {
		RemoveAndReleaseTree(sp.stream)
		sp.stream = nil
	}
	return sp.parse()
}

// Release releases the *Node (and its subtree) that Read() has previously
// returned. Note even if Release is not explicitly called, next Read() call
// will still release the current streaming target node.
func (sp *JSONStreamReader) Release(n *Node) {
	if n == sp.stream {
		sp.stream = nil
	}
	RemoveAndReleaseTree(n)
}

// AtLine returns the **rough** line number of the current JSON decoder.
func (sp *JSONStreamReader) AtLine() int {
	return sp.r.AtLine()
}

// NewJSONStreamReader creates a new instance of JSON streaming reader.
func NewJSONStreamReader(r io.Reader, xpathStr string) (*JSONStreamReader, error) {
	xpathStr = strings.TrimSpace(xpathStr)
	xpathNoFilterStr := removeLastFilterInXPath(xpathStr)
	xpathExpr, err := caches.GetXPathExpr(xpathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid xpath '%s', err: %s", xpathStr, err.Error())
	}
	xpathNoFilterExpr, _ := caches.GetXPathExpr(xpathNoFilterStr)
	lineCountingReader := ios.NewLineCountingReader(r)
	reader := &JSONStreamReader{
		r:         lineCountingReader,
		d:         json.NewDecoder(lineCountingReader),
		xpathExpr: xpathNoFilterExpr,
		xpathFilterExpr: func() *xpath.Expr {
			if xpathStr == xpathNoFilterStr {
				return nil
			}
			return xpathExpr
		}(),
		root: CreateJSONNode(DocumentNode, "", JSONRoot),
	}
	reader.cur = reader.root
	return reader, nil
}

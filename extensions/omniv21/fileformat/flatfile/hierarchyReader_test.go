package flatfile

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/antchfx/xpath"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/maths"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

type testRecReader struct {
	moreReturns [][]interface{}
	readReturns [][]interface{}
}

func (r *testRecReader) setMoreReturns(more bool, err error) *testRecReader {
	r.moreReturns = append(r.moreReturns, []interface{}{more, err})
	return r
}

func (r *testRecReader) MoreUnprocessedData() (more bool, err error) {
	if len(r.moreReturns) <= 0 {
		panic("no return values setup for test MoreUnprocessedData call")
	}
	more = r.moreReturns[0][0].(bool)
	if r.moreReturns[0][1] != nil {
		err = r.moreReturns[0][1].(error)
	}
	r.moreReturns = r.moreReturns[1:]
	return
}

func (r *testRecReader) setReadReturns(matched bool, node *idr.Node, err error) *testRecReader {
	r.readReturns = append(r.readReturns, []interface{}{matched, node, err})
	return r
}

func (r *testRecReader) ReadAndMatch(RecDecl, bool) (matched bool, node *idr.Node, err error) {
	if len(r.readReturns) <= 0 {
		panic("no return values setup for test ReadAndMatch call")
	}
	matched = r.readReturns[0][0].(bool)
	node = r.readReturns[0][1].(*idr.Node)
	if r.readReturns[0][2] != nil {
		err = r.readReturns[0][2].(error)
	}
	r.readReturns = r.readReturns[1:]
	return
}

func TestRead(t *testing.T) {
	for _, test := range []struct {
		name        string
		decls       []testDecl
		recReader   *testRecReader
		targetXPath *xpath.Expr
		expReturns  [][]interface{}
	}{
		{
			name:      "RecReader.MoreUnprocessedData failed first call",
			recReader: (&testRecReader{}).setMoreReturns(false, errors.New("test error")),
			expReturns: [][]interface{}{
				{nil, errors.New("test error")},
			},
		},
		{
			name:      "no more data and decl stack already empty - we're done",
			recReader: (&testRecReader{}).setMoreReturns(false, nil),
			expReturns: [][]interface{}{
				{nil, io.EOF},
			},
		},
		{
			name:      "no more data and last decl min occurs not satisfied",
			decls:     []testDecl{{name: "test decl 1", min: 1}},
			recReader: (&testRecReader{}).setMoreReturns(false, nil),
			expReturns: [][]interface{}{
				{nil, errors.New("decl 'test decl 1' requires 1 min occurs but only got 0")},
			},
		},
		{
			name:  "no more data and last decl min occurs satisfied - we're done",
			decls: []testDecl{{name: "test decl 1", min: 0}},
			recReader: (&testRecReader{}).
				setMoreReturns(false, nil).
				setMoreReturns(false, nil),
			expReturns: [][]interface{}{
				{nil, io.EOF},
			},
		},
		{
			name:      "more data but decl stack empty - unexpected data",
			recReader: (&testRecReader{}).setMoreReturns(true, nil),
			expReturns: [][]interface{}{
				{nil, errors.New("unexpected data")},
			},
		},
		{
			name:  "more data but RecReader.ReadAndMatch failed",
			decls: []testDecl{{name: "test decl 1", min: 0}},
			recReader: (&testRecReader{}).
				setMoreReturns(true, nil).
				setReadReturns(false, nil, errors.New("test error")),
			expReturns: [][]interface{}{
				{nil, errors.New("test error")},
			},
		},
		{
			name:  "more data, but not match cur decl so it's done, but its min not satisified",
			decls: []testDecl{{name: "test decl 1", min: 1}},
			recReader: (&testRecReader{}).
				setMoreReturns(true, nil).
				setReadReturns(false, nil, nil),
			expReturns: [][]interface{}{
				{nil, errors.New("decl 'test decl 1' requires 1 min occurs but only got 0")},
			},
		},
		{
			name:  "more data, but not match cur decl so it's done, and its min satisified",
			decls: []testDecl{{name: "test decl 1"}},
			recReader: (&testRecReader{}).
				setMoreReturns(true, nil).
				setReadReturns(false, nil, nil).
				setMoreReturns(false, nil),
			expReturns: [][]interface{}{
				{nil, io.EOF},
			},
		},
		{
			name:  "more data, match cur decl (which is target but has no children)",
			decls: []testDecl{{name: "test decl 1", target: true}},
			recReader: (&testRecReader{}).
				setMoreReturns(true, nil).
				setReadReturns(true, idr.CreateNode(idr.ElementNode, "test 1"), nil).
				setMoreReturns(false, nil),
			expReturns: [][]interface{}{
				{idr.CreateNode(idr.ElementNode, "test 1"), nil},
				{nil, io.EOF},
			},
		},
		{
			name:  "more data, match cur decl (which is target, and has children)",
			decls: []testDecl{{name: "test decl 1", target: true, children: []testDecl{{}}}},
			recReader: (&testRecReader{}).
				setMoreReturns(true, nil).
				setReadReturns(true, idr.CreateNode(idr.ElementNode, "test 1"), nil).
				setMoreReturns(false, nil).
				setMoreReturns(false, nil),
			expReturns: [][]interface{}{
				{idr.CreateNode(idr.ElementNode, "test 1"), nil},
				{nil, io.EOF},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := NewHierarchyReader(toDeclSlice(test.decls), test.recReader, test.targetXPath)
			for _, expRet := range test.expReturns {
				n, err := r.Read()
				if expRet[0] != nil {
					expNode := expRet[0].(*idr.Node)
					assert.NotNil(t, n)
					assert.Equal(t, expNode.Type, n.Type)
					assert.Equal(t, expNode.Data, n.Data)
				} else {
					assert.Nil(t, n)
				}
				if expRet[1] != nil {
					assert.Error(t, err)
					assert.Equal(t, expRet[1].(error).Error(), err.Error())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestRelease(t *testing.T) {
	target := idr.CreateNode(idr.ElementNode, "test")
	r := &HierarchyReader{target: target}
	r.Release(nil)
	assert.Equal(t, target, r.target)
	r.Release(target)
	assert.Nil(t, r.target)
}

func TestReadRec(t *testing.T) {
	for _, test := range []struct {
		name    string
		reader  *testRecReader
		decl    testDecl
		expNode *idr.Node
		expErr  string
	}{
		{
			name:   "no non-group descendent decl",
			reader: nil,
			decl: testDecl{
				group: true,
				children: []testDecl{
					{group: true},
				},
			},
			expNode: nil,
			expErr:  "",
		},
		{
			name:   "RecReader.ReadAndMatch returns error",
			reader: (&testRecReader{}).setReadReturns(false, nil, errors.New("test error")),
			decl: testDecl{
				group:    true,
				children: []testDecl{{}},
			},
			expNode: nil,
			expErr:  "test error",
		},
		{
			name:    "RecReader.ReadAndMatch returns not matched",
			reader:  (&testRecReader{}).setReadReturns(false, nil, nil),
			decl:    testDecl{},
			expNode: nil,
			expErr:  "",
		},
		{
			name:   "RecReader.ReadAndMatch returns group node",
			reader: (&testRecReader{}).setReadReturns(true, nil, nil),
			decl: testDecl{
				name:     "group decl 1",
				group:    true,
				children: []testDecl{{}},
			},
			expNode: idr.CreateNode(idr.ElementNode, "group decl 1"),
			expErr:  "",
		},
		{
			name: "RecReader.ReadAndMatch returns non-group node",
			reader: (&testRecReader{}).setReadReturns(
				true, idr.CreateNode(idr.ElementNode, "non group decl 1"), nil),
			decl: testDecl{
				name: "non group decl 1",
			},
			expNode: idr.CreateNode(idr.ElementNode, "non group decl 1"),
			expErr:  "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := HierarchyReader{r: test.reader}
			n, err := r.readRec(test.decl)
			if test.expNode != nil {
				assert.NotNil(t, n)
				assert.Equal(t, test.expNode.Type, n.Type)
				assert.Equal(t, test.expNode.Data, n.Data)
			} else {
				assert.Nil(t, n)
			}
			if strs.IsStrNonBlank(test.expErr) {
				assert.Error(t, err)
				assert.Equal(t, test.expErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStack(t *testing.T) {
	r := HierarchyReader{
		stack: make([]stackEntry, 0, initialStackDepth),
	}
	// try to access top of stack while there is nothing in it => panic.
	assert.PanicsWithValue(t,
		"frame requested: 0, but stack length: 0",
		func() {
			r.stackTop()
		})
	// try to shrink empty stack => panic.
	assert.PanicsWithValue(t,
		"stack length is empty",
		func() {
			r.shrinkStack()
		})
	newEntry1 := stackEntry{
		recDecl:  testDecl{},
		recNode:  idr.CreateNode(idr.TextNode, "test"),
		curChild: 5,
		occurred: 10,
	}
	r.growStack(newEntry1)
	assert.Equal(t, 1, len(r.stack))
	assert.Equal(t, newEntry1, *r.stackTop())
	newEntry2 := stackEntry{
		recDecl:  testDecl{},
		recNode:  idr.CreateNode(idr.TextNode, "test 2"),
		curChild: 10,
		occurred: 20,
	}
	r.growStack(newEntry2)
	assert.Equal(t, 2, len(r.stack))
	assert.Equal(t, newEntry2, *r.stackTop())
	// try to access a frame that doesn't exist => panic.
	assert.PanicsWithValue(t,
		"frame requested: 2, but stack length: 2",
		func() {
			r.stackTop(2)
		})
	assert.Equal(t, newEntry1, *r.shrinkStack())
	assert.Nil(t, r.shrinkStack())
}

// for tests to create snapshots of stack entries.
func (s stackEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		RecDecl  string
		RecNode  *string
		CurChild int
		Occurred int
	}{
		RecDecl: s.recDecl.DeclName(),
		RecNode: func() *string {
			if s.recNode == nil {
				return nil
			}
			return strs.StrPtr(s.recNode.Data)
		}(),
		CurChild: s.curChild,
		Occurred: s.occurred,
	})
}

func TestRecDoneRecNext(t *testing.T) {
	// the following indicates the test RecDecl tree structure. The numbers in () are min/max.
	//  root (1/1)
	//    A (1/unbounded)   <-- target
	//      B (0/1)
	//      C (1/2)
	//    D (0/1)
	recDeclB := testDecl{name: "B", min: 0, max: 1}
	recDeclC := testDecl{name: "C", min: 1, max: 2}
	recDeclA := testDecl{name: "A", min: 1, max: maths.MaxIntValue, target: true,
		children: []testDecl{recDeclB, recDeclC}}
	recDeclD := testDecl{name: "D", min: 0, max: 1}
	recDeclRoot := testDecl{name: rootName, min: 1, max: 1, group: true,
		children: []testDecl{recDeclA, recDeclD}}

	for _, test := range []struct {
		name        string
		stack       []stackEntry
		target      *idr.Node
		targetXPath *string
		callRecDone bool
		panicStr    string
		err         string
	}{
		{
			name: "root-A-B, B recDone, moves to C, no target",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, idr.CreateNode(idr.ElementNode, "A"), 0, 0},
				{recDeclB, idr.CreateNode(idr.ElementNode, "B"), 0, 0},
			},
			target:      nil,
			targetXPath: nil,
			callRecDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-A-C, C recDone, stay, no target",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{recDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 0},
			},
			target:      nil,
			targetXPath: nil,
			callRecDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-A-C, C recDone, C over max, A becomes target",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{recDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      nil,
			targetXPath: nil,
			callRecDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-D, D recDone",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 1, 0},
				{recDeclD, idr.CreateNode(idr.ElementNode, "D"), 0, 0},
			},
			target:      nil,
			targetXPath: nil,
			callRecDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-A-C, C.occurred = 1, C recNext",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{recDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 0},
			},
			target:      nil,
			targetXPath: nil,
			callRecDone: false,
			panicStr:    "",
			err:         `decl 'C' requires 1 min occurs but only got 0`,
		},
		{
			name: "root-A-C, C recDone, C over max, A becomes target, but r.target not nil",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{recDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      idr.CreateNode(idr.ElementNode, ""),
			targetXPath: nil,
			callRecDone: true,
			panicStr:    `r.target != nil`,
			err:         "",
		},
		{
			name: "root-A-C, C recDone, C over max, A becomes target, but A.recNode is nil",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, nil, 1, 0},
				{recDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      nil,
			targetXPath: nil,
			callRecDone: true,
			panicStr:    `cur.recNode == nil`,
			err:         "",
		},
		{
			name: "root-A-C, C recDone, C over max, A becomes target, but target xpath no match",
			stack: []stackEntry{
				{recDeclRoot, idr.CreateNode(idr.DocumentNode, rootName), 0, 0},
				{recDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{recDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      nil,
			targetXPath: strs.StrPtr("no_match"),
			callRecDone: true,
			panicStr:    "",
			err:         "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &HierarchyReader{
				stack:  test.stack,
				target: test.target,
			}
			if test.targetXPath != nil {
				r.targetXPathExpr = xpath.MustCompile(*test.targetXPath)
			}
			var err error
			testCall := func() {
				if test.callRecDone {
					r.recDone()
				} else {
					err = r.recNext()
				}
			}
			if test.panicStr != "" {
				assert.PanicsWithValue(t, test.panicStr, testCall)
				return
			}
			testCall()
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				return
			}
			assert.NoError(t, err)
			cupaloy.SnapshotT(t, jsons.BPM(
				struct {
					Stack  []stackEntry
					Target *string
				}{
					Stack: r.stack,
					Target: func() *string {
						if r.target == nil {
							return nil
						}
						return strs.StrPtr(r.target.Data)
					}(),
				}))
		})
	}
}

func TestIsErrFewerThanMinOccurs(t *testing.T) {
	assert.True(t, IsErrFewerThanMinOccurs(ErrFewerThanMinOccurs{}))
	assert.Equal(t, "decl 'd' requires 5 min occurs but only got 0",
		ErrFewerThanMinOccurs{RecDecl: testDecl{name: "d", min: 5}}.Error())
	assert.False(t, IsErrFewerThanMinOccurs(errors.New("test")))
}

func TestIsErrUnexpectedData(t *testing.T) {
	assert.True(t, IsErrUnexpectedData(ErrUnexpectedData{}))
	assert.Equal(t, "unexpected data", ErrUnexpectedData{}.Error())
	assert.False(t, IsErrUnexpectedData(errors.New("test")))
}

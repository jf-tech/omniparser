package edi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/jf-tech/go-corelib/testlib"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/idr"
)

func TestIsErrInvalidEDI(t *testing.T) {
	assert.True(t, IsErrInvalidEDI(ErrInvalidEDI("test")))
	assert.Equal(t, "test", ErrInvalidEDI("test").Error())
	assert.False(t, IsErrInvalidEDI(errors.New("test")))
}

func TestRawSeg(t *testing.T) {
	rawSegName := "test"
	rawSegData := []byte("test data")
	r := ediReader{
		unprocessedRawSeg: newRawSeg(),
	}
	assert.False(t, r.unprocessedRawSeg.valid)
	assert.Equal(t, "", r.unprocessedRawSeg.Name)
	assert.Nil(t, r.unprocessedRawSeg.Raw)
	assert.Equal(t, 0, len(r.unprocessedRawSeg.Elems))
	assert.Equal(t, defaultElemsPerSeg, cap(r.unprocessedRawSeg.Elems))
	r.unprocessedRawSeg.valid = true
	r.unprocessedRawSeg.Name = rawSegName
	r.unprocessedRawSeg.Raw = rawSegData
	r.unprocessedRawSeg.Elems = append(
		r.unprocessedRawSeg.Elems, RawSegElem{1, 1, rawSegData[0:4]}, RawSegElem{2, 1, rawSegData[5:]})
	r.resetRawSeg()
	assert.False(t, r.unprocessedRawSeg.valid)
	assert.Equal(t, "", r.unprocessedRawSeg.Name)
	assert.Nil(t, r.unprocessedRawSeg.Raw)
	assert.Equal(t, 0, len(r.unprocessedRawSeg.Elems))
	assert.Equal(t, defaultElemsPerSeg, cap(r.unprocessedRawSeg.Elems))
}

func TestStack(t *testing.T) {
	r := ediReader{
		stack: newStack(),
	}
	assert.Equal(t, 0, len(r.stack))
	assert.Equal(t, defaultStackDepth, cap(r.stack))
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
		segDecl:  &SegDecl{},
		segNode:  idr.CreateNode(idr.TextNode, "test"),
		curChild: 5,
		occurred: 10,
	}
	r.growStack(newEntry1)
	assert.Equal(t, 1, len(r.stack))
	assert.Equal(t, newEntry1, *r.stackTop())
	newEntry2 := stackEntry{
		segDecl:  &SegDecl{},
		segNode:  idr.CreateNode(idr.TextNode, "test 2"),
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

func TestRuneCountAndHasOnlyCRLF(t *testing.T) {
	for _, test := range []struct {
		name             string
		input            []byte
		expectedCount    int
		expectedOnlyCRLF bool
	}{
		{
			name:             "nil",
			input:            nil,
			expectedCount:    0,
			expectedOnlyCRLF: true,
		},
		{
			name:             "empty",
			input:            []byte{},
			expectedCount:    0,
			expectedOnlyCRLF: true,
		},
		{
			name:             "single LF",
			input:            []byte("\n"),
			expectedCount:    1,
			expectedOnlyCRLF: true,
		},
		{
			name:             "single CR",
			input:            []byte("\r"),
			expectedCount:    1,
			expectedOnlyCRLF: true,
		},
		{
			name:             "multiple CR, LF",
			input:            []byte("\r\n\n\r\r"),
			expectedCount:    5,
			expectedOnlyCRLF: true,
		},
		{
			name:             "leading space + LF",
			input:            []byte("   \n"),
			expectedCount:    4,
			expectedOnlyCRLF: false,
		},
		{
			name:             "trailing space + CR",
			input:            []byte("\r   "),
			expectedCount:    4,
			expectedOnlyCRLF: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			count, onlyCRLF := runeCountAndHasOnlyCRLF(test.input)
			assert.Equal(t, test.expectedCount, count)
			assert.Equal(t, test.expectedOnlyCRLF, onlyCRLF)
		})
	}
}

func verifyErr(t *testing.T, expectedErr string, actual error) {
	if expectedErr == "" {
		assert.NoError(t, actual)
		return
	}
	assert.Error(t, actual)
	assert.Equal(t, expectedErr, actual.Error())
}

func TestGetUnprocessedRawSeg(t *testing.T) {
	type result struct {
		rawSeg RawSeg
		err    string
	}
	for _, test := range []struct {
		name     string
		input    io.Reader
		decl     FileDecl
		expected []result
	}{
		{
			name:  "empty input",
			input: strings.NewReader(""),
			decl: FileDecl{
				SegDelim:  "|",
				ElemDelim: ":",
			},
			expected: []result{
				{rawSeg: RawSeg{}, err: io.EOF.Error()},
			},
		},
		{
			name:  "reading error",
			input: testlib.NewMockReadCloser("read failure", nil),
			decl: FileDecl{
				SegDelim:  "|",
				ElemDelim: ":",
			},
			expected: []result{
				{rawSeg: RawSeg{}, err: `input 'test' at segment no.1 (char[1,1]): cannot read segment, err: read failure`},
			},
		},
		{
			name: "CR seg-delim; multi-seg; CRLF only line; LF included; with comp-delim; with release-char",
			input: strings.NewReader(
				"seg1:c00:c01*e10*c20:c21" + "\r\n" +
					"\n" +
					"seg2*c10?*c10:c11*e20?*e20" + "\n"),
			decl: FileDecl{
				SegDelim:    "\n",
				ElemDelim:   "*",
				CompDelim:   strs.StrPtr(":"),
				ReleaseChar: strs.StrPtr("?"),
			},
			expected: []result{
				{
					rawSeg: RawSeg{
						valid: true,
						Name:  "seg1",
						Raw:   []byte("seg1:c00:c01*e10*c20:c21" + "\r\n"),
						Elems: []RawSegElem{
							{ElemIndex: 0, CompIndex: 1, Data: []byte("seg1")},
							{ElemIndex: 0, CompIndex: 2, Data: []byte("c00")},
							{ElemIndex: 0, CompIndex: 3, Data: []byte("c01")},
							{ElemIndex: 1, CompIndex: 1, Data: []byte("e10")},
							{ElemIndex: 2, CompIndex: 1, Data: []byte("c20")},
							{ElemIndex: 2, CompIndex: 2, Data: []byte("c21")},
						},
					},
				},
				{
					rawSeg: RawSeg{
						valid: true,
						Name:  "seg2",
						Raw:   []byte("seg2*c10?*c10:c11*e20?*e20" + "\n"),
						Elems: []RawSegElem{
							{ElemIndex: 0, CompIndex: 1, Data: []byte("seg2")},
							{ElemIndex: 1, CompIndex: 1, Data: []byte("c10?*c10")},
							{ElemIndex: 1, CompIndex: 2, Data: []byte("c11")},
							{ElemIndex: 2, CompIndex: 1, Data: []byte("e20?*e20")},
						},
					},
				},
				{rawSeg: RawSeg{}, err: io.EOF.Error()},
			},
		},
		{
			name:  "missing seg name",
			input: strings.NewReader("|seg2*e3|"),
			decl: FileDecl{
				SegDelim:  "|",
				ElemDelim: "*",
			},
			expected: []result{
				{rawSeg: RawSeg{}, err: `input 'test' at segment no.1 (char[1,2]): missing segment name`},
			},
		},
		{
			name:  "| seg-delim; multi-seg; no comp-delim; no release-char",
			input: strings.NewReader("seg1*e1*e2|seg2*e3|"),
			decl: FileDecl{
				SegDelim:  "|",
				ElemDelim: "*",
			},
			expected: []result{
				{
					rawSeg: RawSeg{
						valid: true,
						Name:  "seg1",
						Raw:   []byte("seg1*e1*e2|"),
						Elems: []RawSegElem{
							{ElemIndex: 0, CompIndex: 1, Data: []byte("seg1")},
							{ElemIndex: 1, CompIndex: 1, Data: []byte("e1")},
							{ElemIndex: 2, CompIndex: 1, Data: []byte("e2")},
						},
					},
				},
				{
					rawSeg: RawSeg{
						valid: true,
						Name:  "seg2",
						Raw:   []byte("seg2*e3|"),
						Elems: []RawSegElem{
							{ElemIndex: 0, CompIndex: 1, Data: []byte("seg2")},
							{ElemIndex: 1, CompIndex: 1, Data: []byte("e3")},
						},
					},
				},
				{rawSeg: RawSeg{}, err: io.EOF.Error()},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader, err := NewReader("test", test.input, &test.decl, "")
			assert.NoError(t, err)
			for {
				if len(test.expected) == 0 {
					assert.FailNow(t, "reader has more content than expected")
				}
				rawSeg, err := reader.getUnprocessedRawSeg()
				verifyErr(t, test.expected[0].err, err)
				assert.Equal(t, test.expected[0].rawSeg, rawSeg)
				// test a second read without resetting returns exactly the same thing.
				if err == nil {
					rawSeg, err = reader.getUnprocessedRawSeg()
					verifyErr(t, test.expected[0].err, err)
					assert.Equal(t, test.expected[0].rawSeg, rawSeg)
				}
				test.expected = test.expected[1:]
				if err != nil {
					break
				}
				reader.resetRawSeg()
			}
			assert.Equal(t, 0, len(test.expected))
		})
	}
}

func TestRawSegToNode(t *testing.T) {
	assert.PanicsWithValue(t, "unprocessedRawSeg is not valid", func() {
		_, _ = (&ediReader{unprocessedRawSeg: RawSeg{valid: false}}).rawSegToNode(nil)
	})

	for _, test := range []struct {
		name     string
		rawSeg   RawSeg
		decl     *SegDecl
		err      string
		expected string
	}{
		{
			name: "element not found",
			rawSeg: RawSeg{
				valid: true,
				Name:  "ISA",
				Raw:   []byte("ISA*0*1:2*3?**"),
				Elems: []RawSegElem{
					{0, 1, []byte("ISA")},
					{1, 1, []byte("0")},
					{2, 1, []byte("1")},
					{2, 2, []byte("2")},
					{3, 1, []byte("3?*")},
				},
			},
			decl: &SegDecl{
				Elems: []Elem{
					{Name: "e1", Index: 1},
					{Name: "e2c1", Index: 2, CompIndex: testlib.IntPtr(1)},
					{Name: "e2c2", Index: 2, CompIndex: testlib.IntPtr(2)},
					{Name: "e3", Index: 4}, // this one doesn't exist
				},
				fqdn: "ISA",
			},
			err:      `input 'test' at segment no.3 (char[10,20]): unable to find element 'e3' on segment 'ISA'`,
			expected: "",
		},
		{
			name: "success",
			rawSeg: RawSeg{
				valid: true,
				Name:  "ISA",
				Raw:   []byte("ISA*0*1:2*3?**"),
				Elems: []RawSegElem{
					{0, 1, []byte("ISA")},
					{1, 1, []byte("0")},
					{2, 1, []byte("1")},
					{2, 2, []byte("2")},
					{3, 1, []byte("3?*")},
				},
			},
			decl: &SegDecl{
				Elems: []Elem{
					{Name: "e1", Index: 1},
					{Name: "e2c1", Index: 2, CompIndex: testlib.IntPtr(1)},
					{Name: "e2c2", Index: 2, CompIndex: testlib.IntPtr(2)},
					{Name: "e3", Index: 3},
					{Name: "e4", Index: 4, EmptyIfMissing: true},
					{Name: "e5", Index: 5, EmptyIfMissing: true},
				},
				fqdn: "ISA",
			},
			err:      "",
			expected: `{"e1":"0","e2c1":"1","e2c2":"2","e3":"3*","e4":"","e5":""}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := ediReader{
				inputName:         "test",
				releaseChar:       newStrPtrByte(strs.StrPtr("?")),
				unprocessedRawSeg: test.rawSeg,
				r:                 &NonValidatingReader{runeBegin: 10, runeEnd: 20, segCount: 3},
			}
			n, err := r.rawSegToNode(test.decl)
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Nil(t, n)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, idr.JSONify2(n))
			}
		})
	}
}

func (s stackEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		SegDecl  string
		SegNode  *string
		CurChild int
		Occurred int
	}{
		SegDecl: s.segDecl.Name,
		SegNode: func() *string {
			if s.segNode == nil {
				return nil
			}
			return strs.StrPtr(s.segNode.Data)
		}(),
		CurChild: s.curChild,
		Occurred: s.occurred,
	})
}

func TestSegDoneSegNext(t *testing.T) {
	// the following is indicates the test SegDecl tree structure. The numbers in () are min/max.
	//  root (1/1)
	//    A (1/-1)   <-- target
	//      B (0/1)
	//      C (1/2)
	//    D (0/1)
	segDeclB := &SegDecl{
		Name: "B",
		Type: strs.StrPtr(segTypeSeg),
		Min:  testlib.IntPtr(0),
	}
	segDeclC := &SegDecl{
		Name: "C",
		Type: strs.StrPtr(segTypeSeg),
		Max:  testlib.IntPtr(2),
	}
	segDeclA := &SegDecl{
		Name:     "A",
		Type:     strs.StrPtr(segTypeSeg),
		IsTarget: true,
		Max:      testlib.IntPtr(-1),
		Children: []*SegDecl{segDeclB, segDeclC},
	}
	segDeclD := &SegDecl{
		Name: "D",
		Type: strs.StrPtr(segTypeSeg),
		Min:  testlib.IntPtr(0),
	}
	segDeclRoot := &SegDecl{
		Name:     rootSegName,
		Type:     strs.StrPtr(segTypeGroup),
		Children: []*SegDecl{segDeclA, segDeclD},
	}
	for _, test := range []struct {
		name        string
		stack       []stackEntry
		target      *idr.Node
		callSegDone bool
		panicStr    string
		err         string
	}{
		{
			name: "root-A-B, B segDone, moves to C, no target",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 0, 0},
				{segDeclA, idr.CreateNode(idr.ElementNode, "A"), 0, 0},
				{segDeclB, idr.CreateNode(idr.ElementNode, "B"), 0, 0},
			},
			target:      nil,
			callSegDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-A-C, C segDone, stay, no target",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 0, 0},
				{segDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{segDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 0},
			},
			target:      nil,
			callSegDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-A-C, C segDone, C over max, A becomes target",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 0, 0},
				{segDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{segDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      nil,
			callSegDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-D, D segDone",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 1, 0},
				{segDeclD, idr.CreateNode(idr.ElementNode, "D"), 0, 0},
			},
			target:      nil,
			callSegDone: true,
			panicStr:    "",
			err:         "",
		},
		{
			name: "root-A-C, C.occurred = 1, C segNext",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 0, 0},
				{segDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{segDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 0},
			},
			target:      nil,
			callSegDone: false,
			panicStr:    "",
			err:         `input 'test' at segment no.3 (char[20,20]): segment 'C' needs min occur 1, but only got 0`,
		},
		{
			name: "root-A-C, C segDone, C over max, A becomes target, but r.target not nil",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 0, 0},
				{segDeclA, idr.CreateNode(idr.ElementNode, "A"), 1, 0},
				{segDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      idr.CreateNode(idr.ElementNode, ""),
			callSegDone: true,
			panicStr:    `r.target != nil`,
			err:         "",
		},
		{
			name: "root-A-C, C segDone, C over max, A becomes target, but A.segNode is nil",
			stack: []stackEntry{
				{segDeclRoot, idr.CreateNode(idr.DocumentNode, rootSegName), 0, 0},
				{segDeclA, nil, 1, 0},
				{segDeclC, idr.CreateNode(idr.ElementNode, "C"), 0, 1},
			},
			target:      nil,
			callSegDone: true,
			panicStr:    `cur.segNode == nil`,
			err:         "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &ediReader{
				inputName: "test",
				stack:     test.stack,
				target:    test.target,
				r:         &NonValidatingReader{runeBegin: 10, runeEnd: 20, segCount: 3},
			}
			var err error
			testCall := func() {
				if test.callSegDone {
					r.segDone()
				} else {
					err = r.segNext()
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

func TestRead(t *testing.T) {
	for _, test := range []struct {
		name              string
		input             string
		declJSON          string
		xpath             string
		readerCreationErr string
	}{
		{
			name:  "invalid target xpath, failure",
			input: "",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{ "name": "ISA", "min": 0 }
					]
				}`,
			xpath:             "[",
			readerCreationErr: `invalid target xpath '[', err: expression must evaluate to a node-set`,
		},
		{
			name:  "empty input, success",
			input: "",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{ "name": "ISA", "min": 0 }
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "single seg decl, multiple seg instances, success",
			input: "ISA*0*1*2\nISA*3\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": -1,
							"elements": [
								{ "name": "e1", "index": 1 },
								{ "name": "e2", "index": 2, "empty_if_missing": true },
								{ "name": "e3", "index": 3, "default": "x" }
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "2 seg decls, success",
			input: "ISA*0*1*2\nISA*3*4*5\nIEA*6\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": -1,
							"elements": [
								{ "name": "e1", "index": 1 },
								{ "name": "e2", "index": 2, "empty_if_missing": true },
								{ "name": "e3", "index": 3, "empty_if_missing": true }
							]
						},
						{
							"name": "IEA",
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "2 seg groups, filtered target, success",
			input: "ISA*0*1*2|\nISA*3*4*5|\nISA*6*7*8|\r\nIEA*6|\n",
			declJSON: `
				{
					"segment_delimiter": "|",
					"element_delimiter": "*",
					"ignore_crlf": true,
					"segment_declarations": [
						{
							"name": "isa_group",
							"type": "segment_group",
							"max": -1,
							"child_segments": [
								{
									"name": "ISA",
									"is_target": true,
									"elements": [
										{ "name": "e1", "index": 1 },
										{ "name": "e2", "index": 2, "empty_if_missing": true },
										{ "name": "e3", "index": 3, "empty_if_missing": true }
									]
								}
							]
						},
						{
							"name": "iea_group",
							"type": "segment_group",
							"child_segments": [
								{
									"name": "IEA",
									"elements": [
										{ "name": "e1", "index": 1 }
									]
								}
							]
						}
					]
				}`,
			xpath:             ".[e1 != '6']",
			readerCreationErr: "",
		},
		{
			name:  "seg min not satisfied before EOF, failure",
			input: "ISA*0*1*2\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": -1,
							"elements": [
								{ "name": "e1", "index": 1 },
								{ "name": "e2", "index": 2, "empty_if_missing": true },
								{ "name": "e3", "index": 3, "empty_if_missing": true }
							]
						},
						{
							"name": "IEA",
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "missing raw seg name, failure",
			input: "*0*1\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": -1,
							"elements": [
								{ "name": "e1", "index": 1 },
								{ "name": "e2", "index": 2 }
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "raw seg processing wrong, failure",
			input: "ISA*0\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"max": -1,
							"elements": [
								{ "name": "e1", "index": 1 },
								{ "name": "e2", "index": 2 }
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "seg min not satisfied before next seg appearance, failure",
			input: "IEA*0\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						},
						{
							"name": "IEA",
							"elements": [
								{ "name": "e1", "index": 1 }
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "unprocessed raw segment, failure",
			input: "ISA\nUNKNOWN\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "ISA",
							"is_target": true,
                            "min": 0
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
		{
			name:  "multiple root level segments, success",
			input: "ISA\nIEA\nISA\nIEA\n",
			declJSON: `
				{
					"segment_delimiter": "\n",
					"element_delimiter": "*",
					"segment_declarations": [
						{
							"name": "group1",
							"type": "segment_group",
							"is_target": true,
							"child_segments": [
								{
									"name": "ISA"
								},
								{
									"name": "IEA"
								}
							]
						}
					]
				}`,
			xpath:             "",
			readerCreationErr: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var decl FileDecl
			err := json.Unmarshal([]byte(test.declJSON), &decl)
			assert.NoError(t, err)
			reader, err := NewReader("test", strings.NewReader(test.input), &decl, test.xpath)
			if test.readerCreationErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.readerCreationErr, err.Error())
				assert.Nil(t, reader)
				return
			}
			assert.NoError(t, err)
			var records []string
			var finalErr error
			for {
				n, err := reader.Read()
				if err != nil {
					finalErr = err
					break
				}
				records = append(records, strings.ReplaceAll(idr.JSONify2(n), `"`, `'`))
			}
			cupaloy.SnapshotT(t, jsons.BPM(
				struct {
					Records  []string
					FinalErr string
				}{
					Records: records,
					FinalErr: func() string {
						if finalErr == nil {
							return ""
						}
						return finalErr.Error()
					}(),
				}))
		})
	}
}

func TestRelease(t *testing.T) {
	var decl FileDecl
	err := json.Unmarshal([]byte(`
		{
			"segment_delimiter": "\n",
			"element_delimiter": "*",
			"segment_declarations": [
				{
					"name": "ISA",
					"is_target": true,
					"max": -1,
					"elements": [
						{ "name": "e1", "index": 1 }
					]
				}
			]
		}`), &decl)
	assert.NoError(t, err)
	reader, err := NewReader("test", strings.NewReader("ISA*0*1*2\n"), &decl, "")
	assert.NoError(t, err)
	n, err := reader.Read()
	assert.NoError(t, err)
	assert.True(t, n == reader.target)
	reader.Release(n)
	assert.Nil(t, reader.target)
}

func TestIsContinuableError(t *testing.T) {
	r := &ediReader{r: &NonValidatingReader{}}
	assert.True(t, r.IsContinuableError(r.FmtErr("some error")))
	assert.False(t, r.IsContinuableError(ErrInvalidEDI("invalid EDI")))
	assert.False(t, r.IsContinuableError(io.EOF))
}

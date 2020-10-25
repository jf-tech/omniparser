package edi

import (
	"errors"
	"io"
	"strings"
	"testing"

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
	assert.Equal(t, "", r.unprocessedRawSeg.name)
	assert.Nil(t, r.unprocessedRawSeg.raw)
	assert.Equal(t, 0, len(r.unprocessedRawSeg.elems))
	assert.Equal(t, defaultElemsPerSeg, cap(r.unprocessedRawSeg.elems))
	r.unprocessedRawSeg.valid = true
	r.unprocessedRawSeg.name = rawSegName
	r.unprocessedRawSeg.raw = rawSegData
	r.unprocessedRawSeg.elems = append(
		r.unprocessedRawSeg.elems, rawSegElem{1, 1, rawSegData[0:4], false}, rawSegElem{2, 1, rawSegData[5:], false})
	r.resetRawSeg()
	assert.False(t, r.unprocessedRawSeg.valid)
	assert.Equal(t, "", r.unprocessedRawSeg.name)
	assert.Nil(t, r.unprocessedRawSeg.raw)
	assert.Equal(t, 0, len(r.unprocessedRawSeg.elems))
	assert.Equal(t, defaultElemsPerSeg, cap(r.unprocessedRawSeg.elems))
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
		segDecl:  &segDecl{},
		segNode:  idr.CreateNode(idr.TextNode, "test"),
		curChild: 5,
		occurred: 10,
	}
	r.growStack(newEntry1)
	assert.Equal(t, 1, len(r.stack))
	assert.Equal(t, newEntry1, *r.stackTop())
	newEntry2 := stackEntry{
		segDecl:  &segDecl{},
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
		rawSeg rawSeg
		err    string
	}
	for _, test := range []struct {
		name     string
		input    io.Reader
		decl     fileDecl
		expected []result
	}{
		{
			name:  "empty input",
			input: strings.NewReader(""),
			decl: fileDecl{
				SegDelim:  "|",
				ElemDelim: ":",
			},
			expected: []result{
				{rawSeg: rawSeg{}, err: io.EOF.Error()},
			},
		},
		{
			name:  "reading error",
			input: testlib.NewMockReadCloser("read failure", nil),
			decl: fileDecl{
				SegDelim:  "|",
				ElemDelim: ":",
			},
			expected: []result{
				{rawSeg: rawSeg{}, err: `input 'test' between character [1,1]: cannot read segment, err: read failure`},
			},
		},
		{
			name: "CR seg-delim; multi-seg; CRLF only line; LF included; with comp-delim; with release-char",
			input: strings.NewReader(
				"seg1:c00:c01*e10*c20:c21" + "\r\n" +
					"\n" +
					"seg2*c10?*c10:c11*e20?*e20" + "\n"),
			decl: fileDecl{
				SegDelim:    "\n",
				ElemDelim:   "*",
				CompDelim:   strs.StrPtr(":"),
				ReleaseChar: strs.StrPtr("?"),
			},
			expected: []result{
				{
					rawSeg: rawSeg{
						valid: true,
						name:  "seg1",
						raw:   []byte("seg1:c00:c01*e10*c20:c21" + "\r\n"),
						elems: []rawSegElem{
							{elemIndex: 0, compIndex: 1, data: []byte("seg1")},
							{elemIndex: 0, compIndex: 2, data: []byte("c00")},
							{elemIndex: 0, compIndex: 3, data: []byte("c01")},
							{elemIndex: 1, compIndex: 1, data: []byte("e10")},
							{elemIndex: 2, compIndex: 1, data: []byte("c20")},
							{elemIndex: 2, compIndex: 2, data: []byte("c21")},
						},
					},
				},
				{
					rawSeg: rawSeg{
						valid: true,
						name:  "seg2",
						raw:   []byte("seg2*c10?*c10:c11*e20?*e20" + "\n"),
						elems: []rawSegElem{
							{elemIndex: 0, compIndex: 1, data: []byte("seg2")},
							{elemIndex: 1, compIndex: 1, data: []byte("c10?*c10")},
							{elemIndex: 1, compIndex: 2, data: []byte("c11")},
							{elemIndex: 2, compIndex: 1, data: []byte("e20?*e20")},
						},
					},
				},
				{rawSeg: rawSeg{}, err: io.EOF.Error()},
			},
		},
		{
			name:  "missing seg name",
			input: strings.NewReader("|seg2*e3|"),
			decl: fileDecl{
				SegDelim:  "|",
				ElemDelim: "*",
			},
			expected: []result{
				{rawSeg: rawSeg{}, err: `input 'test' between character [1,2]: segment is malformed, missing segment name`},
			},
		},
		{
			name:  "| seg-delim; multi-seg; no comp-delim; no release-char",
			input: strings.NewReader("seg1*e1*e2|seg2*e3|"),
			decl: fileDecl{
				SegDelim:  "|",
				ElemDelim: "*",
			},
			expected: []result{
				{
					rawSeg: rawSeg{
						valid: true,
						name:  "seg1",
						raw:   []byte("seg1*e1*e2|"),
						elems: []rawSegElem{
							{elemIndex: 0, compIndex: 1, data: []byte("seg1")},
							{elemIndex: 1, compIndex: 1, data: []byte("e1")},
							{elemIndex: 2, compIndex: 1, data: []byte("e2")},
						},
					},
				},
				{
					rawSeg: rawSeg{
						valid: true,
						name:  "seg2",
						raw:   []byte("seg2*e3|"),
						elems: []rawSegElem{
							{elemIndex: 0, compIndex: 1, data: []byte("seg2")},
							{elemIndex: 1, compIndex: 1, data: []byte("e3")},
						},
					},
				},
				{rawSeg: rawSeg{}, err: io.EOF.Error()},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader := NewReader("test", test.input, &test.decl)
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
		_, _ = (&ediReader{unprocessedRawSeg: rawSeg{valid: false}}).rawSegToNode(nil)
	})

	for _, test := range []struct {
		name     string
		rawSeg   rawSeg
		decl     *segDecl
		err      string
		expected string
	}{
		{
			name: "fixed_length_in_bytes wrong",
			rawSeg: rawSeg{
				valid: true,
				name:  "ISA",
				raw:   []byte("0123456789"),
			},
			decl: &segDecl{
				FixedLengthInBytes: testlib.IntPtr(11),
				fqdn:               "ISA",
			},
			err:      `input 'test' between character [10,20]: segment 'ISA' expected length 11 byte(s), but got: 10 byte(s)`,
			expected: "",
		},
		{
			name: "element not found",
			rawSeg: rawSeg{
				valid: true,
				name:  "ISA",
				raw:   []byte("ISA*0*1:2*3?**"),
				elems: []rawSegElem{
					{0, 1, []byte("ISA"), false},
					{1, 1, []byte("0"), false},
					{2, 1, []byte("1"), false},
					{2, 2, []byte("2"), false},
					{3, 1, []byte("3?*"), false},
				},
			},
			decl: &segDecl{
				Elems: []elem{
					{Name: "e1", Index: 1},
					{Name: "e2c1", Index: 2, CompIndex: testlib.IntPtr(1)},
					{Name: "e2c2", Index: 2, CompIndex: testlib.IntPtr(2)},
					{Name: "e3", Index: 4}, // this one doesn't exist
				},
				fqdn: "ISA",
			},
			err:      `input 'test' between character [10,20]: unable to find element 'e3' on segment 'ISA'`,
			expected: "",
		},
		{
			name: "success",
			rawSeg: rawSeg{
				valid: true,
				name:  "ISA",
				raw:   []byte("ISA*0*1:2*3?**"),
				elems: []rawSegElem{
					{0, 1, []byte("ISA"), false},
					{1, 1, []byte("0"), false},
					{2, 1, []byte("1"), false},
					{2, 2, []byte("2"), false},
					{3, 1, []byte("3?*"), false},
				},
			},
			decl: &segDecl{
				Elems: []elem{
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
				runeBegin:         10,
				runeEnd:           20,
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

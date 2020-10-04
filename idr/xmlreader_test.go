package idr

import (
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestXMLStreamReader_InvalidXPath(t *testing.T) {
	sp, err := NewXMLStreamReader(strings.NewReader(""), "[invalid")
	assert.Error(t, err)
	assert.Equal(t, "invalid xpath '[invalid', err: expression must evaluate to a node-set", err.Error())
	assert.Nil(t, sp)
}

func TestXMLStreamReader_SuccessWithXPathWithFilter(t *testing.T) {
	s := `
	<ROOT xmlns="uri://default" xmlns:t="uri://test">
		<AAA>
			<CCC>c1</CCC>
			<t:BBB>b1</t:BBB>
			<DDD>d1</DDD>
			<t:BBB>b2<ZZZ z="1">z1</ZZZ></t:BBB>
			<t:BBB>b3</t:BBB>
		</AAA>
		<ZZZ>
			<t:BBB>b4</t:BBB>
			<CCC>c3</CCC>
		</ZZZ>
	</ROOT>`

	sp, err := NewXMLStreamReader(strings.NewReader(s), "/ROOT/*/t:BBB[. != 'b3']")
	assert.NoError(t, err)
	assert.Equal(t, 1, sp.AtLine())

	// First `<t:BBB>` read
	n, err := sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b1", n.InnerText())
	t.Run("IDR snapshot after 1st Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 5, sp.AtLine())

	// Second `<t:BBB>` read
	n, err = sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b2z1", n.InnerText())
	t.Run("IDR snapshot after 2nd Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 7, sp.AtLine())

	// Third `<t:BBB>` read (Note we will skip 'b3' since the streamElementFilter excludes it)
	n, err = sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b4", n.InnerText())
	t.Run("IDR snapshot after 3rd Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 11, sp.AtLine())

	n, err = sp.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestXMLStreamReader_SuccessWithXPathWithoutFilter(t *testing.T) {
	s := `
	<ROOT xmlns="uri://default" xmlns:t="uri://test">
		<AAA>
			<CCC>c1</CCC>
			<t:BBB>b1</t:BBB>
			<DDD>d1</DDD>
			<t:BBB>b2<ZZZ z="1">z1</ZZZ></t:BBB>
			<t:BBB>b3</t:BBB>
		</AAA>
		<ZZZ>
			<t:BBB>b4</t:BBB>
			<CCC>c3</CCC>
		</ZZZ>
	</ROOT>`

	sp, err := NewXMLStreamReader(strings.NewReader(s), "/ROOT/*/t:BBB")
	assert.NoError(t, err)
	assert.Equal(t, 1, sp.AtLine())

	// First `<t:BBB>` read
	n, err := sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b1", n.InnerText())
	t.Run("IDR snapshot after 1st Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 5, sp.AtLine())

	// Second `<t:BBB>` read
	n, err = sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b2z1", n.InnerText())
	t.Run("IDR snapshot after 2nd Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 7, sp.AtLine())

	// Third `<t:BBB>` read
	n, err = sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b3", n.InnerText())
	t.Run("IDR snapshot after 3rd Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 8, sp.AtLine())

	// Fourth `<t:BBB>` read
	n, err = sp.Read()
	assert.NoError(t, err)
	assert.Equal(t, "b4", n.InnerText())
	t.Run("IDR snapshot after 4th Read", func(t *testing.T) {
		cupaloy.SnapshotT(t, JSONify1(rootOf(n)))
	})
	assert.Equal(t, 11, sp.AtLine())

	n, err = sp.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestXMLStreamReader_FailureInvalidElementNodeNamespacePrefix(t *testing.T) {
	s := `<ROOT><non_existing:AAA/></ROOT>`
	sp, err := NewXMLStreamReader(strings.NewReader(s), "/ROOT/non_existing:AAA")
	assert.NoError(t, err)
	n, err := sp.Read()
	assert.Error(t, err)
	assert.Equal(t, "unknown namespace 'non_existing' on ElementNode 'AAA'", err.Error())
	assert.Nil(t, n)
	// repeat Read() again to show/verify that stream reader will stop processing
	// once fatal error is detected.
	n, err = sp.Read()
	assert.Error(t, err)
	assert.Equal(t, "unknown namespace 'non_existing' on ElementNode 'AAA'", err.Error())
	assert.Nil(t, n)
}

func TestXMLStreamReader_FailureInvalidAttributeNodeNamespacePrefix(t *testing.T) {
	s := `<ROOT xmlns:valid="uri://valid-namespace"><AAA non_existing:attr="test" /></ROOT>`
	sp, err := NewXMLStreamReader(strings.NewReader(s), "/ROOT/AAA")
	assert.NoError(t, err)
	n, err := sp.Read()
	assert.Error(t, err)
	assert.Equal(t, "unknown namespace 'non_existing' on AttributeNode 'attr'", err.Error())
	assert.Nil(t, n)
	// repeat Read() again to show/verify that stream reader will stop processing
	// once fatal error is detected.
	n, err = sp.Read()
	assert.Error(t, err)
	assert.Equal(t, "unknown namespace 'non_existing' on AttributeNode 'attr'", err.Error())
	assert.Nil(t, n)
}

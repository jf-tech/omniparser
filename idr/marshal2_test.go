package idr

import (
	"io"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/stretchr/testify/assert"
)

func TestJSONify2XML(t *testing.T) {
	xml := `
		<a:root xmlns="uri-default" xmlns:a="uri-a">
			<a:child a:attr1="a1" a:attr2="a2">
				<a:grandchild>g1</a:grandchild>
			</a:child>
			<a:child a:attr1="a3" a:attr2="a4">
				<a:grandchild>g2</a:grandchild>
			</a:child>
			<child attr1="a5" attr2="a6">
				<grandchild>g3</grandchild>
				<grandchild>g4</grandchild>
				<grandchild>
					<greatgrandchild>gg1</greatgrandchild>
				</grandchild>
			</child>
			<a:child a:attr1="a7" a:attr2="a8">
				<a:grandchild>g5</a:grandchild>
			</a:child>
		</a:root>`
	sp, err := NewXMLStreamReader(strings.NewReader(xml), "/")
	assert.NoError(t, err)
	n, err := sp.Read()
	assert.NoError(t, err)
	cupaloy.SnapshotT(t, jsons.BPJ(JSONify2(n)))
	n, err = sp.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

func TestJSONify2JSON(t *testing.T) {
	j := `
		{
			"a": [],
			"b": 3.1415,
			"c": false,
			"d": null,
			"e": [
				1,
				"two",
				true
			],
			"f": {}
		}`
	sp, err := NewJSONStreamReader(strings.NewReader(j), "/")
	assert.NoError(t, err)
	n, err := sp.Read()
	assert.NoError(t, err)
	cupaloy.SnapshotT(t, jsons.BPJ(JSONify2(n)))
	n, err = sp.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, n)
}

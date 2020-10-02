package idr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsXML(t *testing.T) {
	assert.True(t, IsXML(CreateXMLNode(DocumentNode, "", XMLSpecific{})))
	assert.True(t, IsXML(CreateXMLNode(ElementNode, "A", XMLSpecific{NamespacePrefix: "ns", NamespaceURI: "uri://"})))
	assert.True(t, IsXML(CreateXMLNode(TextNode, "text", XMLSpecific{})))
	assert.True(t, IsXML(CreateXMLNode(AttributeNode, "A", XMLSpecific{})))
	assert.False(t, IsXML(CreateNode(ElementNode, "B")))
}

func TestXMLSpecificOf(t *testing.T) {
	assert.Equal(t, XMLSpecific{}, XMLSpecificOf(CreateXMLNode(ElementNode, "A", XMLSpecific{})))
	assert.Equal(t,
		XMLSpecific{NamespacePrefix: "ns", NamespaceURI: "uri"},
		XMLSpecificOf(
			CreateXMLNode(ElementNode, "A", XMLSpecific{NamespacePrefix: "ns", NamespaceURI: "uri"})))
	assert.PanicsWithValue(t, "node is not XML", func() {
		XMLSpecificOf(CreateNode(ElementNode, "A"))
	})
}

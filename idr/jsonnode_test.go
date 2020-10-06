package idr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJSON(t *testing.T) {
	assert.True(t, IsJSON(CreateJSONNode(DocumentNode, "", JSONRoot)))
	assert.True(t, IsJSON(CreateJSONNode(ElementNode, "a", JSONProp|JSONObj)))
	assert.True(t, IsJSON(CreateJSONNode(TextNode, "text", JSONValue)))
	assert.False(t, IsJSON(CreateNode(ElementNode, "a")))
	assert.False(t, IsJSON(CreateXMLNode(AttributeNode, "a", XMLSpecific{})))
}

func TestJSONTypeOf(t *testing.T) {
	assert.Equal(t, JSONRoot, JSONTypeOf(CreateJSONNode(DocumentNode, "", JSONRoot)))
	assert.Equal(t, JSONProp|JSONObj, JSONTypeOf(CreateJSONNode(ElementNode, "a", JSONProp|JSONObj)))
	assert.PanicsWithValue(t, "node is not JSON", func() {
		JSONTypeOf(CreateNode(ElementNode, "a"))
	})
}

func TestIsJSONType(t *testing.T) {
	assert.True(t, IsJSONRoot(CreateJSONNode(DocumentNode, "", JSONObj|JSONRoot)))
	assert.False(t, IsJSONRoot(CreateJSONNode(ElementNode, "a", JSONProp|JSONObj)))
	assert.False(t, IsJSONRoot(CreateXMLNode(AttributeNode, "a", XMLSpecific{})))

	assert.True(t, IsJSONObj(CreateJSONNode(DocumentNode, "", JSONObj|JSONRoot)))
	assert.False(t, IsJSONObj(CreateJSONNode(ElementNode, "a", JSONProp|JSONArr)))
	assert.False(t, IsJSONObj(CreateXMLNode(AttributeNode, "a", XMLSpecific{})))

	assert.True(t, IsJSONArr(CreateJSONNode(ElementNode, "a", JSONProp|JSONArr)))
	assert.False(t, IsJSONArr(CreateJSONNode(TextNode, "a", JSONValue)))
	assert.False(t, IsJSONArr(CreateXMLNode(TextNode, "a", XMLSpecific{})))

	assert.True(t, IsJSONProp(CreateJSONNode(ElementNode, "a", JSONProp|JSONArr)))
	assert.False(t, IsJSONProp(CreateJSONNode(DocumentNode, "", JSONRoot|JSONObj)))
	assert.False(t, IsJSONProp(CreateXMLNode(ElementNode, "a", XMLSpecific{})))

	assert.True(t, IsJSONValue(CreateJSONNode(DocumentNode, "a", JSONRoot|JSONValue)))
	assert.False(t, IsJSONValue(CreateJSONNode(ElementNode, "a", JSONProp|JSONObj)))
	assert.False(t, IsJSONValue(CreateXMLNode(ElementNode, "a", XMLSpecific{})))
}

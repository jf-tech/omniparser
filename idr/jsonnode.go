package idr

import (
	"fmt"
)

// JSONType is the type of JSON-specific Node.
// Note multiple JSONType can be bit-wise OR'ed together.
type JSONType uint

const (
	// JSONRoot is the type for the root Node in a JSON IDR tree.
	JSONRoot JSONType = 1 << iota
	// JSONObj is the type for a Node in a JSON IDR tree whose value is an object.
	JSONObj
	// JSONArr is the type for a Node in a JSON IDR tree whose value is an array.
	JSONArr
	// JSONProp is the type for a Node in a JSON IDR tree who is a property.
	JSONProp
	// JSONValueStr is the type for a Node in a JSON IDR tree who is a string value.
	JSONValueStr
	// JSONValueNum is the type for a Node in a JSON IDR tree who is a numeric value.
	JSONValueNum
	// JSONValueBool is the type for a Node in a JSON IDR tree who is a boolean value.
	JSONValueBool
	// JSONValueNull is the type for a Node in a JSON IDR tree who is a null value.
	JSONValueNull

	jsonTypeEnd
)

// String converts JSONType to a string.
func (jt JSONType) String() string {
	switch jt {
	case JSONRoot:
		return "JSONRoot"
	case JSONObj:
		return "JSONObj"
	case JSONArr:
		return "JSONArr"
	case JSONProp:
		return "JSONProp"
	case JSONValueStr:
		return "JSONValueStr"
	case JSONValueNum:
		return "JSONValueNum"
	case JSONValueBool:
		return "JSONValueBool"
	case JSONValueNull:
		return "JSONValueNull"
	default:
		return fmt.Sprintf("(unknown JSONType: %d)", jt)
	}
}

// IsJSON checks if a Node is of JSON.
func IsJSON(n *Node) bool {
	_, ok := n.FormatSpecific.(JSONType)
	return ok
}

// JSONTypeOf returns the JSONType of a Node.
// Note if the Node isn't of JSON, this function will panic.
func JSONTypeOf(n *Node) JSONType {
	if !IsJSON(n) {
		panic("node is not JSON")
	}
	return n.FormatSpecific.(JSONType)
}

// IsJSONRoot checks if a given Node is of JSONRoot type.
func IsJSONRoot(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONRoot != 0 }

// IsJSONObj checks if a given Node is of JSONObj type.
func IsJSONObj(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONObj != 0 }

// IsJSONArr checks if a given Node is of JSONArr type.
func IsJSONArr(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONArr != 0 }

// IsJSONProp checks if a given Node is of JSONProp type.
func IsJSONProp(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONProp != 0 }

// IsJSONValueStr checks if a given Node is of JSONValueStr type.
func IsJSONValueStr(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONValueStr != 0 }

// IsJSONValueNum checks if a given Node is of JSONValueNum type.
func IsJSONValueNum(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONValueNum != 0 }

// IsJSONValueBool checks if a given Node is of JSONValueBool type.
func IsJSONValueBool(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONValueBool != 0 }

// IsJSONValueNull checks if a given Node is of JSONValueNull type.
func IsJSONValueNull(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONValueNull != 0 }

// IsJSONValue checks if a given Node is of any JSON value types.
func IsJSONValue(n *Node) bool {
	return IsJSONValueStr(n) || IsJSONValueNum(n) || IsJSONValueBool(n) || IsJSONValueNull(n)
}

// CreateJSONNode creates a JSON Node.
func CreateJSONNode(ntype NodeType, data string, jtype JSONType) *Node {
	n := CreateNode(ntype, data)
	n.FormatSpecific = jtype
	return n
}

package idr

type JSONType uint

const (
	JSONRoot JSONType = 1 << iota
	JSONObj
	JSONArr
	JSONProp
	JSONValue
)

func IsJSON(n *Node) bool {
	_, ok := n.FormatSpecific.(JSONType)
	return ok
}

func JSONTypeOf(n *Node) JSONType {
	if !IsJSON(n) {
		panic("node is not JSON")
	}
	return n.FormatSpecific.(JSONType)
}

func IsJSONRoot(n *Node) bool  { return IsJSON(n) && JSONTypeOf(n)&JSONRoot != 0 }
func IsJSONObj(n *Node) bool   { return IsJSON(n) && JSONTypeOf(n)&JSONObj != 0 }
func IsJSONArr(n *Node) bool   { return IsJSON(n) && JSONTypeOf(n)&JSONArr != 0 }
func IsJSONProp(n *Node) bool  { return IsJSON(n) && JSONTypeOf(n)&JSONProp != 0 }
func IsJSONValue(n *Node) bool { return IsJSON(n) && JSONTypeOf(n)&JSONValue != 0 }

func CreateJSONNode(ntype NodeType, data string, jtype JSONType) *Node {
	n := CreateNode(ntype, data)
	n.FormatSpecific = jtype
	return n
}

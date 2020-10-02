package idr

// XMLSpecific contains XML IDR Node specific information such as namespace.
type XMLSpecific struct {
	NamespacePrefix string
	NamespaceURI    string
}

// IsXML checks if a Node is of XML.
func IsXML(n *Node) bool {
	_, ok := n.FormatSpecific.(XMLSpecific)
	return ok
}

// XMLSpecificOf returns the XMLSpecific field of a Node.
// Note if the Node isn't of XML, this function will panic.
func XMLSpecificOf(n *Node) XMLSpecific {
	if !IsXML(n) {
		panic("node is not XML")
	}
	return n.FormatSpecific.(XMLSpecific)
}

// CreateXMLNode creates an XML Node.
func CreateXMLNode(ntype NodeType, data string, xmlSpecific XMLSpecific) *Node {
	n := CreateNode(ntype, data)
	n.FormatSpecific = xmlSpecific
	return n
}

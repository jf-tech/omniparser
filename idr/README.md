# IDR

**IDR** == **I**ntermediate **D**ata **R**epresentation or **I**n-memory **D**ata **R**epresentation

IDR is an intermediate data structure used by omniparser ingesters to store raw data read from various
formats of inputs, including CSV/txt/XML/EDI/JSON/etc, and then used by schema handlers to perform
transforms. It is flexible and versatile to represent all kinds of data formats supported (or to be
supported) by omniparser.

*Credit:* The basic data structures and various operations and algorithms used by IDR are mostly
inherited/adapted from, modified based on, and inspired by works done in https://github.com/antchfx/xmlquery
and https://github.com/antchfx/xpath. Thank you very much!

The basic building block of an IDR is a `Node` and an IDR is in fact a `Node` tree. Each `Node` has
two parts (see actual code [here](./node.go)):
```
type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type NodeType
	Data string

	FormatSpecific interface{}
}
```
The first part of a `Node` contains the input format agnostic fields, such as tree pointers (like
`Parent`, `FirstChild`, etc), `Type` and `Data`, which we'll explain more in details later. The second
part of a `Node` is format specific data blob. The blob not only offers a place to store format specific
data it also gives IDR code and algorithms a hint on what input format the `Node` is about.

Below we'll go through each input format we support and show how its corresponding IDR looks like.

## XML

Since XML is the most complex input format we'll deal with by IDR. Let's cover it first.

Let's take a look a simple example of XML (from [this sample](../samples/omniv2/xml/1_datetime_parse_and_format.input.xml)):
```
<Root>
    <JustDate>2020/09/22</JustDate>
    <DateTimeWithNoTZ>09/22/2020 12:34:56</DateTimeWithNoTZ>
</Root>
```
This is a simple XML blob with no non-default namespaces and with no attributes. Its corresponding IDR
looks like this (with empty field omitted and tree pointers omitted for clarity):
```
Node(Type: DocumentNode, FormatSpecific: XMLSpecific())
    Node(Type: ElementNode, Data: "Root", FormatSpecific: XMLSpecific())
        Node(Type: TextNode, Data: "\n", FormatSpecific: XMLSpecific())
        Node(Type: ElementNode, Data: "JustData", FormatSpecific: XMLSpecific())
            Node(Type: TextNode, Data: "2020/09/22", FormatSpecific: XMLSpecific())
        Node(Type: TextNode, Data: "\n", FormatSpecific: XMLSpecific())
        Node(Type: ElementNode, Data: "DateTimeWithNoTZ", FormatSpecific: XMLSpecific())
            Node(Type: TextNode, Data: "09/22/2020 12:34:56", FormatSpecific: XMLSpecific())
        Node(Type: TextNode, Data: "\n", FormatSpecific: XMLSpecific())
```
Most of the IDR is quite self-explanatory, but what about those `TextNode`'s with `\n` as `Data`? Turns
out [`xml.Decoder`](https://golang.org/pkg/encoding/xml/#Decoder) treats anything in between two XML
element nodes as text, as long as the two elements are not directly adjacent to each other. Since
there is a newline `'\n'` after the XML element `<Root>` and before `<JustDate>`, the `'\n'` is captured
as a `TextNode`.

Also note in this simple case, each of the `Node` has an empty but none-nil `FormatSpecific`, typed as
[`XMLSpecific`](./xmlnode.go). `XMLSpecific` contains XML namespace information for each of the node,
which we'll see in the [next example](../samples/omniv2/xml/2_multiple_objects.input.xml):
```
<lb0:library xmlns:lb0="uri://something">
    <lb0:books>
        <book title="Harry Potter and the Philosopher's Stone">
            <author>J. K. Rowling</author>
        </book>
    </lb0:books>
</lb0:library>
```
In this example, we'll see how IDR deals with XML namespaces, as well as attributes.

The IDR represents the example above looks like the following (note those "dummy" text nodes sprinkled
in between element nodes are omitted here for clarity; also not including empty `XMLSpecific`):
```
Node(Type: DocumentNode)
    Node(Type: ElementNode, Data: "library", FormatSpecific: XMLSpecific(NamespacePrefix: "lb0", NamespaceURI: "uri://something"))
        Node(Type: ElementNode, Data: "books", FormatSpecific: XMLSpecific(NamespacePrefix: "lb0", NamespaceURI: "uri://something"))
            Node(Type: ElementNode, Data: "book")
                Node(Type: AttributeNode, Data: "title")
                    Node(Type: TextNode, Data: "Harry Potter and the Philosopher's Stone")
                Node(Type: ElementNode, Data: "author")
                    Node(Type: TextNode, Data: "J. K. Rowling")
```
Both `Node`'s represent `<lb0:library>` and `<lb0:books>` include non-empty `XMLSpecific`'s which
contain their namespace prefixes and full URIs while their `Node.Data` contains the element name without
the namespace prefixes.

Note XML attributes on elements are represented as `Node`'s as well, `Type: AttributeNode` specifically.
If an attribute is namespace prefixed, the `AttributeNode` typed `Node` will have a non-empty
`XMLSpecific` set as well. An attribute's value is placed as a `TextNode` underneath its `ElementNode`.
`AttributeNode`'s are guaranteed to be placed before any other child nodes (`TextNode`, or `ElementNode`)
by IDR's XML reader.

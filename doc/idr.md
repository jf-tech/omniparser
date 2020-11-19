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
	ID int64
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node
	Type NodeType
	Data string

	FormatSpecific interface{}
}
```
The first part of a `Node` contains the input format agnostic fields, such as `ID`, tree pointers (like
`Parent`, `FirstChild`, etc), `Type` and `Data`, which we'll explain more in details later. The second
part of a `Node` is format specific data blob. The blob not only offers a place to store format specific
data it also gives IDR code and algorithms a hint on what input format the `Node` is about.

We'll go through each input format we support and show what its corresponding IDR looks like. But before
we dive into format specific IDR, let's take a look what `ID` is about.

## Node.ID
During ingester reading and processing, tons of `Node` will be allocated on heap causing lots of GCs. We
introduced `Node` caching and recycling in this [commit](https://github.com/jf-tech/omniparser/commit/9c55da752971f8329a3a756e8a54cf95b2d246eb)
to alleviate GC pressure. Prior to this, sometimes, we use the memory address of a `Node` as a unique
identifier but that trick was no longer valid, since now `Node` can be recycled and reused. This new `int64`
`ID` is added to help uniquely identify a `Node`'s use, whether the `Node` is allocated brand new for the
use or is recycled for this use. The `ID` value is monotonically increasing but it really doesn't matter - what
it matters is each "acquisition" of a new use of a `Node` will be guaranteed to have a new and unique `ID` value.

## XML

Since XML is the most complex input format we have for IDR, let's cover it first.

Here is a simple XML (from [this sample](../extensions/omniv21/samples/xml/1_datetime_parse_and_format.input.xml)):
```
<Root>
    <JustDate>2020/09/22</JustDate>
    <DateTimeWithNoTZ>09/22/2020 12:34:56</DateTimeWithNoTZ>
</Root>
```
This is a simple XML blob with no non-default namespaces and with no attributes. Its corresponding IDR
looks like this (with `ID`, empty fields and tree pointers omitted for clarity):
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
which we'll see in the [next example](../extensions/omniv21/samples/xml/2_multiple_objects.input.xml):
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

The IDR for the example above looks like the following (note those "dummy" text nodes sprinkled
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
Both `Node`'s representing `<lb0:library>` and `<lb0:books>` include non-empty `XMLSpecific`'s which
contain their namespace prefixes and full URIs while their `Node.Data` contain the element names without
the namespace prefixes.

Note XML attributes on elements are represented as `Node`'s as well, with `Type: AttributeNode`
specifically. If an attribute is namespace-prefixed, the `AttributeNode` typed `Node` will have a non-empty
`XMLSpecific` set as well. An attribute's actual value is placed as a `TextNode` underneath its `ElementNode`.
`AttributeNode`'s are guaranteed to be placed before any other child nodes (`TextNode`, or `ElementNode`)
by IDR's XML reader.

## JSON

Here is a sample JSON (adapted from [this sample](../extensions/omniv21/samples/json/1_single_object.input.json)):
```
{
    "order_id": "1234567",
    "items": [
        {
            "number_purchased": 5
        },
        {
            "item_price": 3.99
            "refundable": false
            "refund_id": null
        }
    ]
}
```
Its corresponding IDR looks like this (with `ID`, and tree pointers omitted for clarity):
```
Node(Type: DocumentNode, FormatSpecific: JSONRoot|JSONObj)
    Node(Type: ElementNode, Data: "order_id", FormatSpecific: JSONProp)
        Node(Type: TextNode, Data: "1234567", FormatSpecific: JSONValueStr)
    Node(Type: ElementNode, Data: "items", FormatSpecific: JSONProp|JSONArr)
        Node(Type: ElementNode, Data: "", FormatSpecific: JSONObj)
            Node(Type: ElementNode, Data: "number_purchased", FormatSpecific: JSONProp)
                Node(Type: TextNode, Data: "5", FormatSpecific: JSONValueNum)
        Node(Type: ElementNode, Data: "", FormatSpecific: JSONObj)
            Node(Type: ElementNode, Data: "item_price", FormatSpecific: JSONProp)
                Node(Type: TextNode, Data: "3.99", FormatSpecific: JSONValueNum)
            Node(Type: ElementNode, Data: "refundable", FormatSpecific: JSONProp)
                Node(Type: TextNode, Data: "false", FormatSpecific: JSONValueBool)
            Node(Type: ElementNode, Data: "refund_id", FormatSpecific: JSONProp)
                Node(Type: TextNode, Data: "", FormatSpecific: JSONValueNull)
```
For JSON IDR, the `FormatSpecific` field contains [`JSONType`](./node.go#L9). `JSONType` are a bunch of
bit-wise flags that can be combined: in this example above, the root flag is `JSONRoot|JSONObj` because it
is the root at the same time, it is an object. If we have a JSON looks like this:
```
[
    ...
]
```
Then the root flag would be `JSONRoot|JSONArr`. Such combination can also be seen on `items` field: its
flag is `JSONProp|JSONArr`, indicating `items` is a property of array value.

Similar to XML case, values (string/number/boolean/null) are added as `TextNode` and anchored below its
corresponding `ElementNode` parent.

## CSV (aka delimited)

Here is a sample CSV (adapted from [this sample](../extensions/omniv21/samples/csv/1_weather_data_csv.input.csv)):
```
DATE|HIGH TEMP C|LOW TEMP F|WIND DIR|WIND SPEED KMH|NOTE|LAT|LONG|UV INDEX
2019/01/31T12:34:56-0800|10.5|30.2|N|33|note 1|37.7749|122.4194|12/4/6
```
The omniparser builtin CSV reader will only return data rows as IDR trees, so for this example, the row
2 will be returned as:
```
Node(Type: DocumentNode)
    Node(Type: ElementNode, Data: "DATE")
        Node(Type: TextNode, Data: "2019/01/31T12:34:56-0800")
    Node(Type: ElementNode, Data: "HIGH_TEMP_C")
        Node(Type: TextNode, Data: "10.5")
    Node(Type: ElementNode, Data: "LOW_TEMP_F")
        Node(Type: TextNode, Data: "30.2")
    Node(Type: ElementNode, Data: "WIND_DIR")
        Node(Type: TextNode, Data: "N")
    Node(Type: ElementNode, Data: "WIND_SPEED_KMH")
        Node(Type: TextNode, Data: "33")
    Node(Type: ElementNode, Data: "NOTE")
        Node(Type: TextNode, Data: "note 1")
    Node(Type: ElementNode, Data: "LAT")
        Node(Type: TextNode, Data: "37.7749")
    Node(Type: ElementNode, Data: "LONG")
        Node(Type: TextNode, Data: "122.4194")
    Node(Type: ElementNode, Data: "UV_INDEX")
        Node(Type: TextNode, Data: "12/4/6")
```
Note some of the `ElementNode.Data` values are from its [schema](../extensions/omniv21/samples/csv/1_weather_data_csv.schema.json)
to avoid column name containing space, which isn't xpath query friendly.

## EDI

Here is snippet from a [sample EDI](../extensions/omniv21/samples/edi/1_canadapost_edi_214.input.txt):

```
ISA*00*          *00*          *02*CPC            *ZZ*00602679321    *191103*1800*U*00401*000001644*0*P*>
GS*QM*CPC*00602679321*20191103*1800*000001644*X*004010
<omitted>
N4*HAMMER*AB*T0C1Z0*CA
<omitted>
```

Note EDI content, while itself is "flat", is inherently hierarchical (like XML) and its hierarchical structure
needs to be described in an accompanying schema. For this example, here is a snippet of the 
[related schema](../extensions/omniv21/samples/edi/1_canadapost_edi_214.schema.json):

```
"segment_declarations": [
    {
        "name": "ISA",
        "child_segments": [
            {
                "name": "GS",
                "child_segments": [
                    {
                        "name": "scanInfo", "type": "segment_group", "min": 0, "max": -1, "is_target": true,
                        "child_segments": [
                            <omitted>
                            {
                                "name": "N4",
                                "elements": [
                                    { "name": "cityName", "index": 1 },
                                    { "name": "provinceCode", "index": 2 },
                                    { "name": "postalCode", "index": 3 },
                                    { "name": "countryCode", "index": 4 }
                                ]
                            },
                            <omitted>
```

The omniparser builtin EDI reader constructs the following IDR node(s) for each segment of data:

```
Node(Type: ElementNode, Data: "ISA")
```

This `"ISA"` IDR element node contains no children.

```
Node(Type: ElementNode, Data: "N4")
    Node(Type: ElementNode, Data: "cityName")
        Node(Type: TextNode, Data: "HAMMER")
    Node(Type: ElementNode, Data: "provinceCode")
        Node(Type: TextNode, Data: "AB")
    Node(Type: ElementNode, Data: "postalCode")
        Node(Type: TextNode, Data: "T0C1Z0")
    Node(Type: ElementNode, Data: "countryCode")
        Node(Type: TextNode, Data: "CA")
```

This `"N4"` IDR element node contains multiple child element nodes for each of the element appearances
in the EDI content.

## Fixed-Length (mostly TXT)

In fixed-length files, we have the concept of an 'envelope'. An envelope can contain a single line of text, or a fixed
number of consecutive lines, or multiple lines grouped together by a header and a footer. An envelope is the basic unit
of record which the fixed-length reader returns to the parser.
 
Here is a sample single-line-envelope TXT (adapted from [this sample](../extensions/omniv21/samples/fixedlength/1_single_row.input.txt)):
```
...
2019/01/31T12:34:56-0800 10.5 30.2  N 33  37.7749 122.4194
...
```
The omniparser builtin fixed-length reader returns a fixed-length envelope as a IDR tree:
```
Node(Type: ElementNode)
    Node(Type: ElementNode, Data: "DATE")
        Node(Type: TextNode, Data: "2019/01/31T12:34:56-0800")
    Node(Type: ElementNode, Data: "HIGH_TEMP_C")
        Node(Type: TextNode, Data: "10.5")
    Node(Type: ElementNode, Data: "LOW_TEMP_F")
        Node(Type: TextNode, Data: "30.2")
    Node(Type: ElementNode, Data: "WIND_DIR")
        Node(Type: TextNode, Data: "N")
    Node(Type: ElementNode, Data: "WIND_SPEED_KMH")
        Node(Type: TextNode, Data: "33")
    Node(Type: ElementNode, Data: "LAT")
        Node(Type: TextNode, Data: "37.7749")
    Node(Type: ElementNode, Data: "LONG")
        Node(Type: TextNode, Data: "122.4194")
```

# XPath Based Record Filtering and Data Extraction

The foundation of omniparser transform operations is anchored on [IDR](./idr.md) and XPath based record
filtering and data extraction. It's vital to understand each supported file format's IDR structure to
effectively and efficiently craft XPath queries in `transform_declarations` to achieve desire transform
objectives.

## Record Filtering

Many times some records ingested are not suitable/desirable to be transformed into output. Omniparser, more
specifically the current latest version (`"omni.2.1"`) handler, allows record level filtering using XPath
query. Let's see one example in CSV:

```
ORDER_ID,CUSTOMER_ID,COUNTRY
1234,CUST_1,US
N/A
1235,CUST_2,AU
```

We want omniparser to ingest and transform records with `order_id=1234,1235` and skip the line with
`'N/A'`. To achieve this, we can insert `xpath` into the root `object` of `FINAL_OUTPUT` in
`transform_declarations`:

```
"transform_declarations": {
    "FINAL_OUTPUT": { "xpath": ".[matches(ORDER_ID, '^[0-9]+$')]", "object": {
        ...
    }}
}
```

Let's take a look how the transform works for first data line `1234,CUST_1,US`:
1. Omniparser reads the first line in and converts it into a [CSV specific IDR tree](./idr.md#csv-aka-delimited):
    ```
    Node(Type: DocumentNode)
        Node(Type: ElementNode, Data: "ORDER_ID")
            Node(Type: TextNode, Data: "1234")
        Node(Type: ElementNode, Data: "CUSTOMER_ID")
            Node(Type: TextNode, Data: "CUST_1")
        Node(Type: ElementNode, Data: "COUNTRY")
            Node(Type: TextNode, Data: "US")
    ```
2. `FINAL_OUTPUT.xpath` is then executed at the root of the IDR tree, and result is a match! So this
line/record will be processed.

Now take a look the second line `N/A`:
1. The IDR tree looks like:
    ```
    Node(Type: DocumentNode)
        Node(Type: ElementNode, Data: "ORDER_ID")
            Node(Type: TextNode, Data: "N/A")
    ```
2. `FINAL_OUTPUT.xpath` is executed at the root of the IDR tree, and result is not a match. This line/record
will be skipped.

Each input format has its own unique IDR structure, record filtering XPath needs to take it into consideration
to be effective.

Clever use of positive/negative regexp [`matches`](https://github.com/antchfx/xpath#expressions) (slightly
slower but very powerful), or [`starts-with`, `ends-with`, `contains`](https://github.com/antchfx/xpath#expressions),
or even direct string comparisons [`==`, `!=`](https://github.com/antchfx/xpath#expressions) in
`FINAL_OUTPUT.xpath` gives schema writers the freedom of either processing certain lines/records, or skipping
certain lines/records.

If `FINAL_OUTPUT` doesn't have `xpath`, which is fairly common, then there is no record filtering, meaning
all records ingested by omniparser file format specific readers will be processed and transformed.

## Data Extraction

The most common use of `xpath` is for data extraction. Consider again the sample CSV and schema in
[Record Filtering](#record-filtering), let's amend the schema to:
```
"transform_declarations": {
    "FINAL_OUTPUT": { "xpath": ".[matches(ORDER_ID, '^[0-9]+$')]", "object": {
        "order_id": { "xpath": "ORDER_ID", "type": "int" },
        "customer_id": { "xpath": "CUSTOMER_ID", "type": "int" },
        "country": { "xpath": "COUNTRY" }
    }}
}
```

The `xpath` attributes on `"order_id"`, `"customer_id"`, and `"country"` tell omniparser where to get
the field string data from. When `xpath` **not** appearing with `object`, `template`, `custom_func`, or
`custom_parse`, then it is a data extraction directive telling omniparser to extract the text data at the
location specified by the `xpath` query. Note in this situation, omniparser will require the result set of
such `xpath` queries to be of a single node: if such `xpath` query results in more than one node, omniparser
will fail the current record transform (but will continue onto the next one as this isn't considered fatal).

## Data Context and Anchoring

Whether `xpath` is used for record filtering or data extraction/anchoring, it's always good to know the
current IDR tree "cursor" position against which an `xpath` query, if present, will be executed.

The current "cursor" position when a transform of `FINAL_OUTPUT` starts is always at the top of an IDR tree.
So record filtering `FINAL_OUTPUT.xpath` is always executed against the root fo the IDR tree. The "cursor"
position remains unchanged until a new anchoring `xpath` is encountered. Typically, schema writers will need
to change cursor anchoring positions more often in hierarchical file formats, such as EDI/JSON/XML, than
"flat" file formats, like fixed-length or CSV.

Let's take a look at a [sample schema](../extensions/omniv21/samples/json/2_multiple_objects.schema.json)
for JSON input:

```
1    "transform_declarations": {
2        "FINAL_OUTPUT": { "xpath": "/publishers/*", "object": {
3            "authors": { "array": [ { "xpath": "books/*/author" } ] },
4            "book_titles": { "array": [ { "xpath": "books/*/title" } ] },
5            "books": { "array": [ { "xpath": "books/*", "object": {
6                "author": { "xpath": "author" },
7                "year": { "xpath": "year", "type": "int" },
8                "price": { "xpath": "price", "type": "float" },
9                "title": { "xpath": "title" }
10            }} ] },
11            "publisher": { "xpath": "name" },
12            "first_book": { "xpath": "books/*[position() = 1]", "custom_func": { "name": "copy" }},
13            "original_book_array": { "xpath": "books", "custom_func": { "name": "copy" }}
41        }}
42    }
```
Notes:
- Line numbers are added for easier reference.
- Only `transform_declarations` section is included here for brevity.

Consider this [input](../extensions/omniv21/samples/json/2_multiple_objects.input.json):
```
1 {
2     "publishers": [
3         {
4             "name": "Scholastic Press",
5             "books": [
6                 {
7                     "title": "Harry Potter and the Philosopher's Stone",
8                     "price": 9.99,
9                     "author": "J. K. Rowling",
10                     "year": 1997
11                 },
12                 {
13                     "title": "Harry Potter and the Chamber of Secrets",
14                     "price": 10.99,
15                     "author": "J. K. Rowling",
16                     "year": 1998
17                 }
18             ]
19         }
20 }
```

Now let's go through the schema and input together to see how `xpath` anchoring is used.

1. schema `2        "FINAL_OUTPUT": { "xpath": "/publishers/*", "object": {`

    This is record filtering, saying, we'd like to process and transform every record matching
    `/publishers/*`. In this simplified input example, there is only one JSON object matches it: it's the
    object starting at line 3 and finishing at line 19. With this line, the transform starts, and now the
    cursor is anchored at the top of this object.

2. schema `3            "authors": { "array": [ { "xpath": "books/*/author" } ] },`

    Unlike `object` transform, `array` transform itself doesn't/may not have `xpath` attribute: an `array`
    transform is a collection of child transforms, each of which can optionally have its own `xpath`.
    This schema line says, `authors` in the output is an array, of which, each element is a string whose
    value comes from the `xpath` data extraction `books/*/author`. So with the input above, we will have
    `"authors": [ "J. K. Rowling", "J. K. Rowling" ]` in the final output.

3. schema `4            "book_titles": { "array": [ { "xpath": "books/*/title" } ] },`

    Very similar to `authors` output above, `book_titles` output will be like:
    `"book_titles": [ "Harry Potter and the Philosopher's Stone", "Harry Potter and the Chamber of Secrets" ]`
    in the final output.

4. schema `5            "books": { "array": [ { "xpath": "books/*", "object": {`

    Similar to `authors` and `book_titles` above, what this line says is, `books` in the output should be an
    array of objects, each of which, the IDR cursor should be anchored on `books/*` for its processing and
    transform. In other words, omniparser will anchor the IDR cursor on the JSON object from line 6 through
    line 11 for the first array element object transform, and then anchor on the JSON object from line 12
    through line 17 for the second array element object transform.

5. schema `6                "author": { "xpath": "author" },` and through line 9
    Recall in 4., omniparser has put the cursor on actual book object. Now line 6 through line 9 simply
    extract string values from the object and put into the corresponding output fields.

6. schema `12            "first_book": { "xpath": "books/*[position() = 1]", "custom_func": { "name": "copy" }},`

    This is an interesting schema construct: we want `first_book` in the output to be a direct copy of the
    first book object inside input's `books` JSON array. `"xpath": "books/*[position() = 1]"` achieves the
    "only first book object" filtering. `"custom_func": { "name": "copy" }` achieves the direct copying.

    As you notice, `custom_func` transform can have (optional) `xpath` attribute as well. If `xpath` is present
    for a `custom_func`, then everything inside the `custom_func`, namely those argument transforms, are all
    anchored on the cursor position prescribed by the `xpath`.

When `xpath` is used for anchoring and cursoring, it can appear with `object`, `template`, `custom_func`, and
`custom_parse`.

## Static and Dynamic XPath Queries

While `xpath` is the most commonly used filtering, anchoring and data extraction directive in schemas, it (the
query itself) is completely static, meaning the query is fixed and static at schema writing time, thus can't
be used where data dependent runtime dynamic query is needed.

Consider the following [sample input](../extensions/omniv21/samples/json/3_xpathdynamic.input.json):
```
[
    {
        "line_items": [
            {
                "product": {
                    "variant": {
                        "option2": "Blue",
                        "option1": "M"
                    },
                    "options": [
                        {
                            "index": 2,
                            "name": "color/pattern",
                            "values": [
                                "Blue",
                                "Green"
                            ]
                        },
                        {
                            "index": 1,
                            "name": "Size",
                            "values": [
                                "M",
                                "L"
                            ]
                        }
                    ]
                }
            }
        ]
    }
]
```
Notice the `options` array specifies what allowed/possible options are for a product and then in `variant`
of `product`, it specifies what actual options are included.

The [sample schema](../extensions/omniv21/samples/json/3_xpathdynamic.schema.json):
```
"transform_declarations": {
    "FINAL_OUTPUT": { "xpath": "/*", "object": {
        "order_info": { "object": {
            "order_items": { "array": [
                { "xpath": "line_items/*", "object": {
                    ...
                    "color": { "xpath_dynamic": {
                        "custom_func": {
                            "name": "concat",
                            "args": [
                                { "const": "product/variant/option" },
                                { "xpath": "product/options/*[name='color/pattern']/index" }
                            ]
                        }
                    }},
                    "size": { "xpath_dynamic": {
                        "custom_func": {
                            "name": "concat",
                            "args": [
                                { "const": "product/variant/option" },
                                { "xpath": "product/options/*[name='Size']/index" }
                            ]
                        }
                    }},
                    ...
                }}
            ]}
        }}
    }}
}
```

The schema wants to transform `optoin1` and `option2` in the input into `color` and `size` in output. The
difficulty is how to figure out `optoin1` is mapped to `color` and `option2` to `size`. If we look at the
input's `options` array, it says `"index": 1` is for size and `"index": 2` is for color. To extract data
for `color` field in the output, we need to dynamically construct an XPath query by
`product/variant/option` + `product/options/*[name='color/pattern']/index`. Similar XPath construction is
needed for `size` field data extraction.

`xpath_dynamic` is used in such a situation. It basically says, unlike `xpath` is always a constant and static
string value, `xpath_dynamic` is computed, by either `custom_func`, or `custom_parse`, or `template`, or
`external`, or `const`, or another `xpath` direct data extraction.

`xpath_dynamic` can be used everywhere `xpath` is used, except on `FINAL_OUTPUT`. `FINAL_OUTPUT` can only
use `xpath`.

## XPath Query Result-set Cardinality

Everytime when an `xpath` or `xpath_dynamic` query is executed against an IDR node (and its subtree), the
result is always a set of nodes: could be an empty set, or a set of one node, or a set of multiple nodes.
Depending on which transform is in play, different outcomes, including error, can follow.

- `xpath`/`xpath_dynamic` used alone, aka data extraction transform:

    - Example: `"field1": { "xpath": "PATH/TO/DATA" }`
    - The result set must be either empty or of a single node. When empty, `""` is used; when a single
    node is returned for the query, the node's text data will be used; when more than one node is returned,
    omniparser will return a transform error (non-fatal).

- `xpath` used in `FINAL_OUTPUT`:

    - Example: `"FINAL_OUTPUT": { "xpath": "/publishers/*", "object": {`
    - The result set can be either empty, or of one node, or of multiple nodes.

- `xpath`/`xpath_dynamic` used in `object`, `custom_func`, `custom_parse`, `template` transform
(other than `FINAL_OUTPUT` or directly under an `array` transform):

    - Example: `"contact": { "xpath": "PATH/TO/CONTACT", "object": {`
    - Example: `"temperature": { "xpath": "PATH/TO/TEMPERATURE", "custom_func": {`
    - Example: `"wind_forecast": { "xpath": "PATH/TO/WIND", "template": {`
    - The result set can only be either empty or of one node. Multiple node result set will cause parser error.

- `xpath`/`xpath_dynamic` used in transform that is directly under `array` transform:

    - Example: `"titles": { "array": [ { "xpath": "books/*/title" } ] }`
    - Example: `"titles": { "array": [ { "xpath": "books/*/title" }, { "xpath": "movies/*/title" } ] }`
    - The first example is the most commonly used scenario, that is, the `array` contains homogeneous element
    transforms. In this case, the `xpath` can return empty, or one node, or multiple nodes and results will
    be used as the array's elements.
    - The second example shows the flexibility of `array` transform, that it can contain different transforms:
    one set of titles is of book titles and another set of movie titles. All titles, books' or movies', are
    contained by the array. Similar to the first case, both `xpath` result sets can return empty, one node or
    multiple nodes. All are fine and accepted by the parser.

## Supported XPath Features

Omniparser relies on https://github.com/antchfx/xpath (thank you!) for XPath query parsing and execution.
Check its github page for the full syntax and function support list.

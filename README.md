# omniparser
![CI](https://github.com/jf-tech/omniparser/workflows/CI/badge.svg) [![codecov](https://codecov.io/gh/jf-tech/omniparser/branch/master/graph/badge.svg)](https://codecov.io/gh/jf-tech/omniparser) [![Go Report Card](https://goreportcard.com/badge/github.com/jf-tech/omniparser)](https://goreportcard.com/report/github.com/jf-tech/omniparser)

A parser in naive Golang that ingests and transforms input data of various formats (CSV, txt, XML, EDI, JSON)
into desired JSON output based on a schema spec written in JSON.

Golang Version: 1.14.2

## Demo in Playground

Use https://omniparser.herokuapp.com/ (might need to wait for a few seconds for heroku instance to wake up)
for trying out schemas and inputs, yours and from sample library, to see how transform works.

![](./cli/cmd/web/playground-demo.gif)

Take a detailed look at samples here:
- [json examples](./samples/omniv2/json)
- [xml examples](./samples/omniv2/xml).

## Simple Example (JSON -> JSON Transform)
- Input:
    ```
    {
        "order_id": "1234567",
        "tracking_number": "1z9999999999999999",
        "items": [
            {
                "item_sku": "ab123",
                "item_price": 12.34,
                "number_purchased": 5
            },
            {
                "item_sku": "ck763-23",
                "item_price": 3.12,
                "number_purchased": 2
            }
        ]
    }
    ```
- Schema:
    ```
    {
        "parser_settings": {
            "version": "omni.2.0",
            "file_format_type": "json"
        },
        "transform_declarations": {
            "FINAL_OUTPUT": { "xpath": ".", "object": {
                "order_id": { "xpath": "order_id" },
                "tracking_number": { "custom_func": {
                    "name": "upper",
                    "args": [ { "xpath": "tracking_number" } ]
                }},
                "items": { "array": [{ "xpath": "items/*", "object": {
                    "sku":  { "custom_func": {
                        "name": "substring",
                        "args": [
                            { "custom_func": { "name": "upper", "args": [ { "xpath": "item_sku" }]}},
                            { "const": "0", "_comment": "start index" },
                            { "const": "5", "_comment": "sub length" }
                        ]
                    }},
                    "total_price": { "custom_func": {
                        "name": "javascript",
                        "args": [
                            { "const": "num * price" },
                            { "const": "num:int" }, { "xpath": "number_purchased" },
                            { "const": "price:float" }, { "xpath": "item_price" }
                        ]
                    }}
                }}]}
            }}
        }
    }
    ```
- Code:
    ```
    schema, err := omniparser.NewSchema("schema-name", strings.NewReader("..."))
    if err != nil { ... }
    transform, err := parser.NewTransform("input-name", strings.NewReader("..."), &transformctx.Ctx{})
    if err != nil { ... }
    if !transform.Next() { ... }  
    b, err := transform.Read()
    if err != nil { ... }
    fmt.Println(string(b))
    ```
- Output:
    ```
    {
        "order_id": "1234567",
        "tracking_number": "1Z9999999999999999",
        "items": [
            {
                "sku": "AB123",
                "total_price": "61.7"
            },
            {
                "sku": "CK763",
                "total_price": "6.24"
            }
        ]
    }
    ```

## Why
- No good ETL transform/parser library exists in Golang.
- Even looking into Java and other languages, choices aren't many and all have limitations:
    - [Smooks](https://www.smooks.org/) is dead, plus its EDI parsing/transform is too heavyweight, needing code-gen.
    - [BeanIO](http://beanio.org/) can't deal with EDI input.
    - [Jolt](https://github.com/bazaarvoice/jolt) can't deal with anything other than JSON input.
    - [JSONata](https://jsonata.org/) still only JSON -> JSON transform.
- Many of the parsers/transforms don't support streaming read, loading entire input into memory - not acceptable in some situations.

## Requirements
- Golang 1.14

    This is only needed for `javascript` engine integration. Please raise an issue if you think 1.14 is too high, and
    you don't need `javascript` custom_func. Then we may consider moving `javascript` custom_func into a separate
    extension repo/package; the rest of the library is just golang 1.12.

## Recent Feature Additions
- added trie based high performance `times.SmartParse`.
- command line interface (one-off `transform` cmd or long-running http `server` mode).
- javascript engine integration as a custom_func.
- JSON stream parser.
- Extensibility:
    - Ability to provide custom functions.
    - Ability to provide custom schema handler.
    - Ability to customize the built-in omniv2 schema handler's parsing code.
    - Ability to provide a new file format support to built-in omniv2 schema handler.

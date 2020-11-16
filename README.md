# omniparser
![CI](https://github.com/jf-tech/omniparser/workflows/CI/badge.svg) [![codecov](https://codecov.io/gh/jf-tech/omniparser/branch/master/graph/badge.svg)](https://codecov.io/gh/jf-tech/omniparser) [![Go Report Card](https://goreportcard.com/badge/github.com/jf-tech/omniparser)](https://goreportcard.com/report/github.com/jf-tech/omniparser)

Omniparser is written in naive Golang that ingests input data of various formats (**CSV, txt, XML, EDI, JSON**, and custom formats) in streaming fashion 
and transforms data into desired JSON output based on a schema written in JSON.

Golang Version: 1.14.2

## Demo in Playground

Use https://omniparser.herokuapp.com/ (may need to wait for a few seconds for heroku instance to wake up)
for trying out schemas and inputs, yours and from sample library, to see how transform works.

![](./cli/cmd/web/playground-demo.gif)

Take a detailed look at samples here:
- [csv examples](extensions/omniv21/samples/csv)
- [fixed-length examples](extensions/omniv21/samples/fixedlength)
- [json examples](extensions/omniv21/samples/json)
- [xml examples](extensions/omniv21/samples/xml).
- [edi examples](extensions/omniv21/samples/edi).

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
            "version": "omni.2.1",
            "file_format_type": "json"
        },
        "transform_declarations": {
            "FINAL_OUTPUT": { "object": {
                "order_id": { "xpath": "order_id" },
                "tracking_number": { "custom_func": {
                    "name": "upper",
                    "args": [ { "xpath": "tracking_number" } ]
                }},
                "items": { "array": [{ "xpath": "items/*", "object": {
                    "sku":  { "custom_func": {
                        "name": "javascript",
                        "args": [
                            { "const": "sku.toUpperCase().substring(0, 5)" },
                            { "const": "sku" }, { "xpath": "item_sku" }
                        ]
                    }},
                    "total_price": { "custom_func": {
                        "name": "javascript",
                        "args": [
                            { "const": "num * price" },
                            { "const": "num" }, { "xpath": "number_purchased", "type": "int" },
                            { "const": "price" }, { "xpath": "item_price", "type": "float" }
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
    transform, err := schema.NewTransform("input-name", strings.NewReader("..."), &transformctx.Ctx{})
    if err != nil { ... }
    for {
        b, err := transform.Read()
        if err == io.EOF { break }
        if err != nil { ... }
        fmt.Println(string(b))
    }
    ```
- Output:
    ```
    {
        "order_id": "1234567",
        "tracking_number": "1Z9999999999999999",
        "items": [
            {
                "sku": "AB123",
                "total_price": 61.7
            },
            {
                "sku": "CK763",
                "total_price": 6.24
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

## Recent Major Feature Additions/Changes
- Added fixed-length file format support in omniv21 handler.
- Added EDI file format support in omniv21 handler.
- Major restructure/refactoring
    - Upgrade omni schema version to `omni.2.1` due a number of incompatible schema changes:
        - `'result_type'` -> `'type'`
        - `'ignore_error_and_return_empty_str` -> `'ignore_error'`
        - `'keep_leading_trailing_space'` -> `'no_trim'` 
    - Changed how we handle custom functions: previously we always use strings as in param type as well as result param
    type. Not anymore, all types are supported for custom function in and out params.
    - Changed the way how we package custom functions for extensions: previously we collect custom functions from all
    extensions and then pass all of them to the extension that is used; This feels weird, now changed to only the custom
    functions included in a particular extension are used in that extension.
    - Deprecated/removed most of the custom functions in favor of using 'javascript'. 
    - A number of package renaming.
- Added CSV file format support in omniv2 handler.
- Introduced IDR node cache for allocation recycling. 
- Introduced [IDR](./idr/README.md) for in-memory data representation.
- Added trie based high performance `times.SmartParse`.
- Command line interface (one-off `transform` cmd or long-running http `server` mode).
- `javascript` engine integration as a custom_func.
- JSON stream parser.
- Extensibility:
    - Ability to provide custom functions.
    - Ability to provide custom schema handler.
    - Ability to customize the built-in omniv2 schema handler's parsing code.
    - Ability to provide a new file format support to built-in omniv2 schema handler.

## Footnotes
- omniparser is a collaboration effort of [jf-tech](https://github.com/jf-tech/), [Simon](https://github.com/liangxibing) and [Steven](http://github.com/wangjia007bond).

# omniparser
![CI](https://github.com/jf-tech/omniparser/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/jf-tech/omniparser/branch/master/graph/badge.svg)](https://codecov.io/gh/jf-tech/omniparser)
[![Go Report Card](https://goreportcard.com/badge/github.com/jf-tech/omniparser)](https://goreportcard.com/report/github.com/jf-tech/omniparser)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/jf-tech/omniparser)](https://pkg.go.dev/github.com/jf-tech/omniparser)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Omniparser is a native Golang ETL parser that ingests input data of various formats (**CSV, txt, fixed length/width,
XML, EDI/X12/EDIFACT, JSON**, and custom formats) in streaming fashion and transforms data into desired JSON output
based on a schema written in JSON.

Min Golang Version: 1.14

## Licenses and Sponsorship
Omniparser is publicly available under [MIT License](./LICENSE).
[Individual and corporate sponsorships](https://github.com/sponsors/jf-tech/) are welcome and gratefully
appreciated, and will be listed in the [SPONSORS](./sponsors/SPONSORS.md) page.
[Company-level sponsors](https://github.com/sponsors/jf-tech/) get additional benefits and supports
granted in the [COMPANY LICENSE](./sponsors/COMPANY_LICENSE.md).

## Documentation

Docs:
- [Getting Started](./doc/gettingstarted.md): a tutorial for writing your first omniparser schema.
- [IDR](./doc/idr.md): in-memory data representation of ingested data for omniparser.
- [XPath Based Record Filtering and Data Extraction](./doc/xpath.md): xpath queries are essential to omniparser schema
writing. Learn the concept and tricks in depth.
- [All About Transforms](./doc/transforms.md): everything about `transform_declarations`.
- [Use of `custom_func`, Specially `javascript`](./doc/use_of_custom_funcs.md): An in depth look of how `custom_func`
is used, specially the all mighty `javascript` (and `javascript_with_context`).
- [CSV Schema in Depth](./doc/csv2_in_depth.md): everything about schemas for CSV input.
- [Fixed-Length Schema in Depth](./doc/fixedlength2_in_depth.md): everything about schemas for fixed-length (e.g. TXT)
input
- [JSON/XML Schema in Depth](./doc/json_xml_in_depth.md): everything about schemas for JSON or XML input.
- [EDI Schema in Depth](./doc/edi_in_depth.md): everything about schemas for EDI input.
- [Programmability](./doc/programmability.md): Advanced techniques for using omniparser (or some of its components) in
your code.

References:
- [Custom Functions](./doc/customfuncs.md): a complete reference of all built-in custom functions.

Examples:
- [CSV Examples](extensions/omniv21/samples/csv2)
- [Fixed-Length Examples](extensions/omniv21/samples/fixedlength2)
- [JSON Examples](extensions/omniv21/samples/json)
- [XML Examples](extensions/omniv21/samples/xml).
- [EDI Examples](extensions/omniv21/samples/edi).
- [Custom File Format](extensions/omniv21/samples/customfileformats/jsonlog)
- [Custom Funcs](extensions/omniv21/samples/customfuncs)

In the example folders above you will find pairs of input files and their schema files. Then in the
`.snapshots` sub directory, you'll find their corresponding output files.

## Online Playground (not functioning)

~~Use [The Playground](https://omniparser-prod-omniparser-qd0sj4.mo2.mogenius.io/)  (may need to wait for a few seconds for instance to wake up)
for trying out schemas and inputs, yours or existing samples, to see how ingestion and transform work.~~

As for now (2023/03/14), all of our previous free docker hosting solutions went away and we haven't found another one yet. For now please clone the repo and use `./cli.sh` as described in the [Getting Started](./doc/gettingstarted.md) page.

## Why
- No good ETL transform/parser library exists in Golang.
- Even looking into Java and other languages, choices aren't many and all have limitations:
    - [Smooks](https://www.smooks.org/) is dead, plus its EDI parsing/transform is too heavyweight, needing code-gen.
    - [BeanIO](http://beanio.org/) can't deal with EDI input.
    - [Jolt](https://github.com/bazaarvoice/jolt) can't deal with anything other than JSON input.
    - [JSONata](https://jsonata.org/) still only JSON -> JSON transform.
- Many of the parsers/transforms don't support streaming read, loading entire input into memory - not acceptable in some
situations.

## Requirements
- Golang 1.14 or later.

## Recent Major Feature Additions/Changes
- 2022/09: v1.0.4 released: added `csv2` file format that supersedes the original `csv` format with support of hierarchical and nested records.
- 2022/09: v1.0.3 released: added `fixedlength2` file format that supersedes the original `fixed-length` format with support of hierarchical and nested envelopes.
- 1.0.0 Released!
- Added `Transform.RawRecord()` for caller of omniparser to access the raw ingested record.
- Deprecated `custom_parse` in favor of `custom_func` (`custom_parse` is still usable for
back-compatibility, it is just removed from all public docs and samples).
- Added `NonValidatingReader` EDI segment reader.
- Added fixed-length file format support in omniv21 handler.
- Added EDI file format support in omniv21 handler.
- Major restructure/refactoring
    - Upgrade omni schema version to `omni.2.1` due a number of incompatible schema changes:
        - `'result_type'` -> `'type'`
        - `'ignore_error_and_return_empty_str` -> `'ignore_error'`
        - `'keep_leading_trailing_space'` -> `'no_trim'`
    - Changed how we handle custom functions: previously we always use strings as in param type as well as result param
    type. Not anymore, all types are supported for custom function in and out params.
    - Changed the way we package custom functions for extensions: previously we collected custom functions from all
    extensions and then passed all of them to the extension that is used; this feels weird, now only the custom
    functions included in a particular extension are used in that extension.
    - Deprecated/removed most of the custom functions in favor of using 'javascript'.
    - A number of package renaming.
- Added CSV file format support in omniv2 handler.
- Introduced IDR node cache for allocation recycling.
- Introduced [IDR](./doc/idr.md) for in-memory data representation.
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
- omniparser is a collaboration effort of [jf-tech](https://github.com/jf-tech/),[Simon](https://github.com/liangxibing)
and [Steven](http://github.com/wangjia007bond).

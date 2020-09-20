# omniparser
![CI](https://github.com/jf-tech/omniparser/workflows/CI/badge.svg) [![codecov](https://codecov.io/gh/jf-tech/omniparser/branch/master/graph/badge.svg)](https://codecov.io/gh/jf-tech/omniparser) [![Go Report Card](https://goreportcard.com/badge/github.com/jf-tech/omniparser)](https://goreportcard.com/report/github.com/jf-tech/omniparser)

A data transform parser in naive golang that transforms input data of various formats (CSV, txt, XML, EDI, JSON)
into desired JSON output based on a schema spec written in JSON.

## Simple Example
- Input:
    ```
    TBD
    ```
- Schema:
    ```
    TBD
    ```
- Code:
    ```
    TBD
    ```
- Output:
    ```
    TBD
    ```
## Playground

You can use https://omniparser.herokuapp.com/ for trying out schemas and inputs and see how the transform works.
You can try out these [json examples](./samples/omniv2/json) or [xml examples](./samples/omniv2/xml). 

## Why

## Recent Feature Additions
- command line interface
- javascript engine integration as a custom_func.
- JSON stream parser
- Extensibility
    - Ability to provide custom functions.
    - Ability to provide custom schema plugins.
    - Ability to customize the built-in omniv2 plugin's parsing code.
    - Ability to provide a new file format support to built-in omniv2 plugin.

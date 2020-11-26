# All About Transforms

The current latest version of schema for omniparser is `omni.2.1`, in which the most important section is
`transform_declarations`. Previously in [Getting Started](./gettingstarted.md) we've shown a glimpse of
what `transform_declarations` looks like. Now let's formally cover this schema section.

`transform_declarations` is a collection of templates defining what pieces of output should look like,
including those pieces' structure definitions and transformation directives. One of those pieces is
`FINAL_OUTPUT`, a specially named template that omniparser will use for constructing the result record.
Each template can reference other templates, recursively, as long as there are no circular references.
Omniparser has template result caching, meaning, if there are multiple fields referencing the same template
at the same IDR tree cursor position, then the transform/computation result for the first field will be
cached and used in the second and subsequent fields. Because of this, using templates is in fact recommended
for readability and reusability. Below is a skeleton `transform_declarations`:
```
"transform_declarations": {
    "FINAL_OUTPUT": { "object": {
        "field1": { ... },
        "field2": { "template": "template2" },
        "field3": { "object": {
            "field4": { ... },
            "field5": { "template": "template2" }
        }}
    }},
    "template2": { "object": {
        "field6": { "template": "template3" }
        ...
    }},
    "template3": { "object": {
        ...
    }},
}
```
Note even though in the example/skeleton above, all the `template?` templates are of `object` transform, a
template can in fact be of any transform type, which we'll cover next.

## Transform Types

We have the following transform types in `omni.2.1` schema version:
- Direct field extraction (or in short, **field**): e.g. `{ "xpath": "<xpath query to the field value>" }`
(or use `xpath_dynamic`). This transform directive extracts the text/string value from the IDR tree node
matched by the XPath query. As we mentioned earlier in [XPath](./xpath.md) doc that field transform's `xpath`
or `xpath_dynamic` result set must yield zero or one IDR node, or the current transform will fail.

- Constant (or in short, **const**): e.g. `{ "const": "this is a constant" }`.

- External value (or in short, **external**): e.g. `{ "external": "<external var name>" }`. This is a
transform that looks up a value by the provided name in the external values passed to omniparser inside
[`transformctx.Ctx`](../transformctx/ctx.go) during [`NewTransform(...)`](../transform.go) creation time.
For example, if we want the output to include the input file name, we can author the schema like this:
    ```
    "transform_declarations": { "object: {
        ...
        "input_name": { "external": "input_name" },
        ...
    }}
    ```
    We call `NewTransform` like this:
    ```
    transform, err := NewTransform(?, ?,
        &transformctx.Ctx{
            ExternalProperties: map[string]string {
                "input_name": <the actual input file path string>,
            }})
    ```

- Object (**object**): e.g. `{ "object" : {...} }`. This transform directive tells omniparser an object
definition and structure is needed here. Note that even though vast majority of schemas use `object`
transform directive for `FINAL_OUTPUT`, it is not actually required. `FINAL_OUTPUT` can be of any transform
type.

- Array (**array**): e.g. `{ "array": [ {...}, {...}, ... ] }`. Inside the `[]` of an `array` transform
directive there can be zero, or one, or more transform directives of any type. Let's take a look at a few
examples:
    - Example 1:
        ```
        "field": { "array": [] }
        ```
        `"field"` in the output will be an empty JSON array. (Actually `"field"` will be omitted from the
        output, unless `"keep_empty_or_null"` flag is specified, which we will cover later in
        [Miscellaneous](#miscellaneous).)

    - Example 2:
        ```
        "field": { "array": [
            { "const": "Sunday" },
            { "const": "Monday" },
            ...
            { "const": "Saturday" }
        ]}
        ```
        `"field"` in the output will be a JSON array containing week day string literals.

    - Example 3:
        ```
        "field": { "array": [
            { "const": "3.1415", "type": "float" },
            { "const": "true", "type": "boolean" },
        ]}
        ```
        `"field"` in the output will be a JSON array containing numeric value of pi and a boolean value.

    - Example 4:
        ```
        "field": { "array": [ { "xpath": "books/*/title" } ] }
        ```
        `"field"` in the output will be a JSON array containing all the book titles resulted from query.

    - Example 5:
        ```
        "field": { "array": [
            { "xpath": "cookbooks/*/title" },
            { "const": "---" },
            { "xpath": "mathbooks/*/title" }
        ]}
        ```
        `"field"` in the output will be a JSON array containing all the cook book titles, a string literal
        `---`, and all the math book titles.

    - Example 6:
        ```
        "field": { "array": [
            { "object": {...} },
            { "template": "template1" },
            { "custom_func": {...} }
        ]}
        ```
        `"field"` in the output will be a JSON array containing an object, the result from `template1`
        (whatever the result type is, be it a string, a numeric value, or an object, even), and the result
        from a `custom_func` invocation.

- Template (**template**): e.g. `{ "template": "<template name>" }`.

- Custom Function Call (**custom_func**): e.g. `{ "custom_func": {...} }`. See more details about
`custom_func` transform directive [here](./use_of_custom_funcs.md).

- Custom Parse (**custom_parse**): e.g. `{ "custom_parse": "<custom parse id>" }`. See more details about
`custom_parse` transform directive [here](./programmability.md).

## Miscellaneous

Several attributes can be specified on some or all transform directives:

1. `xpath` (or `xpath_dynamic`) can be used for data extraction or IDR cursor anchoring with the following
transform types: field (in fact field has nothing else but an `xpath` or `xpath_dynamic`), `object`,
`template`, `custom_func` and `custom_parse`. See more details about use of `xpath` (or `xpath_dynamic`)
[here](./xpath.md).

2. `type` tells omniparser the result from the transform needs a type cast. Supported type cast types are:
`int`, `float`, `boolean`, and `string`. Not specifying `type` means keep whatever the result type the
transform yields. Note type casting is only allowed when the result type from a transform is of primitive
type, such as integer, float, bool, and string, or a (non-fatal) parser error will be raised and the
transform for the current record will be abandoned.

3. `no_trim` tells omniparser not to trim the leading and trailing white spaces, if the transform result
type is string. It has no effect if the result type is not a string. Omniparser will by default trim any
leading and trailing spaces for a string typed field. Sometimes we simply want to preserve white spaces:
    ```
    "field": { "custom_func": {
        "name": "concat",
        "args": [
            { "xpath": "./FIRST_NAME" },
            { "const": " ", "no_trim": true },
            { "xpath": "./LAST_NAME" }
        ]
    }}
    ```
    We have to specify `no_trim` or omniparser will auto trim the white space constant, and the result will
    be first name concatenated with last name without a space delimiter.

4. `keep_empty_or_null` tells omniparser not to omit the value output even if it is null (for object, array)
or empty (for string). Omniparser will by default omit any value that is null or empty. Sometimes, we simply
want a result to be included in the output whether it's null/empty or not:
    ```
    "field": { "keep_empty_or_null": true, "object": {
        ...
    }}
    ```
    If for some reason, the object result is null, the output will still have this: `"field": {}`.

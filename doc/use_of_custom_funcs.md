# Use of `custom_func`, Specially `javascript`

`custom_func` is a transform that allows schema writer to alter, compose, transform and aggregate existing
data from the input. Among [all `custom_func`](./customfuncs.md),
[`javascript`](./customfuncs.md#javascript) is the most important one to understand and master.

A `custom_func` has 4 basic parts: `xpath`/`xpath_dynamic`, `name`, `args`, and `type`.

Like any other transforms, `custom_func` uses optional `xpath`/`xpath_dynamic` directive to move the current
IDR tree cursor. See [here](xpath.md#data-context-and-anchoring) for more details.

`name` is self-explanatory.

`args` is a list of arguments, which themselves are transforms recursively, to the function.

Optional `type` indicates a result type cast is needed. Valid types are `'string'`, `'int'`, `'float'`,
and `'boolean'`. Not specifying `type` tells omniparser to keep whatever type of the result from the
`custom_func` as is.

## Basic Examples

1. Fixed Argument List

    Look at the following transform example:
    ```
    "carrier": { "custom_func": { "name": "lower", "args": [ { "xpath": "./CARRIER_NAME" } ] } },
    ```
    This transform, in English, takes the value of the immediate child node `CARRIER_NAME` from the current
    IDR tree cursor position, and returns it in lower-case.

2. Variable Argument List

    Look at the following transform example (adapted from
    [here](../extensions/omniv21/samples/fixedlength/2_multi_rows.schema.json)):
    ```
    "event_datetime": { "custom_func": {
        "name": "concat",
        "args": [
            { "xpath": "event_date" },
            { "const": "T" },
            { "xpath": "event_time" }
        ]
    }},
    ```
    This transform, in English, takes the values of a child node `event_date`, a constant string `T` and a
    child node `event_time`, and returns them concatenated.

3. Chaining/Composability

    Arguments of a `custom_func` transform can also be `custom_func`, thus enabling chaining and
    composability. Look at the following example (adapted from
    [here](../extensions/omniv21/samples/fixedlength/2_multi_rows.schema.json)):
    ```
    "event_date_template": { "custom_func": {
        "name": "dateTimeToRFC3339",
        "args": [
            { "custom_func": {
                "name": "concat",
                "args": [
                    { "xpath": "event_date" },
                    { "const": "T" },
                    { "xpath": "event_time" }
                ]
            }},
            { "xpath": "event_timezone", "_comment": "input timezone" },
            { "const": "", "_comment": "output timezone" }
        ]
    }}
    ```
    This transform, in English, concatenates child nodes to produce a full event datetime string and then
    use `dateTimeToRFC3339` to normalize the datetime string into RFC3339 standard format.

    There is no limit on how deep `custom_func` chaining can be.

4. `xpath`/`xpath_dynamic` Anchoring

    Schema writer can also use `xpath` (or `xpath_dynamic`) to change current IDR tree cursor to make
    data extractions on arguments easier. Consider the same transform as above but imagine this time
    the all the event date time related fields are not at the current IDR cursor node, but rather in a
    child node `data`. Instead of writing each data extract `xpath` in the arguments as `"data/..."`, we
    can simply move the cursor to `data`, by specifying `xpath` on `custom_func` itself.
    ```
    "event_date_template": { "xpath": "data", "custom_func": {
        "name": "dateTimeToRFC3339",
        "args": [
            { "custom_func": {
                "name": "concat",
                "args": [
                    { "xpath": "event_date" },
                    { "const": "T" },
                    { "xpath": "event_time" }
                ]
            }},
            { "xpath": "event_timezone", "_comment": "input timezone" },
            { "const": "", "_comment": "output timezone" }
        ]
    }}
    ```

## `javascript` and `javascript_with_context`

Omniparser has several basic `custom_func` like `lower`, `upper`, `dateTimeToRFC3339`, `uuidv3`, etc, among
which the most important, flexible and powerful one is `javascript` (and its sibling
`javascript_with_context`).

`javascript` is a `custom_func` transform that executes a JavaScript with optional input arguments.
Omniparser uses https://github.com/dop251/goja, a native Golang ECMAScript implementation thus **free of
external C/C++ lib dependencies**.

A simple example (adapted from [here](../extensions/omniv21/samples/csv/1_weather_data_csv.schema.json)):
```
"temp_in_f": { "custom_func": {
    "name": "javascript",
    "args": [
        { "const": "Math.floor((temp_c * 9 / 5 + 32) * 10) / 10" },
        { "const": "temp_c" }, { "xpath": ".", "type": "float" }
    ]
}}
```
This transform takes the value of the current IDR node, assuming temperature data in celsius, converts
it to fahrenheit.

The first argument is typically a `const` transform that contains a javascript code. The rest of the
arguments always come in pairs. In each pair, the first argument specify an input argument name, and the
second specifies the value of the argument. Remember chaining is allowed for advanced composability.

The result type is whatever the type the script return value is, unless schema writer adds a `type` cast
in the `custom_func` transform to force a type conversion.

If there is any exception thrown in the script, `javascript` transform will fail with an error. If the
result from the script is `NaN`, `null`, `Infinity` or `Undefined`, the transform will fail with an error.

Another example (adapted from [here](../extensions/omniv21/samples/csv/1_weather_data_csv.schema.json)):
```
"uv_index": { "custom_func": {
    "name": "javascript",
    "args": [
        { "const":  "uv.split('/').map(function(s){return s.trim();}).filter(function(s){return !!s;})" },
        { "const": "uv" }, { "xpath": "UV_INDEX" }
    ]
}},
```
where `UV_INDEX` column contains text like `"12/4/6"`.

The script above splits the input by `'/'`, trims away spaces, tosses out empty ones and returns it
as an array, so the result for `"uv_index"` in the output JSON would look like this:
```
"uv_index": [
    "12",
    "4",
    "6"
],
```

So far the input arguments in the samples above are all of singular value. We can also support input
argument of array, thus enabling aggregation (from
[here](../extensions/omniv21/samples/json/2_multiple_objects.schema.json)):
```
"sum_price_times_10": { "custom_func": {
    "name": "javascript",
    "args": [
        { "const": "t=0; for (i=0; i<prices.length; i++) { t+=prices[i]*10; } Math.floor(t*100)/100;" },
        { "const": "prices" }, { "array": [ { "xpath": "books/*/price", "type": "float" } ] }
    ]
}},
```
Contrived, this transform takes all the price values from `"books/*/price"` XPath query, inflates each
by 10 (why oh why?! :)), sums them all up, and returns the sum with 2 decimal places.

Input arguments to `javascript` function can be of simple primitive types (such as string, numbers, etc)
but also objects or arrays, as illustrated above.

To provide ultimate freedom of parsing and transform, `javascript` has an even more powerful sibling
function `javascript_with_context`. `javascript_with_context` is very similar to `javascript`, except that
omniparser automatically injects the current IDR node and its sub-tree as a JSON object into the script
under the global variable name `_node`, thus allowing the script to parse, and transform the current
IDR node tree as it see fit. (You may ask why not just have `javascript` and auto-inject `_node`? It
is because converting IDR node tree to JSON isn't exactly cheap and for vast majority cases, `_node`
isn't needed so `javascript` is perfectly sufficient.)

Consider the following example:
```
"full_name": { "xpath": "./personal_info", "custom_func": {
    "name": "javascript_with_context",
    "args": [
        { "const": "var n = JSON.parse(_node); n.['Last Name'] + ', ' + n.['First Name']" }
    ]
}}
```
assuming the current IDR context for this `"full_name"` transform is:
```
Node(Type: ElementNode)
    Node(Type: ElementNode, Data: "First Name")
        Node(Type: TextNode, Data: "John")
    Node(Type: ElementNode, Data: "Last Name")
        Node(Type: TextNode, Data: "Doe")
    Node(Type: ElementNode, Data: "Age")
        Node(Type: TextNode, Data: "35")
```

When `javascript_with_context` is invoked, omniparser will convert the IDR tree above into a JSON object:
```
{
    "First Name": "John",
    "Last Name": "Doe",
    "Age": "35"
}
```
thus allowing the script to parse the JSON object in and do something about it.

Theoretically, the entire `FINAL_OUTPUT` transform can be done with `javascript_with_context`. However,
the cost/con of doing so or similarly "large-scale" `javascript_with_context` is 1) multiple round trips
of serializing IDR into JSON then parsing JSON into javascript object and 2) it's just hard to write that
much javascript in one line -- the current limitation of schema being strictly JSON which doesn't support
multi-line string literals.

## Error Handling

If any of the argument tranforms return error, or the custom function itself fails, an error will be
relayed out, unless `ignore_error` is specified.

Look at the following example (adapted from
[here](../extensions/omniv21/samples/fixedlength/2_multi_rows.schema.json)):
```
"event_date_template": { "custom_func": {
    "name": "dateTimeToRFC3339",
    "args": [
        { "custom_func": {
            "name": "concat",
            "args": [
                { "xpath": "event_date" },
                { "const": "T" },
                { "xpath": "event_time" }
            ]
        }},
        { "xpath": "event_timezone", "_comment": "input timezone" },
        { "const": "", "_comment": "output timezone" }
    ],
    "ignore_error": true
}}
```

If say the `event_date` and `event_time` contain invalid characters, and `dateTimeToRFC3339` would
typically fail to convert it to RFC3339 standard format, thus failing out the transform of
`FINAL_OUTPUT` for the current record. However, because of `"ignore_error": true`, instead, this
`custom_func` would simply return `nil/null` without error.

If an argument transform value is `nil/null` (possibly due to argument transform failure coupled with
its own `"ignore_error": true`), then this argument's value will be whatever the default value of
the argument type dictates, such as `0` for `int`, `0.0` for `float`, `""` for `string`, etc.

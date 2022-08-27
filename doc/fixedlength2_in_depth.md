- [Fixed-Length (`fixedlength2`) Schema in Depth](#fixed-length---fixedlength2---schema-in-depth)
  * [`envelope` Basics](#-envelope--basics)
  * [`envelope`s Hierarchy](#-envelope-s-hierarchy)
  * [`file_declaration` for `fixedlength2`](#-file-declaration--for--fixedlength2-)
  * [Sample 1: `file_declaration` for Repeated Single-Row `envelope`](#sample-1---file-declaration--for-repeated-single-row--envelope-)
  * [Sample 2: `file_declaration` for Repeated Multi Fixed-Number-of-Rows `envelope`](#sample-2---file-declaration--for-repeated-multi-fixed-number-of-rows--envelope-)
  * [Sample 3: `file_declaration` for Repeated Variable Length `envelope` Bounded by `header`/`footer`](#sample-3---file-declaration--for-repeated-variable-length--envelope--bounded-by--header---footer-)
  * [Sample 4: `file_declaration` for Nested Hierarchical `envelope`s with Different Types](#sample-4---file-declaration--for-nested-hierarchical--envelope-s-with-different-types)
  * [Fixed-Length IDR Structure](#fixed-length-idr-structure)
  * [Migration from `fixed-length` Schemas](#migration-from--fixed-length--schemas)

# Fixed-Length (`fixedlength2`) Schema in Depth

> Note: With the introduction of `fixedlength2` schema, the old `fixed-length` schema is now
> considered **_deprecated_**. It will continue to work, and its [documentation](./fixedlength_in_depth.md)
> remains but it will not be primarily linked from the homepage [README.md](../README.md). There
> is a [migration](#migration-from--fixed-length--schemas) section near the end of this page to
illustrate the simple schema migration steps from `fixed-length` to `fixedlength2`.

Fixed-length (sometimes also called fixed-width) schema has 3 parts: `parser_settings`, `file_declaration`,
and `transform_declarations`. We've covered `parser_settings` in [Getting Started](./gettingstarted.md);
we've covered `transform_declarations` in depth in [All About Transforms](./transforms.md). Before we go
into `file_declaration`, we need to talk about the concept of `envelope`.

## `envelope` Basics
An `envelope` is a basic data unit ingested in from the input, processed and transformed by omniparser for
a fixed length input file. Here are a few examples:

- Single line `envelope` (from [this sample](../extensions/omniv21/samples/fixedlength2/1_single_row.input.txt)):
    ```
    2019/01/31T12:34:56-0800 10.5 30.2  N 33  37.7749 122.4194
    2020/07/31T01:23:45-0500   39   95 SE  8  32.7767  96.7970
    ```
    Here each line in the input is an `envelope`, individually ingested, processed and transformed.

- Multi-line `envelope` (from [this sample](../extensions/omniv21/samples/fixedlength2/2_multi_rows.input.txt)):
    ```
    H001TW                  0689311345         030                           DEL                 ...
    H0020689311345         BM  040670500067120
    H0020689311345         CN  100000103732
    H0020689311345         OA  2400 Highway 155 South
    H0020689311345         OC  Locust Grove
    H0020689311345         ON  THE HOME DEPOT 6705
    H0020689311345         OO  W841206858
    H0020689311345         OP  30248
    H0020689311345         OST GA
    H0020689311345         OT  US
    H0020689311345         PO  7029083C
    H001TW                  0689311348         030                           DEL                 ...
    H0020689311348         BM  040677700277714
    H0020689311348         CN
    H0020689311348         OA  2500 Highway 155 South
    H0020689311348         OC  Locust Grove
    H0020689311348         ON  THE HOME DEPOT 6777
    H0020689311348         OO  W938003272
    H0020689311348         OP  30248
    H0020689311348         OST GA
    H0020689311348         OT  US
    H0020689311348         PO  6816137C
    ```
    Each `envelope` consists of 11 consecutive lines, and these 11 consecutive lines are ingested, processed and
    transformed as a unit.

    Single line `envelope` is obviously a special case of multi-line `envelope`.

- Variable length `envelope` encapsulated by a `header` and a `footer` (from
[this sample](../extensions/omniv21/samples/fixedlength2/3_header_footer.input.txt)):
    ```
    A010 20191105
    ...
    ...
    A999
    V010 V
    ...
    ...
    V999
    V010 V
    ...
    ...
    V999
    ...
    ...
    Z001
    Z999
    ```
    Here we have 3 different `envelope`s:
    * first `envelope` is encapsulated by `A010` and `A999`. This envelope has one occurrence.
    * second `envelope` is encapsulated by `V010` and `V999`. This envelope has multiple occurrences.
    * last `envelope` is encapsulated by `Z001` and `Z999`. This envelope has one occurrence.

    Note if the `footer` is the same as the `header`, such `envelope` will consist of only one line.
    In such case, `footer` can be omitted for brevity of schema writing.

## `envelope`s Hierarchy

In most use cases, a fixed-length input consists of a single-line `envelope` that repeats itself.
Sometimes, we encounter input that consists of multi-line (but fixed number of lines) `envelope`
that repeats itself. Occasionally, `envelope` with `header`/`footer`. But in some highly customized
scenarios, we do see some nested and mixed `envelope` situation:

```
<envelope1> {0,}
    <envelope2> {1,2}
    <envelope3> {0,1}
```

where in the pseudo example above, `envelope1` is followed by `envelope2` 1 or 2 times, then
followed by `envelope3` 0 or 1 time, then repeats itself (`envelope1`). The following is adapted and
simplified from the [nested sample](../extensions/omniv21/samples/fixedlength2/4_nested.input.txt):
```
HDR...
GRH...

NWR...
SPU...
SPT...
SPT...

NWR...
SPU...
SPU...
SPT...

GRT...
TRL...
```

While this is a flattened text file with fixed length data encoded within, it, however, represents
a hierarchical and nested data structure:
```
HDR
    GRH
        NWR
            SPU
                SPT
    GRT
TRL
```
- There is 1 and only 1 `HDR` in the input.
- `HDR` is paired with `TRL`.
- Underneath `HDR`, there could be 0 or many repeats of `GRH`.
- Each `GRH` is paired with a `GRT`.
- Underneath `GRT`, there could be 0 or many repeats of `NWR`.
- There is not closing pair, so to speak, for each of `NWR`.
- Underneath `NWR`, there could be 0 or many repeats of `SPU`.
- Underneath `SPU`, there could be 0 or many repeats of `SPT`.

If this nested structure embedded in a flat txt file reminds you of [EDI](./edi_in_depth.md), you
are not alone :). In the wild business world, there are often such blend of things created: half
EDI (with the concepts of hierarchy, loops, min/max occurrences, etc) half fixed-length with all
the data pieces are in fixed locations without use of any delimiters.

## `file_declaration` for `fixedlength2`

```
"file_declaration": {
    "envelopes": [                                   <= required
        {
            "name": <envelope name>,                 <= optional
            "rows": <integer value>,                 <= optional
            "header": <regex>,                       <= optional
            "footer": <regex>,                       <= optional
            "type": <envelope|envelope_group>,       <= optional
            "is_target": <true|false>,               <= optional
            "min": <integer value>,                  <= optional
            "max": <integer value>,                  <= optional
            "columns": [                             <= optional
                {
                    "name": "<column name>",         <= optional
                    "start_pos": <integer>,          <= required
                    "length": <integer>,             <= required
                    "line_index": "<integer>",       <= optional
                    "line_pattern": "<line regexp>"  <= optional
                },
                <more columns>
            ],
            "child_envelopes": [                     <= optional
                <more envelopes>
            ]
        }
    ]
}
```

- `name`: the name of the `envelope` is used in `xpath` query in `transform_declarations`. It is
optional in simple use cases where the `envelope` name isn't needed in any of the transform's xpath
query, such as in this [example](../extensions/omniv21/samples/fixedlength2/1_single_row.schema.json).

- `rows`: the fixed number of the rows in this envelope. If `rows` specified, `header`/`footer` must
not be used; If none of `rows`, `header`, or `footer` is specified, the `envelope` is defaulted to
being a rows based `envelope` with `rows` equals 1.

- `header`/`footer`: the regex pattern matches the starting and ending lines of a variable-length
`envelope`. `rows` must not be used when `header`/`footer` are used. If `footer` is omitted, it
defaults to be the value of `header`, which means an `envelope` with only `header` specified will
only match a single line for the envelope.

- `type`: specifies whether an `envelope` is a sold `envelope` or an `envelope_group` which serves
as a container for child `envelope`s. If `type` is `envelope_group`, `rows`/`header`/`footer` must
not be used, as they are only relevant to solid/concrete `envelope`. Also `envelope_group` cannot
co-exists with `columns` for the same reason.

- `is_target`: specifies if an `envelope` is a streaming target. `is_target` concept here is the
same as in [EDI](./edi_in_depth.md). Only one `envelope` (group or not) can have `is_target: true`.
For ease and brevity of schema writing, if no `is_target: true` is specified, then the first
`envelope` will be auto-marked as `is_target: true`.

- `min`: specifies the minimum number of consecutive occurrences of the `envelope` in the data
stream. `min` can be specified on `envelope_group`. If omitted, `min` defaults to 0.

- `max`: specifies the maximum number of consecutive occurrences of the `envelope` in the data
stream. `max` can be specified on `envelope_group`. If omitted, `max` defaults to -1, which means
unlimited.

- `columns`: specifies a collection of columns of the `envelope`.
- `column.start_pos`: specifies the starting character position (1-based) for the column's data.
- `column.length`: specifies the length of the column's data.
- `column.line_index`: used in multi-line `envelope` (`rows` based or `header`/`footer` based)
where the index indicates which line this column's data will be extracted from. 1-based.
- `column.line_pattern`: used in multi-line `envelope` (`rows` based or `header`/`footer` based)
where the pattern identifies which line this column's data will be extracted from.

- `child_envelopes`: specifies, recursively, any hierarchical and nested child envelope structure.

## Sample 1: `file_declaration` for Repeated Single-Row `envelope`

Full sample input is [here](../extensions/omniv21/samples/fixedlength2/1_single_row.input.txt).
Full sample schema is [here](../extensions/omniv21/samples/fixedlength2/1_single_row.schema.json).

```
    "file_declaration": {
        "envelopes" : [
            {
                "columns": [
                    { "name": "DATE", "start_pos": 1, "length": 24 },
                    { "name": "HIGH_TEMP_C", "start_pos": 26, "length": 4 },
                    { "name": "LOW_TEMP_F", "start_pos": 31, "length": 4 },
                    { "name": "WIND_DIR", "start_pos": 36, "length": 2 },
                    { "name": "WIND_SPEED_KMH", "start_pos": 9, "length": 2 },
                    { "name": "LAT", "start_pos": 42, "length": 8 },
                    { "name": "LONG", "start_pos": 51, "length": 8 }
                ]
            }
        ]
    }
```

Note, many of the `envelope` fields are omitted to use default values for simplicity and brevity:
- `rows`/`header`/`footer` omitted and defaulted to `rows:1`.
- `type` omitted and defaulted to `type:envelope`.
- `is_target` omitted and defaulted to `is_target:true`.
- `min` omitted and defaulted to `min:0`.
- `max` omitted and defaulted to `max:-1`.
- `column.line_index` omitted since each `envelope` contains only 1 line.
- `column.line_pattern` omitted since each `envelope` contains only 1 line.

## Sample 2: `file_declaration` for Repeated Multi Fixed-Number-of-Rows `envelope`

Full sample input is [here](../extensions/omniv21/samples/fixedlength2/2_multi_rows.input.txt).
Full sample schema is [here](../extensions/omniv21/samples/fixedlength2/2_multi_rows.schema.json).

```
    "file_declaration": {
        "envelopes" : [
            {
                "rows": 11,
                "columns": [
                    { "name": "tracking_number_h001", "start_pos": 464,  "length": 30, "line_pattern": "^H001" },
                    { "name": "destination_country", "start_pos": 607,  "length": 2, "line_pattern": "^H001" },
                    { "name": "guaranteed_delivery_date", "start_pos": 376,  "length": 8, "line_pattern": "^H001" },
                    { "name": "event_date", "start_pos": 142,  "length": 8, "line_pattern": "^H001" },
                    { "name": "event_time", "start_pos": 150, "length": 8, "line_pattern": "^H001" },
                    { "name": "event_timezone", "start_pos": 158, "length": 4, "line_pattern": "^H001" },
                    { "name": "event_city", "start_pos": 564,  "length": 30, "line_pattern": "^H001" },
                    { "name": "event_state", "start_pos": 594, "length": 2, "line_pattern": "^H001" },
                    { "name": "scan_facility_zip", "start_pos": 596, "length": 11, "line_pattern": "^H001" },
                    { "name": "tracking_number_h002_cn", "start_pos": 28,  "length": 50, "line_pattern": "^H002.{19}CN" }
                ]
            }
        ]
    }
```

## Sample 3: `file_declaration` for Repeated Variable Length `envelope` Bounded by `header`/`footer`

Full sample input is [here](../extensions/omniv21/samples/fixedlength2/3_header_footer.input.txt).
Full sample schema is [here](../extensions/omniv21/samples/fixedlength2/3_header_footer.schema.json).

```
    "file_declaration": {
        "envelopes" : [
            {
                "name": "HEADER", "min": 1, "max": 1, "header": "^A010", "footer": "^A999",
                "columns": [ { "name": "carrier", "start_pos": 6, "length": 6, "line_pattern": "^A060" } ]
            },
            {
                "name": "REC", "header": "^V010", "footer": "^V999", "is_target": true,
                "columns": [
                    { "name": "tracking_number", "start_pos": 6, "length": 15, "line_pattern": "^V020" },
                    { "name": "delivery_date", "start_pos": 6, "length": 8, "line_pattern": "^V045" },
                    { "name": "observation_type", "start_pos": 6, "length": 1, "line_pattern": "^V060" },
                    { "name": "reason_for_observation", "start_pos": 6, "length": 2, "line_pattern": "^V070" },
                    { "name": "date_observation", "start_pos": 6, "length": 8, "line_pattern": "^V080" },
                    { "name": "time_observation", "start_pos": 6, "length": 6, "line_pattern": "^V081" },
                    { "name": "weight_in_grams", "start_pos": 6, "length": 6, "line_pattern": "^V110" },
                    { "name": "postal_code_addressee", "start_pos": 6, "length": 6, "line_pattern": "^V160" },
                    { "name": "city_name_addressee", "start_pos": 6, "length": 24, "line_pattern": "^V180" },
                    { "name": "country_code_addressee", "start_pos": 6, "length": 2, "line_pattern": "^V200" }
                ]
            },
            { "name": "FOOTER",  "min": 1, "max": 1, "header": "^Z001", "footer": "^Z999" }
        ]
    }
```

Note the use of `is_target:true` on `envelope` named `REC`.

## Sample 4: `file_declaration` for Nested Hierarchical `envelope`s with Different Types

Full sample input is [here](../extensions/omniv21/samples/fixedlength2/4_nested.input.txt).
Full sample schema is [here](../extensions/omniv21/samples/fixedlength2/4_nested.schema.json).

```
    "file_declaration": {
        "envelopes": [
            {
                "name": "HDR", "header": "^HDR", "min": 1, "max": 1,
                "child_envelopes": [
                    {
                        "name": "GRH", "header": "^GRH",
                        "child_envelopes": [
                            {
                                "name": "NWR", "header": "^NWR", "is_target": true,
                                "columns": [ { "name": "title", "start_pos": 20, "length": 60 } ],
                                "child_envelopes": [
                                    {
                                        "name": "SPU", "header": "^SPU",
                                        "columns": [ { "name": "publisher_name", "start_pos": 31, "length": 45 } ],
                                        "child_envelopes": [
                                            {
                                                "name": "SPT", "header": "^SPT",
                                                "columns": [ { "name": "territory_id", "start_pos": 51, "length": 4 } ]
                                            }
                                        ]
                                    },
                                    {
                                        "name": "SWR", "header": "^SWR",
                                        "columns": [
                                            { "name": "writer_id", "start_pos": 20, "length": 9 },
                                            { "name": "last_name", "start_pos": 29, "length": 45 }
                                        ],
                                        "child_envelopes": [
                                            {
                                                "name": "SWT", "header": "^SWT",
                                                "columns": [ { "name": "territory_id", "start_pos": 45, "length": 4 } ]
                                            }
                                        ]
                                    }
                                ]
                            },
                            { "name": "GRT", "header": "^GRT" }
                        ]
                    },
                    { "name": "TRL", "header": "^TRL" }
                ]
            }
        ]
    }
```

The schema is lengthy but fairly straightforward:
- The input/schema has a singleton/global header `envelope` named `HDR`, it is paired with a trailer
`envelope` `TRL`. Therefore both `envelope`s specify `min`/`max` to 1.
- Inside the global singleton header/trailer, we have unlimited repeats of an `envelope` named `GRH`,
which is also paired with a trailer `GRT`.
- Under `GRH`, we have unlimited repeats of `NWR` which is our streaming and transform target.
- Each `NWR` can have two "loopy" children, `SPU` and `SWR`, each of which has unlimited repeats of
child `envelope`, `SPT` and `SWT`, respectively.
- Each envelope is a single line, thus no `line_index` or `line_pattern` is used.

## Fixed-Length IDR Structure

See [here](./idr.md#fixed-length-mostly-txt) for more details.

## Migration from `fixed-length` Schemas

If one looks at the documentation for the old `fixed-length` schema
[here](./fixedlength_in_depth.md), you notice `fixed-length` and `fixedlength2` schemas are really
similar! It certainly makes migration incredibly trivial:

- For [single-row `envelope` based schema](#sample-1---file-declaration--for-repeated-single-row--envelope-),
migrating from `fixed-length` schema to `fixedlength2` schema requires only 1 change:
    ```
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "fixed-length"   ===> change it to "fixedlength2", yes the "-" is gone.
    }
    ```

- For [multi-row `envelope` based schema](sample-2---file-declaration--for-repeated-multi-fixed-number-of-rows--envelope-),
migrating from `fixed-length` schema to `fixedlength2` schema requires 2 changes:
    - Change 1: the same `parser_settings.file_format_type` change mentioned above.
    - Change 2:
        ```
        "file_declaration": {
            "envelopes" : [
                {
                    "by_rows": 11,          ===> change it to "rows".
                    "columns": [
                        ...
                        ...
                    ]
                }
            ]
        },
        ```

- For [`header`/`footer` `envelope` based schema](#sample-3---file-declaration--for-repeated-variable-length--envelope--bounded-by--header---footer-),
migrating from `fixed-length` schema to `fixedlength2` schema requires 3 changes:
    - Change 1: the same `parser_settings.file_format_type` change mentioned above.
    - Change 2:
        ```
        "file_declaration": {
            "envelopes" : [
                {
                    "by_header_footer": { "header": "^A010", "footer": "^A999" },
                    ...
                }
            ]
        },
        ```
        Any mentioning of `"by_header_footer": { "header": "^A010", "footer": "^A999" }` changes to
        `"header": "^A010", "footer": "^A999"`. Basically remove the `by_header_footer` field, and
        move `header` and `footer` one level up:
        ```
        "file_declaration": {
            "envelopes" : [
                {
                    "header": "^A010", "footer": "^A999",
                    ...
                }
            ]
        },
        ```
    - Change 3:
        ```
        "file_declaration": {
            "envelopes" : [
                {
                    "not_target": true,
                    ...
                }
                {
                    ...
                }
            ]
        },
        ```
        Any mentioning of `"not_target": true`, remove that line.
        Any `envelope` that doesn't have `"not_target": true` (there should be one and only one of
        such `envelope`), add a line `"is_target": true` to it:
        ```
        "file_declaration": {
            "envelopes" : [
                {
                    ...
                }
                {
                    "is_target": true,
                    ...
                }
            ]
        },
        ```

That's all the migration changes needed!

# Fixed-Length Schema in Depth (**DEPRECATED**)

> Note: this version of `fixed-length` schema has been deprecated and superseded by
> [`fixedlength2`](./fixedlength2_in_depth.md) schema. Its functionality and support
> will continue but (incredibly easy)
> [migration to `fixedlength2`](./fixedlength2_in_depth.md#migration-from-fixed-length-schemas)
> is strongly recommended.
>

Fixed-length (sometimes also called fixed-width) schema has 3 parts: `parser_settings`, `file_declaration`,
and `transform_declarations`. We've covered `parser_settings` in [Getting Started](./gettingstarted.md);
we've covered `transform_declarations` in depth in [All About Transforms](./transforms.md). Before we go
into `file_declaration`, we need to talk about the concept of `envelope`.

An `envelope` is a basic data unit ingested in from the input, processed and transformed by omniparser for
a fixed length input file. Here are a few examples:

- Single line `envelope` (from [this sample](../extensions/omniv21/samples/fixedlength/1_single_row.input.txt)):
    ```
    2019/01/31T12:34:56-0800 10.5 30.2  N 33  37.7749 122.4194
    2020/07/31T01:23:45-0500   39   95 SE  8  32.7767  96.7970
    ```
    Here each line in the input is an `envelope`, individually ingested, processed and transformed.

- Multi-line `envelope` (from [this sample](../extensions/omniv21/samples/fixedlength/2_multi_rows.input.txt)):
    ```
    H001TW                  0689311345         030                           DEL       HMDEPOT                                                   20190826124704  EST                                                   1230004321         W841206858                                                                                                         James Bond                    20190827                                                  RHONDA                        W841206858                    128 Bird Ave                                                          HAPPYVALLEY                   FL54321      USELGS               030                                                     1                                       105.00   THDNR     1    1003164515          DLB000977979                                      Appt set between 00 00 AND 00 00 on 08 27 19                                                                                                                                                                                                                                                                AG                                                                                      21200    1600    7029083C                                                                                    314 Algebra Blvd                                            MEDIAN              FL31415  US
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
    H001TW                  0689311348         030                           DEL       HMDEPOT                                                   20190826124704  EST                                                   1230001234         W938003272                                                                                                         Jason Bourne                  20190827                                                  RHONDA                        W938003272                    123 S 45ST ST                                                         MAGIC BEACH                   FL12345      USELGS               030                                                     1                                       1.00     THDNR     1    1003621621                                                            Appt set between 00 00 AND 00 00 on 08 27 19                                                                                                                                                                                                                                                                AG                                                                                      21030    1430    6816137C                                                                                    314 Algebra Blvd                                            MEDIAN              FL31415  US
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

    Single line `envelope` is obviously a special case of multi-line envelope.

- Variable length `envelope` encapsulated by a `header` and a `footer` (from
[this sample](../extensions/omniv21/samples/fixedlength/3_header_footer.input.txt)):
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
    Here we have 3 different `envelopes`:
    - first `envelope` is encapsulated by `A010` and `A999`. This envelope has one occurrence.
    - second `envelope` is encapsulated by `V010` and `V999`. This envelope has multiple occurrences.
    - last `envelope` is encapsulated by `Z001` and `Z999`. This envelope has one occurrence.

## `file_declaration` for Fixed Rows `envelope`

`file_declaration` section looks as follows when dealing with envelopes of fixed number of rows:
```
"file_declaration": {
    "envelopes": [                              <= required
        {
            "by_rows": <integer value>,             <= optional
            "columns": [                            <= optional
                {
                    "name": "<unique column name>", <= optional
                    "start_pos": <integer>,         <= required
                    "length": <integer>,            <= required
                    "line_pattern": "<line regexp>" <= optional
                },
                ...
            ]
        }
    ]
}
```
Note **for `by_rows` fixed-length schema, there must be one and only one envelope**.

- `by_rows`: defines how many consecutive lines form an `envelope`. If not specified, default to 1.

- `columns`: defines a number of columns/fields that will hold speific pieces of data extracted from
those rows.

- `columns.name`: name of a column, used in as IDR node name and later in XPath queries in
`transform_declarations`. Required and must be unique.

- `columns.start_pos`: specifies the starting character position (1-based) for the column's data.

- `columns.length`: specifies the length of the column's data.

- `columns.line_pattern`: used in multi-line `envelope` where the pattern identifies from which line
this column's data will be extracted.

An example of single-line envelope `file_declaration` might look like:
```
"file_declaration": {
    "envelopes" : [
        {
            "columns": [
                { "name": "DATE", "start_pos": 1, "length": 24 },
                { "name": "HIGH_TEMP_C", "start_pos": 26, "length": 4 },
                ...
            ]
        }
    ]
},
```

An example of multi-line envelope `file_declaration` might look like:
```
"file_declaration": {
    "envelopes" : [
        {
            "by_rows": 11,
            "columns": [
                { "name": "tracking_number_h001", "start_pos": 464,  "length": 30, "line_pattern": "^H001" },
                ...
                { "name": "tracking_number_h002_cn", "start_pos": 28,  "length": 50, "line_pattern": "^H002.{19}CN" }
            ]
        }
    ]
},
```

## `file_declaration` for Header/Footer `envelope`

`file_declaration` section looks as follows when dealing with envelopes bounded by header/footer:
```
"file_declaration": {
    "envelopes": [                              <= required
        {
            "name": "<unique envelope name>",       <= optional
            "by_header_footer": {                   <= required
                "header": "<header regexp>",        <= required
                "footer": "<footer regexp>",        <= required
            },
            "not_target": <true/false>,             <= optional
            "columns": [                            <= optional
                {
                    "name": "<unique column name>", <= required
                    "start_pos": <integer>,         <= required
                    "length": <integer>,            <= required
                    "line_pattern": "<line regexp>" <= optional
                },
                ...
            ]
        },
        ...
    ]
}
```
Note there can be multiple envelopes for a fixed-length schema when dealing with envelopes bounded by
headers/footers.

Contrary to `by_rows` envelope and schema which is simple and easy to understand, `by_header_footer`
envelopes and schema needs a bit explanation before diving into each schema attribute.

**A`by_header_footer` fixed-length schema can have multiple of envelopes**. These envelopes' order
must match their appearance order in the input files, although each one is optional. (We could've
made out-of-order envelope matching possible, but based on experiences, such scenarios rarely exist
thus not worth pursuing at the cost of complexity.) Of these envelopes, **one and only one must be
the target envelope**, instance of which omniparser will ingest, transform and return. All other
non-target envelopes are considered global envelopes. Each envelope can have 0 or more instances
from the input. All global envelopes' instances are permanently kept in the IDR tree while target
envelope instance is kept in the returned IDR tree until the next `Read()` call is invoked by the
client, thus making stream-processing large files without memory constraints possible.

Typically "global" envelopes are for the global header and footer for an input, and usually their
numbers of instances are limited (1 usually). Target envelope is for the repeating data blocks of
an input. If you look at the example shown early under *"Variable length `envelope` encapsulated by
a `header` and a `footer`"*, you can see the global header envelope is wrapped by `A010` and `A999`;
the global footer envelope is wrapped by `Z001` and `Z999`; and repeating data block envelope is
wrapped by `V010` and `V999`.

Each `by_header_footer` envelope has a unique name: if the name is not directly given in the schema,
then it's randomly and uniquely generated. Because when an instance of the target envelope is
returned, the IDR tree node returned is anchored on the target envelope (see more details about
fixed-length IDR tree structure [here](./idr.md#fixed-length-mostly-txt)), thus XPath queries to any
data inside the target envelope don't need the node name. As a result, usually there is no need to
specify a name for the target envelope. If, however, there are data on the global envelopes that
transform needs, typically some data/info from the global header envelope, then such global envelopes
should be named, and transform can refer to such data by XPath queries like
`../<global_envelope_name>/<path_to_such-data>`. Understanding the IDR tree structure for fixed-length
format is the key to understand how you can extract data.

- `name`: a unique name of the envelope. Optional. If not specified, which is common, a unique id
is generated by omniparser.

- `by_header_footer.header`: a regexp pattern identifies the line of the beginning of the envelope.

- `by_header_footer.footer`: a regexp pattern identifies the line of the end of the envelope. Note
`footer` can be the same as the `header`, in case we have an envelope of a single line.

- `not_target`: whether the envelope is a target envelope or not. Optional, and by default it's false:
the envelope is a target envelope. Note there can be one and only one target envelope in a
`by_header_footer` fixed-length schema.

- `columns`: identical to `columns` in `by_rows` fixed_length schema.

A schema for the earlier example might look like this:
```
"file_declaration": {
    "envelopes" : [
        {
            "name": "GLOBAL",
            "by_header_footer": { "header": "^A010", "footer": "^A999" },
            "not_target": true,
            "columns": [ { "name": "carrier", "start_pos": 6, "length": 6, "line_pattern": "^A060" } ]
        },
        {
            "by_header_footer": { "header": "^V010", "footer": "^V999" },
            "columns": [
                { "name": "tracking_number", "start_pos": 6, "length": 15, "line_pattern": "^V020" },
                { "name": "delivery_date", "start_pos": 6, "length": 8, "line_pattern": "^V045" },
                ...
            ]
        },
        {
            "by_header_footer": { "header": "^Z001", "footer": "^Z999" },
            "not_target": true
        }
    ]
},
```

Note 1: since we want to have, in the `FINAL_OUTPUT`, a field `carrier`, and the carrier information
is in the global header envelope, thus we give a name to the global header envelope `GLOBAL` thus we
can refer to the carrier name data by XPath `../GLOBAL/carrier`.

Note 2: there is no data we need from global footer envelope, thus we can simply keep it "nameless" (
though technically it is not true, as omniparser will assign a random/unique name to it).

Note 3: target envelope is "nameless" and it is the only envelope does not have `"not_target": true`.

## Fixed-Length IDR Structure

See [here](./idr.md#fixed-length-mostly-txt) for more details.

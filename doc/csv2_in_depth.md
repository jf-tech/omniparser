- [CSV Schema in Depth](#csv-schema-in-depth)
  - [CSV `file_declaration`](#csv-file_declaration)
  - [CSV Specific IDR Structure](#csv-specific-idr-structure)
  - [Use Case: Simple CSV with No Header](#use-case-simple-csv-with-no-header)
  - [Use Case: Simple CSV with Header But Header Verification Not Needed](#use-case-simple-csv-with-header-but-header-verification-not-needed)
  - [Use Case: Simple CSV with Header And Header Verification Needed](#use-case-simple-csv-with-header-and-header-verification-needed)
  - [Use Case: CSV with Fixed Multi-row Records](#use-case-csv-with-fixed-multi-row-records)
  - [Use Case: CSV with Variable Multi-row Records Defined by Header and Footer](#use-case-csv-with-variable-multi-row-records-defined-by-header-and-footer)
  - [Use Case: CSV with Nested CSV Records](#use-case-csv-with-nested-csv-records)
  - [Migration from `'csv'` Schemas](#migration-from-csv-schemas)

# CSV Schema in Depth

> Note: With the introduction of `csv2` schema, the old `csv` schema is now considered
> **_deprecated_**. It will continue to work, and its [documentation](./csv_in_depth.md)
> remains but it will not be directly linked from the homepage [README.md](../README.md). There
> is a [migration](#migration-from-csv-schemas) section near the end of this page
> to illustrate the simple schema migration steps from `csv` to `csv2`.

CSV (comma separated values) schema covers any delimiter (so not just comma) based flat file format. A
complete `"omni.2.1"` CSV schema has 3 parts: `parser_settings`, `file_declaration`, and
`transform_declarations`. We've covered `parser_settings` in [Getting Started](./gettingstarted.md);
we've covered `transform_declarations` in depth in [All About Transforms](./transforms.md); we've covered
some parts of CSV `file_declarations` in [Getting Started](./gettingstarted.md), we'll go into more details
about it here.

## CSV `file_declaration`

Full `file_declaration` schema looks as follows:
```
"file_declaration": {
    "delimiter": "<delimiter>"                      <= required
    "replace_double_quotes": true/false,            <= optional
    "records": [
        {
            "name": <record name>,                  <= optional
            "rows": <integer value>,                <= optional
            "header": <regex>,                      <= optional
            "footer": <regex>,                      <= optional
            "type": <record|record_group>,          <= optional
            "is_target": <true|false>,              <= optional
            "min": <integer value>,                 <= optional
            "max": <integer value>,                 <= optional
            "columns": [                            <= optional
                {
                    "name": "<column name>",        <= optional
                    "index": <integer>,             <= optional
                    "line_index": "<integer>",      <= optional
                    "line_pattern": "<line regexp>" <= optional
                },
                <...more columns...>
            ],
            "child_records": [                     <= optional
                <...more records...>
            ]
        }
        <...more records...>
    ]
}
```

- `delimiter`: self-explanatory.
    - Note 1: the delimiter must be of a single character.
    - Note 2: the delimiter doesn't have to be limited to ASCII character(s), utf8 rune is supported.

- `replace_double_quotes`: omniparser will replace any occurrences of double quotes `"` with single
quotes `'`.

    While [CSV RFC](https://tools.ietf.org/html/rfc4180) clearly specifies the rules of using double
    quotes and how to escape them, (un)surprisingly many CSV data providers have bad implementations
    and unescaped double quotes are quite frequently seen and breaking CSV parsing. The worst offender
    is this:

    ```
    COLUMN_1|COLUMN_2|COLUMN_3
    ...
    data 1|"data 2 has a leading double quote" then some|data 3
    ...
    ```

    Unfortunately Golang's [CSV parser](https://golang.org/pkg/encoding/csv/#Reader) cannot (nor should
    it) deal with this situation (even if you set `LazyQuotes` to `true`): it would start gobbling
    everything after the leading double quote in `COLUMN_2` as string literals, even passing the delimiter
    `|`. Usually it would consume **many many** lines until it finally hits another leading quote by luck.

    Given asking data providers/partners to fix double quote escaping properly according to CSV RFC is
    usually mission impossible from our experience, we've added `replace_double_quotes` flag to do a
    (frankly quite harsh) double quote to single quote replacement - yes, the content is altered, but at
    least the parsing and transform will succeed and minor differences result is the least evil we can do.
    **Use it only as a last resort**.

- `records.*.name`: the name of a record. Most the time there is no need to specify it, unless your
record name appears in some transformation XPath query.

- `records.*.rows`: defines the fixed number of rows of a record. If not specified, **and** `header`
is not specified, then it is defaulted to 1, which is vast majority of the case: a single row record.

- `records.*.header`: defines the regexp pattern to match the first line of a record.

- `records.*.footer`: defines the regexp pattern to match the last line of a record. If not specified,
(and assuming `header` is specified) then `header` alone can/will match a single line for the record.

- `records.*.type`: specifies whether an `record` is a sold `record` or a `record_group` which serves
as a container for child `record`s. If `type` is `record_group`, `rows`/`header`/`footer` must
not be used, as they are only relevant to solid/concrete `record`. Also `record_group` cannot
co-exists with `columns` for the same reason.

- `records.*.is_target`: specifies whether the record is our ingestion and transform target or not.
There can only be one record with `is_target=true`; if no record has explicit declaration of
`is_target`, then the first record will be defaulted to `is_target`.

- `records.*.min`: specifies the minimum number of consecutive occurrences of the `record` in the
data stream. `min` can be specified on `record_group` as well. If omitted, `min` defaults to 0.

- `records.*.max`: specifies the maximum number of consecutive occurrences of the `record` in the
data stream. `max` can be specified on `record_group` as well. If omitted, `max` defaults to -1,
which means unlimited.

- `records.*.columns`: specifies a collection of columns of the `record`.

- `records.*.columns.*.name`: specifies the name of the column.

- `records.*.columns.*.index`: specifies from which data column the value will be extract. If omitted,
it will have a default value of previous `column.index + 1`. If the first `column.index` isn't specified,
it will default to 1.

- `records.*.columns.*.line_index`: used in multi-line `record` (`rows` based or `header`/`footer` based)
where the index indicates which line this column's data will be extracted from. 1-based.

- `records.*.columns.*.line_pattern`: used in multi-line `record` (`rows` based or `header`/`footer` based)
where the pattern identifies which line this column's data will be extracted from.

- `records.*.child_records`: specifies, recursively, any hierarchical and nested child record structure.

## CSV Specific IDR Structure

See [here](./idr.md#csv-aka-delimited) for more details.

## Use Case: Simple CSV with No Header

Sample input:
```
1,2,3,4
A,B,C,D
...
```

This is one of most common use cases and for it, the `file_declaration` part of the schema is really
easy:

```
  "file_declaration": {
      "delimiter": ",",
      "records": [{ "columns": [
          { "name": "COL1" },
          { "name": "COL2" },
          { "name": "COL3" },
          { "name": "COL4" }
      ]}]
  },
```

Note CSV reader will always discard any empty lines, so if there is any before any data lines begin
or sprinkled in between data lines, no worries.

## Use Case: Simple CSV with Header But Header Verification Not Needed

Sample input:
```
C1,C2,C3,C4
1, 2, 3, 4
A, B, C, D
...
```

It has a header, but sometimes (based on our past experiences) we simply don't care about the header
row. So just to "skip" it:

```
  "file_declaration": {
      "delimiter": ",",
      "records": [
          { "rows": 1, "min": 1, "max": 1 },
          {
              "is_target": true,
              "columns": [
                  { "name": "COL1" },
                  { "name": "COL2" },
                  { "name": "COL3" },
                  { "name": "COL4" }
          ]}
      ]
  },
```
- We introduce a header record and define it as a record of a single-row, and of a single instance.
- Then we add `"is_target": true` to the actual data record since now we have two records and we
must tell the parser which one is the target to ingest and transform.

## Use Case: Simple CSV with Header And Header Verification Needed

Sample input:
```
C1,C2,C3,C4
1, 2, 3, 4
A, B, C, D
...
```

It has a header, we do care about the header row and enforce/verify its column names:

```
  "file_declaration": {
      "delimiter": ",",
      "records": [
          {
              "min": 1, "max": 1,
              "header": "^C1,C2,C3,C4$"
          },
          {
              "is_target": true,
              "columns": [
                  { "name": "COL1" },
                  { "name": "COL2" },
                  { "name": "COL3" },
                  { "name": "COL4" }
          ]}
      ]
  },
```
- Now if the header row is missing, or the column names are not exactly "C1,C2,C3,C4", then parser
will fail the ingestion operation. One note, in the `"header"` regexp pattern, you must **not** add
any space before/after the delimiter, whether the actual header row contains spaces around
delimiters or not.

## Use Case: CSV with Fixed Multi-row Records

Sample input:
```
1, 2, 3, 4
A, B, C, D
5, 6, 7, 8
W, X, Y, Z
```

Each record consists of two lines, and the record columns extract data from various columns from
both lines.

```
  "file_declaration": {
      "delimiter": ",",
      "records": [
          {
              "rows": 2,
              "columns": [
                  { "name": "COL1", "index": 3, "line_index": 1 },
                  { "name": "COL2", "index": 4, "line_index": 1 },
                  { "name": "COL3", "index": 1, "line_index": 2 },
                  { "name": "COL4", "index": 2, "line_index": 2 }
          ]}
      ]
  },
```
- `"rows": 2` says this record consists of fixed-number (2) rows.
- Lacking of `"min"`/`"max"` indicates the record can repeat 0 or many times.
- Since this is the sole record, `"is_target": true` is inferred.
- `COL1` data will be extracted from the first line (`line_index` is 1-based), column 3 (`index` is 1-based).
- `COL2` data will be extracted from the first line, column 4.
- `COL3` data will be extracted from the second line column 1.
- `COL4` data will be extracted from the second line column 2.

If we have this following simple `transform_declarations`:
```
 "transform_declarations": {
   "FINAL_OUTPUT": { "object": {
     "col1": { "xpath": "COL1" },
     "col2": { "xpath": "COL2" },
     "col3": { "xpath": "COL3" },
     "col4": { "xpath": "COL4" }
   }}
 }
```

Then the output for the above sample input will be:
```
[
	{
		"col1": "3",
		"col2": "4",
		"col3": "A",
		"col4": "B"
	},
	{
		"col1": "7",
		"col2": "8",
		"col3": "W",
		"col4": "X"
	}
]
```

## Use Case: CSV with Variable Multi-row Records Defined by Header and Footer

Sample input:
```
B
1, 2, 3, 4
A, B, C, D
E
B
5, 6, 7, 8
E
```

Each record is marked with a `B` at the beginning and an `E` at the end, and contains variable
number of rows. For the ingestion and transform, say, we're only interested in the rows with numeric
data.

```
  "file_declaration": {
      "delimiter": ",",
      "records": [
          {
              "header": "^B$", "footer": "^E$",
              "columns": [
                  { "name": "COL1", "line_pattern": "^[0-9]" },
                  { "name": "COL2", "line_pattern": "^[0-9]" },
                  { "name": "COL3", "line_pattern": "^[0-9]" },
                  { "name": "COL4", "line_pattern": "^[0-9]" }
          ]}
      ]
  },
```
- `"header"` and `"footer"` specifies the marker pattern of the record. Note if `"footer"` isn't
specified, then `"header"` alone will turn into a single line matching.
- `COL1` data will be extracted from column 1 of the first line inside the record that starts with a numeric value.
- `COL2` data will be extracted from column 2 of the first line inside the record that starts with a numeric value.
- `COL3` data will be extracted from column 3 of the first line inside the record that starts with a numeric value.
- `COL4` data will be extracted from column 4 of the first line inside the record that starts with a numeric value.

If we have this following simple `transform_declarations`:
```
 "transform_declarations": {
   "FINAL_OUTPUT": { "object": {
     "col1": { "xpath": "COL1" },
     "col2": { "xpath": "COL2" },
     "col3": { "xpath": "COL3" },
     "col4": { "xpath": "COL4" }
   }}
 }
```

Then the output for the above sample input will be:
```
[
	{
		"col1": "1",
		"col2": "2",
		"col3": "3",
		"col4": "4"
	},
	{
		"col1": "5",
		"col2": "6",
		"col3": "7",
		"col4": "8"
	}
]
```

## Use Case: CSV with Nested CSV Records

Sample input:
```
H,WJ84,20200504,085048
D,WJ84,"XF460080188GB","","OL7 0PZ","Post Office at AL4 9RB","EVPPA"
D,WJ84,"XF460080259GB","","OL7 0PZ","Post Office at CO3 4RZ","EVPPA"
H,WJ85,20200505,100248
D,WJ85,"XF460080456GB","","OL7 0PZ","Post Office at AL4 9RB","EVPPA"
D,WJ85,"XF460080758GB","","OL7 0PZ","Post Office at CO3 4RZ","EVPPA"
```

Each record starts with an `"H,"` row which contains some info about, say, a wire operation info,
such as wire number (column 2) and date (column 3) and time (column 4).
Then inside the record, we have a bunch of "D," rows each of which specifies an item in the wire
operation, which contains item code (column 3), destination (column 5), etc.

This is clearly a nested structure, we can define the `file_declaration` as follows:

```
    "file_declaration": {
        "delimiter": ",",
        "records": [
            {
                "name": "H", "header": "^H,", "is_target": true,
                "columns": [
                    { "name": "WIRE_NUM", "index": 2 },
                    { "name": "DATE", "index": 3 },
                    { "name": "TIME", "index": 4 }
                ],
                "child_records": [
                    {
                        "name": "D", "header": "^D,", "min": 1,
                        "columns": [
                            { "name": "ITEM_ID", "index": 3 },
                            { "name": "DESTINATION", "index": 5 },
                            { "name": "EVENT_LOCATION", "index": 6 },
                            { "name": "EVENT_CODE", "index": 7 }
                        ]
                    }
                ]
            }
        ]
    },
```

If we have this following simple `transform_declarations`:
```
     "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "wire_number": { "xpath": "WIRE_NUM" },
            "wire_time": { "xpath": ".", "template": "wire_date_time_template" },
            "items": { "array": [ { "xpath": "D", "object": {
                "id": { "xpath": "ITEM_ID"},
                "destination": { "xpath": "DESTINATION"},
                "event_loc": { "xpath": "EVENT_LOCATION"},
                "event_code": { "xpath": "EVENT_CODE"}
            }}]}
        }},
        "wire_date_time_template": { "custom_func": {
            "name": "dateTimeToRFC3339",
            "args": [
                { "custom_func": {
                    "name": "concat",
                    "args": [
                        { "xpath": "DATE" },
                        { "const": "T" },
                        { "xpath": "TIME" }
                    ]
                }},
                { "const": "", "_comment": "input timezone" },
                { "const": "", "_comment": "output timezone" }
            ]
        }}
    }
```

Then the output for the above sample input will be:
```
[
	{
		"items": [
			{
				"destination": "OL7 0PZ",
				"event_code": "EVPPA",
				"event_loc": "Post Office at AL4 9RB",
				"id": "XF460080188GB"
			},
			{
				"destination": "OL7 0PZ",
				"event_code": "EVPPA",
				"event_loc": "Post Office at CO3 4RZ",
				"id": "XF460080259GB"
			}
		],
		"wire_number": "WJ84",
		"wire_time": "2020-05-04T08:50:48"
	},
	{
		"items": [
			{
				"destination": "OL7 0PZ",
				"event_code": "EVPPA",
				"event_loc": "Post Office at AL4 9RB",
				"id": "XF460080456GB"
			},
			{
				"destination": "OL7 0PZ",
				"event_code": "EVPPA",
				"event_loc": "Post Office at CO3 4RZ",
				"id": "XF460080758GB"
			}
		],
		"wire_number": "WJ85",
		"wire_time": "2020-05-05T10:02:48"
	}
]
```

## Debug Info
If debug information is desired to be added into each record's IDR therefore becoming available for
`transform_declarations` use, one can specify `"debug": 1` in `parser_settings` section of the schema:
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv2",
        "debug": 1
    },
    ...
}
```
If the `csv2` parser detects `"debug": 1` setting, it will add the following debug info into the
current record's IDR structure:

- xpath: `__debug/line_num`: contains the starting line number in the CSV file of the current
record.

Check this [sample](../extensions/omniv21/samples/csv2/1_single_row.schema.json) for usage pattern.

## Migration from `'csv'` Schemas

If one looks at the documentation for the old `csv` schema [here](./csv_in_depth.md), you notice
`csv` and `csv2` schemas are fairly similar! It certainly makes migration easy:

- Change 1: `'parser_settings.file_format_type'` changes from `'csv'` to `'csv2'`.

- Change 2:

    If `file_declaration.header_row_index` is used, which means the old `csv` schema wants to
    verify the header row, we need to add a header record in the new schema:
    ```
        "records": [
            {
                "min": 1, "max": 1,
                "header": "^<COL 1>|<COL 2>|...|<COL N>$"
            },
    ```
    This way, the parser will actively look for one and only one line at the beginning (empty lines
    always ignored) to be exactly matched by that is specified in the `"header"` regexp.

- Change 3: simply remove `file_declaration.data_row_index`.

- Change 4: create a record for the data row, and move the old schema's `columns` inside the new
record, and remove `alias`, since we can specify the column name to be XPath query friendly, thus
voiding the need for `alias`.

That's it.

# CSV Schema in Depth

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
    "delimiter": "<delimiter>"                  <= required
    "replace_double_quotes": true/false,        <= optional
    "header_row_index": integer >= 1,           <= optional
    "data_row_index": integer >= 1,             <= required
    "columns": [                                <= required, must not be empty array.
        {
            "name": "<column name>",            <= required
            "alias": "<alias name>"             <= optional
        },
        ...
    ]
}
```

- `delimiter`: self-explanatory.
    - Note 1: the delimiter doesn't have to be of a single character.
    - Note 2: the delimiter doesn't have to be limited to ASCII character(s), utf8 is supported.

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

- `header_row_index`: line number (1-based) where the column header declaration sits in the input.

    If specified, omniparser will read columns in from the given line and compare them
    against the declared column names specified in `columns` section. If the actual columns read from the
    input have fewer values than the declared columns, omniparser will fail; If the actual columns read
    from the input have more values than the declared columns, the excessive actual columns are ignored;
    if any column value mismatches, omniparser will fail. All header verification failures are considered
    fatal, i.e. the entire ingestion and transform operation aginst the input will fail.

    If not specified, no header verification is done and omniparser will assume the column names and order
    declared in the `columns` section for the input data.

- `data_row_index`: line number (1-based) where the first actual data line starts in the input. Required.

- `columns.name`: the name of a column.
    - Note 1: it must match the corresponding column header value if `header_row_index` is specified.
    - Note 2: if name contains white space, then `alias` use is advised (to make XPath query possible).

- `columns.alias`: the alias of a column. Optional.

    If a column's name contains space, while it's completely legitimate in CSV, it would make the XPath
    based transform hard/impossible later. In situations like this, we strongly advise schema writer to
    use `alias` to assign an alias to the column that is XPath friendly, such as containing no spaces.

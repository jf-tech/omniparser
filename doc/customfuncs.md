# Custom Function Reference

## Global `custom_func` Available to All Extensions and Versions of Schema Handlers

> ### coalesce

**Synopsis**: `coalesce` returns the first non-empty string of the input strings. If no input
strings are given or all of them are empty, then empty string is returned. Note: a blank
string (with only whitespaces) is not considered as empty.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#Coalesce).

**Example**:
```
"tracking_number": { "custom_func": {
    "name": "coalesce",
    "args": [
        { "xpath": "tracking_number_h002_cn" },
        { "xpath": "tracking_number_h001" }
    ]
}}
```
If IDR node `tracking_number_h002_cn` value is `""` and `tracking_number_h001` value is `"ABC"`,
then the result field `tracking_number` value is `"ABC"`.
---

> ### concat

**Synopsis**: `concat` concatenates a number of strings together. If no strings specified, `""` is
returned.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#Concat).

**Example**:
```
"event_date_time": { "custom_func": {
    "name": "concat",
    "args": [
        { "xpath": "event_date" },
        { "const": "T" },
        { "xpath": "event_time" }
    ]
}}
```
If IDR node `event_date` value is `"12/31/2020"` and `event_time` value is `"12:34:56"`,
then the result field `event_date_time` value is `"12/31/2020T12:34:56"`.
---

> ### dateTimeLayoutToRFC3339

**Synopsis**: `dateTimeLayoutToRFC3339` parses a datetime string according to a given layout, and
normalizes and returns it in RFC3339 format.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#DateTimeLayoutToRFC3339).

**Example**:
```
"day_before_month": { "custom_func": {
    "name": "dateTimeLayoutToRFC3339",
    "args": [
        { "xpath": "DayBeforeMonth" },
        { "const": "02/01/06T15:04:05", "_comment": "layout" },
        { "const": "false", "_comment": "layoutTZ" },
        { "const": "Pacific/Auckland", "_comment": "fromTZ" },
        { "const": "America/Los_Angeles", "_comment": "toTZ" }
    ]
}}
```
If IDR node `DayBeforeMonth` value is `"31/12/20T12:34:56"`, then the result field
`day_before_month` value is `"2020-12-30T15:34:56-08:00"`: first, input `"31/12/20T12:34:56"`
is parsed in the time zone of `"Pacific/Auckland"` which is GMT+13 at that moment of time.
So the input datetime string, if translated into UTC, is `"2020-12-30T23:34:56Z"`. Now the caller
specifies the desired output timezone to be `"America/Los_Angeles"` which is GMT-8 at that moment
of time, so the eventual output is `"2020-12-30T15:34:56-08:00"`.

Param `layoutTZ` should be a boolean string value `"true"` or `"false"`, depending on if the
provided `layout` param contains TZ info (such as `Z` suffix, or tz offset like `-08:00`) in it or
not. If `layout` has TZ info, then `fromTZ` param will be ignored; if `layout` and input don't have
TZ info, and `fromTZ` is specified, then the datetime string will be parsed in as if it's in the
`fromTZ` timezone. If `toTZ` is empty, then whatever the tz from the input parsing will remain intact;
or the parsed input datetime will be converted into the `toTZ`.

If you're not sure, please check
[this sample](../extensions/omniv21/samples/xml/1_datetime_parse_and_format.schema.json) to find out
more subtleties about date time parsing and conversion.
---

> ### dateTimeToEpoch

**Synopsis**: `dateTimeToEpoch` parses a datetime string intelligently, and returns its epoch number.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#DateTimeToEpoch).

**Example**:
```
"epoch": { "custom_func": {
    "name": "dateTimeToEpoch",
    "args": [
        { "xpath": "event_datetime" },
        { "const": "", "_comment": "fromTZ" },
        { "const": "SECOND", "_comment": "unit" }
    ]
}}
```
If IDR node `event_datetime` value is `"12/31/2020T12:34:56Z"`, then the result field
`epoch` value is `"1609418096"`: first, input `"12/31/2020T12:34:56Z"`
will be parsed in as is (since it has `Z` suffix, so it is time-zoned, plus `fromTZ` is `""`).
Then the function converts the input datetime to epoch seconds.

Param `unit` has two valid values: `"SECOND"` or `"MILLISECOND"`.

---

> ### dateTimeToRFC3339

**Synopsis**: `dateTimeToRFC3339` parses a datetime string intelligently, normalizes and returns it
in RFC3339 format.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#DateTimeToRFC3339).

**Example**:
```
"no_tz_date_time_use_to_tz": { "custom_func": {
    "name": "dateTimeToRFC3339",
    "args": [
        { "xpath": "DateTimeWithNoTZ" },
        { "const": "", "_comment": "fromTZ" },
        { "const": "America/Los_Angeles", "_comment": "toTZ" }
    ]
}},
```
This `custom_func` params and behavior is very similar to
[`dateTimeLayoutToRFC3339`](#datetimelayouttorfc3339) except that it doesn't need a `layout` and it
instead tries to parse the input datetime string intelligently.

If you're not sure, please check
[this sample](../extensions/omniv21/samples/xml/1_datetime_parse_and_format.schema.json) to find out
more subtleties about date time parsing and conversion.
---

> ### epochToDateTimeRFC3339

**Synopsis**: `epochToDateTimeRFC3339` translates an epoch timestamp into an RFC3339 formatted datetime
string.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#EpochToDateTimeRFC3339).

**Example**:
```
"epoch_to_datetime": { "custom_func": {
    "name": "epochToDateTimeRFC3339",
    "args": [
        { "xpath": "event_epoch" },
        { "const": "MILLISECOND", "_comment": "unit" }
    ]
}}
```
If IDR node `event_epoch` value is `"1609418096000"`, then the result field
`epoch_to_datetime` value is `"2020-12-31T12:34:56Z"`: first, input `"1609418096000"`
will be parsed in as epoch value in milliseconds. Then the function converts the epoch value
to RFC3339 string.

Param `unit` has two valid values: `"SECOND"` or `"MILLISECOND"`.

There is an optional param at the end, `tz`: if not specified, the output will be in UTC (`Z`)
time zone; if specified, it must be of a standard IANA time zone string, such as
`"America/Los_Angeles"`.
---

> ### lower

**Synopsis**: `lower` lowers the case of an input string.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#Lower).

**Example**:
```
"carrier": { "custom_func": { "name": "lower", "args": [ { "xpath": "../GLOBAL/carrier" } ] } },
```
If IDR node `../GLOBAL/carrier` value is `"ABC"`, then the result field `carrier` value is `"abc"`.

---

> ### now

**Synopsis**: `now` returns the current time in UTC in RFC3339 format.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#Now).

**Example**:
```
"now_datetime": { "custom_func": { "name": "now" }},
```
The result field `now_datetime` value will be the current system datetime in UTC in RFC3339.

---

> ### upper
> 
**Synopsis**: `upper` uppers the case of an input string.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#Upper).

**Example**:
```
"carrier": { "custom_func": { "name": "upper", "args": [ { "xpath": "../GLOBAL/carrier" } ] } },
```
If IDR node `../GLOBAL/carrier` value is `"abc"`, then the result field `carrier` value is `"ABC"`.

---

> ### uuidv3

**Synopsis**: `uuidv3` uses MD5 to produce a consistent/stable UUID for an input string.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/customfuncs#UUIDv3).

**Example**:
```
"unique_customer_order_id": { "custom_func": {
    "name": "uuidv3",
    "args": [
        { "custom_func": {
            "name": "concat",
            "args: [
                { "xpath": "customer_id" },
                { "const": "/" },
                { "xpath": "order_id" }
            ]
        }}
    ]
}}
```
The result field `unique_customer_order_id` will contain the value of an MD5 hash of the concatenated
string of customer_id value, `"/"`, and order_id value.

---

## `omni.2.1` Schema Handler Specific `custom_func`

> ### copy

**Synopsis**: `copy` copies the current contextual `idr.Node` and returns it as a JSON marshaling
friendly `interface{}`.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/extensions/omniv21/customfuncs#CopyFunc).

**Example**:
```
"first_book": { "xpath": "book[position() = 1]", "custom_func": { "name": "copy"} },
```
The result field `first_book` will be an exact copy of first `book` node from the input.

---

> ### javascript

**Synopsis**: `javascript` runs a javascript.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/extensions/omniv21/customfuncs#JavaScript).

**Example**:
```
"avg_price": {
    "custom_func": {
        "name": "javascript",
        "args": [
            { "const": "t=0; for(i=0; i<prices.length; i++) { t+=prices[i]; } Math.floor(t*100/prices.length)/100;" },
            { "const": "prices" }, { "array": [ { "xpath": "books/*/price", "type": "float" } ] }
        ]
    }
},
```
The result field `avg_price` will contain the average price (to the 2nd decimal places) of all the
book prices.

For more information about `javascript`, check this
[in-depth explanation](./use_of_custom_funcs.md#javascript-and-javascript_with_context).
---

> ### javascript_with_context

**Synopsis**: `javascript_with_context` runs a javascript with contextual `_node` provided.

**Pkg doc**: [here](https://pkg.go.dev/github.com/jf-tech/omniparser/extensions/omniv21/customfuncs#JavaScriptWithContext).

**Example**:
```
"full_name": { "xpath": "./personal_info", "custom_func" {
    "name": "javascript_with_context",
    "args": [
        { "const": "var n = JSON.parse(_node); n.['Last Name'] + ', ' + n.['First Name']" }
    ]
}}
```
The result field `full_name` will be a concatenated string of the last name and the first name from
the current IDR node.

For more information about `javascript_with_context`, check this
[in-depth explanation](./use_of_custom_funcs.md#javascript-and-javascript_with_context).
---

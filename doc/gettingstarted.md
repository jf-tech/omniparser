# Getting Started

This page is a step-by-step introduction of how to write an omniparser schema (specifically tailored
to the latest `"omni.2.1"` schema version) and how to ingest and transform inputs programmatically
and by the CLI tool.

## Prerequisites and Notes

- Golang 1.14 installed.
- [`github.com/jf-tech/omniparser`](https://github.com/jf-tech/omniparser) cloned (assuming the clone
location is: `~/dev/jf-tech/omniparser/`)
- The guide assumes Mac OS or Linux dev environment. If you're on Windows, you need make minor
adjustments from the bash commands posted below. 

## The Input
We'll use a simple CSV input for this guide. We have separate pages that go into details about other
input formats.

Consider the following CSV input (simplified and changed from the full example
[here](../extensions/omniv21/samples/csv/1_weather_data_csv.input.csv)):

```
                    DATE|HIGH TEMP C|LOW TEMP F|WIND DIR|WIND SPEED KMH
01/31/2019 12:34:56-0800|       10.5|      30.2|       N|            33
07/31/2020 01:23:45-0500|         39|        95|      SE|             8
```

Bear in mind that this example is somewhat contrived (such as mixed use celsius and fahrenheit) so
that we can illustrate capabilities of omniparser schema.

## The Desired Output
Omniparser ingests and transforms an input (stream) into a number of output records in JSON. For CSV
a basic input unit for ingestion and transform is a single data line. For this example, we want to
transform each of the data line into the following JSON output:
```
[
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": 50.9,
		"low_temperature_fahrenheit": 30.2,
		"wind": "North 20.5 mph"
	},
	{
		"date": "2020-07-31T01:23:45-05:00",
		"high_temperature_fahrenheit": 102.2,
		"low_temperature_fahrenheit": 95,
		"wind": "South East 4.97 mph"
	}
]
```
As you can see, in the desired output, we'd like to standardize all the input temperatures into the
same fahrenheit unit; we'd also like to do some translation such that the wind direction and wind
speed are "massaged" into a more human readable string; lastly, we'd like to normalize the date text
into [RFC-3339](https://tools.ietf.org/html/rfc3339) standard format.

## CLI (command line interface)

Before we get into schema writing, let's first get familiar with omniparser CLI so that we can easily
and incrementally test our schema writing.

Assuming you have the git repo cloned at `~/dev/jf-tech/omniparser/`, simply run this bash script:
```
~/dev/jf-tech/omniparser/cli.sh help
```

For this guide, we will need to use `transform` command:
```
~/dev/jf-tech/omniparser/cli.sh transform --help
```

Now assuming your temporary working directory for this guide is: `~/Downloads/omniparser/guide/`,
let's create two files there, one for the input and one for the schema:
```
$ cd ~/Downloads/omniparser/guide/
$ touch input.csv
$ touch schema.json
```
Use any editor to cut & paste the CSV content from [The Input](#the-input) into `input.csv`, and
now run omniparser CLI from `~/Downloads/omniparser/guide/`:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json 
Error: unable to perform schema validation: EOF
```
Expected, given the `schema.json` is still empty.

Now we're ready to go!

## Schema Writing

### `parser_settings`

This is the common part of all omniparser schemas, the header `parser_settings`:
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv"
    }
}
```
It's self-explanatory. Now let's run the CLI again:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json 
Error: schema 'schema.json' validation failed: (root): transform_declarations is required
```

A schema is to tell the parser how to do transform, which is missing from the current schema, let's
add that...

### `transform_declarations` and `FINAL_OUTPUT`

`transform_declarations` is a section in the schema with specific instructions to parser how to do
transformation. Let's add an empty `transform_declarations` for now:
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv"
    },
    "transform_declarations": {}
}
```
Run the CLI we get another error:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
Error: schema 'schema.json' validation failed: transform_declarations: FINAL_OUTPUT is required
```
Let's add an empty `FINAL_OUTPUT` in:
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv"
    },
    "transform_declarations": {
        "FINAL_OUTPUT": {}
    }
}
```
`FINAL_OUTPUT` is the special name reserved for the transform template that will be used for
the output. Given the section is called `transform_declarations` you might have guessed we can have
multiple templates defined in it. Each template can reference other templates. There must be one
and only one template called `FINAL_OUTPUT`.

Run the CLI we get a new error:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
Error: schema 'schema.json' validation failed: (root): file_declaration is required
```

Ah...seems like a new section `file_declaration` is needed...

### `file_declaration`

While `transform_declarations` contains instructions to the parser on how to transform the ingested
data into the desired output format, we still owe the parser the instructions how to ingest the input
stream, there comes `file_declaration`. (Note not all input formats require a `file_declaration`
section, e.g. JSON and XML inputs need no `file_declaration` in their schemas.)

For CSV, we need to define the following settings:
- What's the delimiter character, comma or something else?
- Is there a header in the CSV input that defines the names of each column? If no, what each column
should be called during ingestion and transformation?
- Where do the actual data lines begin?

For this guide example, the settings are:
- delimiter is `|`
- Yes there is a header line with all the columns names defined.
- The actual data lines start at line 2. (Line number is 1-based.)

Let's add these:
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv"
    },
    "file_declaration": {
        "delimiter": "|",
        "header_row_index": 1,
        "data_row_index": 2,
        "columns": [
            { "name": "DATE" },
            { "name": "HIGH TEMP C" },
            { "name": "LOW TEMP F" },
            { "name": "WIND DIR" },
            { "name": "WIND SPEED KMH" }
        ]
    },
    "transform_declarations": {
        "FINAL_OUTPUT": {}
    }
```

Run the CLI again:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
[
	"01/31/2019 12:34:56-0800       10.5      30.2       N            33",
	"07/31/2020 01:23:45-0500         39        95      SE             8"
]
```

Wow, we're getting something! Although not exactly matching what we want in
[The Desired Output](#the-desired-output), it's a big step forward!

Seems like the ingestion and transform directly copy each line from the `input.csv` and output
as a JSON text string in the final output JSON array.

To understand what's behind the scene and build the foundation of understanding the inner working
of the parser, we need to deviate from the schema writing for a moment...

### How Each Ingested Data Record Is Represented in Memory

Short answer: [IDR](../idr/README.md).

IDR is an in-memory data representation format used by omniparser for the ingested data from
all input formats. If you're interested in more technical details, check the IDR doc mentioned above.

CSV has a very simple IDR representation: each data line is mapped to an IDR tree, where each column
is mapped to the tree's leaf nodes. So for our sample input csv here, the first data line would be
represented by the following IDR:
```
|
+--"DATE"
|    +--"01/31/2019 12:34:56-0800"
|
+--"HIGH TEMP C"
|    +--"       10.5"
|
+--"LOW TEMP F"
|    +--"         39"
|
+--"WIND DIR"
|    +--"       N"
|
+--"WIND SPEED KMH"
     +--"            33"
```

You can imaginarily convert the IDR into XML which helps you understand the extensive use of XPath
queries later in transformation:
```
<>
   <DATE>01/31/2019 12:34:56-0800</DATE>
   <HIGH TEMP C>       10.5</HIGH TEMP C>
   <LOW TEMP F>         39</LOW TEMP F>
   <WIND DIR>       N</WIND DIR>
   <WIND SPEED KMH>            33</WIND SPEED KMH>
</>
```
Note XML/XPath don't like element name containing spaces. While IDR doesn't care about names with
spaces, XPath queries used in transforms do care and will break. So we'd like to **assign some
XPath friendly column name aliases in our schema, if the raw column names containing special chars**:

Let's make small modifications to our schema:
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv"
    },
    "file_declaration": {
        "delimiter": "|",
        "header_row_index": 1,
        "data_row_index": 2,
        "columns": [
            { "name": "DATE" },
            { "name": "HIGH TEMP C", "alias": "HIGH_TEMP_C" },
            { "name": "LOW TEMP F", "alias": "LOW_TEMP_F" },
            { "name": "WIND DIR", "alias": "WIND_DIR" },
            { "name": "WIND SPEED KMH", "alias": "WIND_SPEED_KMH" }
        ]
    },
    "transform_declarations": {
        "FINAL_OUTPUT": {}
    }
}
```

Rerun the CLI to ensure everything is still working. Now the IDR and its imaginary converted XML
equivalent look like this:
```
<>
   <DATE>01/31/2019 12:34:56-0800</DATE>
   <HIGH_TEMP_C>       10.5</HIGH_TEMP_C>
   <LOW_TEMP_F>         39</LOW_TEMP_F>
   <WIND_DIR>       N</WIND_DIR>
   <WIND_SPEED_KMH>            33</WIND_SPEED_KMH>
</>
```
Remember this, and we'll move onto some real transformation!

### `transform_declarations` and `FINAL_OUTPUT` for Real

Recall that we want to convert a data line such as
```
01/31/2019 12:34:56-0800|       10.5|      30.2|       N|            33
```
into a JSON object output such as
```
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": 50.9,
		"low_temperature_fahrenheit": 30.2,
		"wind": "North 20.5 mph"
	}
```

So let's define the object skeleton in the `FINAL_OUTPUT` in `schema.json`:
```
    "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "date": { "xpath": "DATE" },
            "high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C" },
            "low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F" },
            "wind": { "xpath": "WIND_DIR" }
        }}
    }
```
(Note for brevity, `parser_settings` and `file_declaration` are omitted above.)

We first defined `FINAL_OUTPUT` is an `object`, which contains a number of fields, such as `date`,
`high_temperature_fahrenheit`, etc.

For each of the field, we simply copy the data text as its value from an XPath query run on the IDR.

Remember for the first data line, its corresponding IDR (or the IDR's equivalent XML) looks like:
```
<>
   <DATE>01/31/2019 12:34:56-0800</DATE>
   <HIGH_TEMP_C>       10.5</HIGH_TEMP_C>
   <LOW_TEMP_F>         39</LOW_TEMP_F>
   <WIND_DIR>       N</WIND_DIR>
   <WIND_SPEED_KMH>            33</WIND_SPEED_KMH>
</>
```
Thus, an XPath query `"xpath": "DATE"` on the root of the IDR would return `01/31/2019 12:34:56-0800`, which is
used as the value for the field `date`. So on and so forth for all other fields.

Run the CLI, we have:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
[
	{
		"date": "01/31/2019 12:34:56-0800",
		"high_temperature_fahrenheit": "10.5",
		"low_temperature_fahrenheit": "30.2",
		"wind": "N"
	},
	{
		"date": "07/31/2020 01:23:45-0500",
		"high_temperature_fahrenheit": "39",
		"low_temperature_fahrenheit": "95",
		"wind": "SE"
	}
]
```

Much better! The result, at least structurally, looks like our desired output.

One small observation you might've taken is that any leading and trailing white spaces are stripped
during the transformation. (Don't worry, there is a way to preserve leading/trailing white spaces if
you wonder.)

Yes the result looks better, but still far from the ideal. Let's fix the issues one by one.

### Fix `FINAL_OUTPUT.date`

Recall we want the `date` field in the desired output to be RFC-3339 compliant. We can use a parser
built-in function to achieve this:
```
    "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "date": { "custom_func": {
                "name": "dateTimeToRFC3339",
                "args": [
                    { "xpath": "DATE" },
                    { "const": "", "_comment": "input timezone" },
                    { "const": "", "_comment": "output timezone" }
                ]
            }},
            "high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C" },
            "low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F" },
            "wind": { "xpath": "WIND_DIR" }
        }}
    }
```

Run CLI we have:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
[
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": "10.5",
		"low_temperature_fahrenheit": "30.2",
		"wind": "N"
	},
	{
		"date": "2020-07-31T01:23:45-05:00",
		"high_temperature_fahrenheit": "39",
		"low_temperature_fahrenheit": "95",
		"wind": "SE"
	}
]
```
Yes!!

So basically we changed the simple `"date": { "xpath": "DATE" },` directive into a function call:
```
            "date": { "custom_func": {
                "name": "dateTimeToRFC3339",
                "args": [
                    { "xpath": "DATE" },
                    { "const": "", "_comment": "input timezone" },
                    { "const": "", "_comment": "output timezone" }
                ]
            }},
```

These built-in functions are called `custom_func` (and yes, programmatic users of omniparser have
the ability to add additional functions). To invoke a `custom_func`, you need to provide the name
of the function, in this case `dateTimeToRFC3339`, and a list of arguments the function requires.
(You can find the full references to all built-in `custom_func` [here](./customfuncs.md).)

The first argument here is `{ "xpath": "DATE" },` basically providing the function the input datetime
string. The second argument `dateTimeToRFC3339` requires specifies what time zone the input datetime
string is in. Since the datetime strings in the guide sample CSV already contain time zone offsets
(`-0800`, `-0500`), an empty string is supplied to the input time zone argument. The third argument is
the desired output time zone. If, say, we want to standardize all the `date` fields in the output to
be in time zone of `America/Los_Angeles`, we can specify it in the third argument, and the
`custom_func` will perform the correct time zone shifts for us.

### Fix `FINAL_OUTPUT.high_temperature_fahrenheit`

Note the `HIGH_TEMP_C` in input data is in celsius, and the desired output calls for fahrenheit. Let's
fix that:
```
    "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "date": { "custom_func": {
                "name": "dateTimeToRFC3339",
                "args": [
                    { "xpath": "DATE" },
                    { "const": "", "_comment": "input timezone" },
                    { "const": "", "_comment": "output timezone" }
                ]
            }},
            "high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C", "template": "template_c_to_f" },
            "low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F" },
            "wind": { "xpath": "WIND_DIR" }
        }},
        "template_c_to_f": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "Math.floor((temp_c * 9 / 5 + 32) * 10) / 10" },
                    { "const": "temp_c" }, { "xpath": ".", "type": "float" }
                ]
            }
        }
    }
```

Here we introduce two new things: 1) template and 2) custom_func `javascript`.

1) Template

    Template is for schema snippet reuse or sometimes simply for readability. Imagine we have a
    schema in which we need to transform multiple fahrenheit data into celsius values, instead doing
    the math again and again, we can write a template (`template_c_to_f`) that can be reused.

    In our case, we replaced the simple direct-copying directive
    `"high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C" },` to template directive
    `"high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C", "template": "template_c_to_f" },`

    Note we still have `"xpath": "HIGH_TEMP_C"` clause before `"template": "template_c_to_f"`. What it
    says is to run/apply the template against the IDR sub-tree anchored (or rooted) on the node
    `<HIGH_TEMP_C>`.

2) custom_func `javascript`

    Now let's take a closer look at the template `template_c_to_f`:
    ```
           "template_c_to_f": {
               "custom_func": {
                   "name": "javascript",
                   "args": [
                       { "const": "Math.floor((temp_c * 9 / 5 + 32) * 10) / 10" },
                       { "const": "temp_c" }, { "xpath": ".", "type": "float" }
                   ]
               }
           }
    ```
    custom_func `javascript` takes a number of arguments: the first one is the actual script string,
    and all remaining arguments are to provide values for all the variables declared in the script
    string, in this particular case, only one variable `temp_c`. All remaining arguments come in
    pairs. The first in each pair always declares what variable the second in pair is about. And the
    second in each pair provides the actual value for the variable. In this example, we see variable
    `temp_c` should have a value based on the XPath query `"."` and converted into `float` type.
    Remember this template's invocation is anchored on the IDR node `<HIGH_TEMP_C>`, thus XPath query
    `"."` returns its text value `"10.5"`, after which it was converted into numeric value `10.5`
    before the math computation starts.

    Type conversion should be used only when needed. When we convert text `"10.5"` into numeric float
    value `10.5`, `"type": "float"` is used. However when the script is done, the result is already
    in float, there is no need to specify `"type": "float"` for the `custom_func` directive.

Now let's run CLI:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
[
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": 50.9,
		"low_temperature_fahrenheit": "30.2",
		"wind": "N"
	},
	{
		"date": "2020-07-31T01:23:45-05:00",
		"high_temperature_fahrenheit": 102.2,
		"low_temperature_fahrenheit": "95",
		"wind": "SE"
	}
]
```
Great! `high_temperature_fahrenheit` looks nice now. But `low_temperature_fahrenheit` still needs a
minor fix.

### Fix `low_temperature_fahrenheit`

Observe the output above you can see `low_temperature_fahrenheit` output value is a string, not a
numeric value. That should be an easy fix:
```
    "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "date": { "custom_func": {
                "name": "dateTimeToRFC3339",
                "args": [
                    { "xpath": "DATE" },
                    { "const": "", "_comment": "input timezone" },
                    { "const": "", "_comment": "output timezone" }
                ]
            }},
            "high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C", "template": "template_c_to_f" },
            "low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F", "type": "float" },
            "wind": { "xpath": "WIND_DIR" }
        }},
        "template_c_to_f": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "Math.floor((temp_c * 9 / 5 + 32) * 10) / 10" },
                    { "const": "temp_c" }, { "xpath": ".", "type": "float" }
                ]
            }
        }
    }
```
Basically changing `"low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F" }` to
`"low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F", "type": "float" }`.

Run CLI again, we have:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
[
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": 50.9,
		"low_temperature_fahrenheit": 30.2,
		"wind": "N"
	},
	{
		"date": "2020-07-31T01:23:45-05:00",
		"high_temperature_fahrenheit": 102.2,
		"low_temperature_fahrenheit": 95,
		"wind": "SE"
	}
]
```

Almost there! The `wind` field is a bit tricky to fix...

### Fix `wind`

Recall from [The Desired Output](#the-desired-output) we want to have the field `wind` some human
readable wind stat:
```
	{
        ...
		"wind": "North 20.5 mph"
	},
```

Recall the first data line's IDR (XML equivalent) looks like:
```
<>
   ...
   <WIND_DIR>       N</WIND_DIR>
   <WIND_SPEED_KMH>            33</WIND_SPEED_KMH>
</>
```
So `wind` value needs to derive from two columns in the input CSV data line. Let's look at them one
by one.

1) Wind Direction

    In the input, the wind direction is abbreviated (such as `"N"`, `"E"`, `"SW"`, etc). In the
    desired output we want it to be English. So we need some mapping, for which again we resort to the
    all mighty custom function `javascript`:
    ```
        "wind_acronym_mapping": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "dir=='N'?'North':dir=='NE'?'North East':dir=='E'?'East':dir=='SE'?'South East':dir=='S'?'South':dir=='SW'?'South West':dir=='W'?'West':dir=='NW'?'North West':'Tornado'"},
                    { "const": "dir" }, { "xpath": "." }
                ]
            }
        }
    ```
    A giant/long `? :` ternary operator infested javascript line maps wind direction abbreviations
    into English phrases.

2) Wind Speed

    In the input, the column `WIND_SPEED_KMH` unit is kilometers per hour (kmh) while the desired
    output calls for miles per hour (mph). Let's do the conversion:
    ```
    Math.floor(kmh * 0.621371 * 100) / 100
    ```
    (Several uses of `Math.floor(...*100/100)` throughout this page is to limit the number of decimal
    places to be more human readable.) 

Put 1) and 2) together, we can have the new transform schema look like this:
```
    "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "date": { "custom_func": {
                "name": "dateTimeToRFC3339",
                "args": [
                    { "xpath": "DATE" },
                    { "const": "", "_comment": "input timezone" },
                    { "const": "", "_comment": "output timezone" }
                ]
            }},
            "high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C", "template": "template_c_to_f" },
            "low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F", "type": "float" },
            "wind": { "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "win_dir + ' ' + Math.floor(kmh * 0.621371 * 100) / 100 + ' mph'" },
                    { "const": "win_dir" }, { "xpath": "WIND_DIR", "template": "wind_acronym_mapping" },
                    { "const": "kmh" }, { "xpath": "WIND_SPEED_KMH", "type": "float" }
                ]
            }}
        }},
        "template_c_to_f": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "Math.floor((temp_c * 9 / 5 + 32) * 10) / 10" },
                    { "const": "temp_c" }, { "xpath": ".", "type": "float" }
                ]
            }
        },
        "wind_acronym_mapping": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "dir=='N'?'North':dir=='NE'?'North East':dir=='E'?'East':dir=='SE'?'South East':dir=='S'?'South':dir=='SW'?'South West':dir=='W'?'West':dir=='NW'?'North West':'Tornado'"},
                    { "const": "dir" }, { "xpath": "." }
                ]
            }
        }
    }
```

Run CLI one last time, we have:
```
$ ~/dev/jf-tech/omniparser/cli.sh transform -i input.csv -s schema.json
[
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": 50.9,
		"low_temperature_fahrenheit": 30.2,
		"wind": "North 20.5 mph"
	},
	{
		"date": "2020-07-31T01:23:45-05:00",
		"high_temperature_fahrenheit": 102.2,
		"low_temperature_fahrenheit": 95,
		"wind": "South East 4.97 mph"
	}
]
```

Perfectly match [The Desired Output](#the-desired-output)!

## Using omniparser Programmatically

While it's strongly recommended to use CLI for schema development, it is our eventual goal to use
omniparser and schemas programmatically to enable high speed / high volume processing. Below is the
code snippet of showing how to achieve this:

```
    schema, err := omniparser.NewSchema("your schema name", strings.NewReader("your schema content"))
    if err != nil { ... }
    transform, err := schema.NewTransform("your input name", strings.NewReader("your input content"), &transformctx.Ctx{})
    if err != nil { ... }
    for {
        output, err := transform.Read()
        if err == io.EOF {
            break
        }
        // output contains a []byte of the ingested and transformed record. 
    }
```

## Summary

### The Input
```
                    DATE|HIGH TEMP C|LOW TEMP F|WIND DIR|WIND SPEED KMH
01/31/2019 12:34:56-0800|       10.5|      30.2|       N|            33
07/31/2020 01:23:45-0500|         39|        95|      SE|             8
```

### The Schema
```
{
    "parser_settings": {
        "version": "omni.2.1",
        "file_format_type": "csv"
    },
    "file_declaration": {
        "delimiter": "|",
        "header_row_index": 1,
        "data_row_index": 2,
        "columns": [
            { "name": "DATE" },
            { "name": "HIGH TEMP C", "alias": "HIGH_TEMP_C" },
            { "name": "LOW TEMP F", "alias": "LOW_TEMP_F" },
            { "name": "WIND DIR", "alias": "WIND_DIR" },
            { "name": "WIND SPEED KMH", "alias": "WIND_SPEED_KMH" }
        ]
    },
    "transform_declarations": {
        "FINAL_OUTPUT": { "object": {
            "date": { "custom_func": {
                "name": "dateTimeToRFC3339",
                "args": [
                    { "xpath": "DATE" },
                    { "const": "", "_comment": "input timezone" },
                    { "const": "", "_comment": "output timezone" }
                ]
            }},
            "high_temperature_fahrenheit": { "xpath": "HIGH_TEMP_C", "template": "template_c_to_f" },
            "low_temperature_fahrenheit": { "xpath": "LOW_TEMP_F", "type": "float" },
            "wind": { "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "win_dir + ' ' + Math.floor(kmh * 0.621371 * 100) / 100 + ' mph'" },
                    { "const": "win_dir" }, { "xpath": "WIND_DIR", "template": "wind_acronym_mapping" },
                    { "const": "kmh" }, { "xpath": "WIND_SPEED_KMH", "type": "float" }
                ]
            }}
        }},
        "template_c_to_f": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "Math.floor((temp_c * 9 / 5 + 32) * 10) / 10" },
                    { "const": "temp_c" }, { "xpath": ".", "type": "float" }
                ]
            }
        },
        "wind_acronym_mapping": {
            "custom_func": {
                "name": "javascript",
                "args": [
                    { "const": "dir=='N'?'North':dir=='NE'?'North East':dir=='E'?'East':dir=='SE'?'South East':dir=='S'?'South':dir=='SW'?'South West':dir=='W'?'West':dir=='NW'?'North West':'Tornado'"},
                    { "const": "dir" }, { "xpath": "." }
                ]
            }
        }
    }
}
```

### The Code
```
    schema, err := omniparser.NewSchema("your schema name", strings.NewReader("your schema content"))
    if err != nil { ... }
    transform, err := schema.NewTransform("your input name", strings.NewReader("your input content"), &transformctx.Ctx{})
    if err != nil { ... }
    for {
        output, err := transform.Read()
        if err == io.EOF {
            break
        }
        // output contains a []byte of the ingested and transformed record. 
    }
```

### The Output
```
[
	{
		"date": "2019-01-31T12:34:56-08:00",
		"high_temperature_fahrenheit": 50.9,
		"low_temperature_fahrenheit": 30.2,
		"wind": "North 20.5 mph"
	},
	{
		"date": "2020-07-31T01:23:45-05:00",
		"high_temperature_fahrenheit": 102.2,
		"low_temperature_fahrenheit": 95,
		"wind": "South East 4.97 mph"
	}
]
```

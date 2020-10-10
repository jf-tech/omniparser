// Code generated - DO NOT EDIT.

package validation

const (
    JSONSchemaCSVFileDeclaration =
`
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "github.com/jf-tech/omniparser:csv_file_declaration",
  "title": "omniparser schema: csv/file_declaration",
  "type": "object",
  "properties": {
    "file_declaration": {
      "type": "object",
      "properties": {
        "delimiter": {
          "type": "string",
          "minLength": 1,
          "maxLength": 1
        },
        "replace_double_quotes": { "type": "boolean" },
        "header_row_index": { "type": "integer", "minimum": 1 },
        "data_row_index": { "type": "integer", "minimum": 1 },
        "columns": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "name": { "type": "string", "minLength": 1 },
              "alias": { "type": "string", "pattern": "^[_a-zA-Z0-9]+$" }
            },
            "required": [ "name" ],
            "additionalProperties": false
          },
          "minItems": 1
        }
      },
      "required": [ "delimiter", "data_row_index", "columns" ],
      "additionalProperties": false
    }
  },
  "required": [ "file_declaration" ]
}

`
)

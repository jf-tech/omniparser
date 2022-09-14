// Code generated - DO NOT EDIT.

package validation

const (
    JSONSchemaCSV2FileDeclaration =
`
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "github.com/jf-tech/omniparser:csv2_file_declaration",
    "title": "omniparser schema: csv2/file_declaration",
    "type": "object",
    "properties": {
        "file_declaration": {
            "type": "object",
            "properties": {
                "delimiter": { "type": "string", "minLength": 1, "maxLength": 1 },
                "replace_double_quotes": { "type": "boolean" },
                "records": { "$ref": "#/definitions/child_records_type" }
            },
            "required": [ "delimiter" ],
            "additionalProperties": false
        }
    },
    "required": [ "file_declaration" ],
    "definitions": {
        "child_records_type": {
            "type": "array",
            "items": {
                "oneOf": [
                    { "$ref": "#/definitions/record_group_type" },
                    { "$ref": "#/definitions/record_rows_based_type" },
                    { "$ref": "#/definitions/record_header_footer_based_type" }
                ]
            },
            "$comment": "empty child_records is fine"
        },
        "record_group_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "type": { "const": "record_group" },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "child_records": { "$ref": "#/definitions/child_records_type" }
            },
            "required": [ "type", "child_records" ], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "record_rows_based_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "rows": { "type": "integer", "minimum": 1 },
                "type": { "const": "record" },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "columns": { "$ref": "#/definitions/columns_type" },
                "child_records": { "$ref": "#/definitions/child_records_type" }
            },
            "required": [], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "record_header_footer_based_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "header": { "type": "string", "minLength": 1 },
                "footer": { "type": "string", "minLength": 1 },
                "type": { "const": "record" },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "columns": { "$ref": "#/definitions/columns_type" },
                "child_records": { "$ref": "#/definitions/child_records_type" }
            },
            "required": [ "header" ], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "columns_type": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "name": { "type": "string", "minLength": 1 },
                    "index": { "type": "integer", "minimum": 1 },
                    "line_index": { "type": "integer", "minimum": 1 },
                    "line_pattern": { "type": "string", "minLength": 1 }
                },
                "required": [ "name" ],
                "additionalProperties": false
            }
        }
    }
}

`
)

// Code generated - DO NOT EDIT.

package validation

const (
    JSONSchemaFixedLength2FileDeclaration =
`
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "github.com/jf-tech/omniparser:fixedlength2_file_declaration",
    "title": "omniparser schema: fixedlength2/file_declaration",
    "type": "object",
    "properties": {
        "file_declaration": {
            "type": "object",
            "properties": {
                "envelopes": { "$ref": "#/definitions/child_envelopes_type" }
            },
            "required": [ "envelopes" ],
            "additionalProperties": false
        }
    },
    "required": [ "file_declaration" ],
    "definitions": {
        "child_envelopes_type": {
            "type": "array",
            "items": {
                "oneOf": [
                    { "$ref": "#/definitions/envelope_group_type" },
                    { "$ref": "#/definitions/envelope_rows_based_type" },
                    { "$ref": "#/definitions/envelope_header_footer_based_type" }
                ]
            },
            "$comment": "empty child_envelopes is fine"
        },
        "envelope_group_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "type": { "const": "envelope_group" },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "child_envelopes": { "$ref": "#/definitions/child_envelopes_type" }
            },
            "required": [ "type", "child_envelopes" ], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "envelope_rows_based_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "rows": { "type": "integer", "minimum": 1 },
                "type": { "const": "envelope" },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "columns": { "$ref": "#/definitions/columns_type" },
                "child_envelopes": { "$ref": "#/definitions/child_envelopes_type" }
            },
            "required": [], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "envelope_header_footer_based_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "header": { "type": "string", "minLength": 1 },
                "footer": { "type": "string", "minLength": 1 },
                "type": { "const": "envelope" },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "columns": { "$ref": "#/definitions/columns_type" },
                "child_envelopes": { "$ref": "#/definitions/child_envelopes_type" }
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
                    "start_pos": { "type": "integer", "minimum": 1 },
                    "length": { "type": "integer", "minimum": 1 },
                    "line_index": { "type": "integer", "minimum": 1 },
                    "line_pattern": { "type": "string", "minLength": 1 }
                },
                "required": [ "name", "start_pos", "length" ],
                "additionalProperties": false
            }
        }
    }
}

`
)

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
                "envelopes": {
                    "type": "array",
                    "items": {
                        "oneOf": [
                            { "$ref": "#/definitions/envelope_rows_based" },
                            { "$ref": "#/definitions/envelope_header_footer_based" }
                        ]
                    }
                }
            },
            "required": [ "envelopes" ],
            "additionalProperties": false
        }
    },
    "required": [ "file_declaration" ],
    "definitions": {
        "envelope_rows_based": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "rows": { "type": "integer", "minimum": 1 },
                "type": { "type": "string", "enum": [ "envelope", "envelope_group" ] },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "columns": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/column_type"
                    }
                },
                "child_envelopes": {
                    "type": "array",
                    "items": {
                        "oneOf": [
                            { "$ref": "#/definitions/envelope_rows_based" },
                            { "$ref": "#/definitions/envelope_header_footer_based" }
                        ]
                    }
                }
            },
            "required": [], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "envelope_header_footer_based": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "header": { "type": "string", "minLength": 1 },
                "footer": { "type": "string", "minLength": 1 },
                "type": { "type": "string", "enum": [ "envelope", "envelope_group" ] },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "columns": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/column_type"
                    }
                },
                "child_envelopes": {
                    "type": "array",
                    "items": {
                        "oneOf": [
                            { "$ref": "#/definitions/envelope_rows_based" },
                            { "$ref": "#/definitions/envelope_header_footer_based" }
                        ]
                    }
                }
            },
            "required": [ "header" ], "$comment": "yes, 'name' is actually optional",
            "additionalProperties": false
        },
        "column_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "start_pos": { "type": "integer", "minimum": 1 },
                "length": { "type": "integer", "minimum": 1 },
                "line_pattern": { "type": "string", "minLength": 1 }
            },
            "required": [ "name", "start_pos", "length" ],
            "additionalProperties": false
        }
    }
}

`
)

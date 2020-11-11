// Code generated - DO NOT EDIT.

package validation

const (
    JSONSchemaFixedLengthFileDeclaration =
`
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "github.com/jf-tech/omniparser:fixedlength_file_declaration",
    "title": "omniparser schema: fixedlength/file_declaration",
    "type": "object",
    "properties": {
        "file_declaration": {
            "type": "object",
            "properties": {
                "envelopes": {
                    "oneOf": [
                        { "$ref": "#/definitions/envelopes_by_rows_type" },
                        { "$ref": "#/definitions/envelopes_by_header_footer_type" }
                    ],
                    "$comment": "by_rows and by_header_footer envelopes cannot be mixed"
                }
            },
            "required": [ "envelopes" ],
            "additionalProperties": false
        }
    },
    "required": [ "file_declaration" ],
    "definitions": {
        "envelope_columns_type": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "name": {
                        "type": "string",
                        "minLength": 1
                    },
                    "start_pos": {
                        "type": "integer",
                        "minimum": 1,
                        "$comment": "start_pos is 1-based and rune-based"
                    },
                    "length": {
                        "type": "integer",
                        "minimum": 1,
                        "$comment": "length is rune-based"
                    },
                    "line_pattern": {
                        "type": "string",
                        "minLength": 1,
                        "$comment": "regex to match a line for the column; if not specified, '.*' will be used"
                    }
                },
                "required": [ "name", "start_pos", "length" ],
                "additionalProperties": false
            },
            "minItems": 1
        },
        "envelope_by_rows_type": {
            "type": "object",
            "properties": {
                "by_rows": {
                    "type": "integer",
                    "minimum": 1,
                    "$comment": "by_rows is optional, if not specified, default to 1 by code"
                },
                "columns": { "$ref": "#/definitions/envelope_columns_type" }
            },
            "required": [ "columns" ],
            "additionalProperties": false
        },
        "envelopes_by_rows_type": {
            "type": "array",
            "items": { "$ref": "#/definitions/envelope_by_rows_type" },
            "minItems": 1,
            "maxItems": 1,
            "$comment": "there can be one and only one by_rows envelope, if it is used"
        },
        "envelope_by_header_footer_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "by_header_footer": {
                    "type": "object",
                    "properties": {
                        "header": { "type": "string", "minLength": 1 },
                        "footer": { "type": "string", "minLength": 1 }
                    },
                    "required": [ "header", "footer" ],
                    "additionalProperties": false
                },
                "columns": { "$ref": "#/definitions/envelope_columns_type" },
                "not_target": { "type": "boolean" }
            },
            "required": [ "by_header_footer" ],
            "additionalProperties": false
        },
        "envelopes_by_header_footer_type": {
            "type": "array",
            "items": { "$ref": "#/definitions/envelope_by_header_footer_type" },
            "minItems": 1
        }
    }
}

`
)

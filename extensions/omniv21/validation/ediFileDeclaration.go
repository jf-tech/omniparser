// Code generated - DO NOT EDIT.

package validation

const (
    JSONSchemaEDIFileDeclaration =
`
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "github.com/jf-tech/omniparser:edi_file_declaration",
    "title": "omniparser schema: edi/file_declaration",
    "type": "object",
    "properties": {
        "file_declaration": {
            "type": "object",
            "properties": {
                "segment_delimiter": { "type": "string", "minLength": 1 },
                "element_delimiter": { "type": "string", "minLength": 1 },
                "component_delimiter": { "type": "string", "minLength": 1 },
                "release_character": { "type": "string", "minLength": 1 },
                "ignore_crlf": { "type": "boolean" },
                "segment_declarations": {
                    "type": "array",
                    "items": {
                      "$ref": "#/definitions/segment_declaration_type"
                    }
                }
            },
            "required": [ "segment_delimiter", "element_delimiter", "segment_declarations" ],
            "additionalProperties": false
        }
    },
    "required": [ "file_declaration" ],
    "definitions": {
        "segment_declaration_type": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "minLength": 1 },
                "type": { "type": "string", "enum": [ "segment", "segment_group" ] },
                "is_target": { "type": "boolean" },
                "min": { "type": "integer", "minimum": 0 },
                "max": { "type": "integer", "minimum": -1 },
                "elements": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": { "type": "string", "minLength": 1 },
                            "index": { "type": "integer", "minimum": 1 },
                            "component_index": { "type": "integer", "minimum": 1 },
                            "empty_if_missing": { "type": "boolean","$comment": "deprecated, use 'default'" },
                            "default": { "type": "string" }
                        },
                        "required": [ "name", "index" ],
                        "additionalProperties": false
                    }
                },
                "child_segments": {
                    "type": "array",
                    "items": {
                      "$ref": "#/definitions/segment_declaration_type"
                    }
                }
            },
            "required": [ "name" ],
            "additionalProperties": false
        }
    }
}
`
)

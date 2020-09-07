package schemavalidate

//go:generate sh -c "cd jsonschemas && go run gen.go -json parser_settings.json -varname JSONSchemaParserSettings > ../parserSettings.go"
//go:generate sh -c "cd jsonschemas && go run gen.go -json transform_declarations.json -varname JSONSchemaTransformDeclarations > ../transformDeclarations.go"

import (
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidate validates a schema based on its json schema. Any validation error, if
// present, is context formatted, i.e. schema name is prefixed in the error msg.
func SchemaValidate(schemaName string, schemaContent []byte, jsonSchema string) error {
	jsonSchemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	targetSchemaLoader := gojsonschema.NewBytesLoader(schemaContent)
	result, err := gojsonschema.Validate(jsonSchemaLoader, targetSchemaLoader)
	if err != nil {
		return fmt.Errorf("unable to perform schema validation: %s", err)
	}
	if result.Valid() {
		return nil
	}
	var errs []string
	for _, err := range result.Errors() {
		errs = append(errs, err.String())
	}
	if len(errs) == 1 {
		return fmt.Errorf("schema '%s' validation failed: %s", schemaName, errs[0])
	}
	return fmt.Errorf("schema '%s' validation failed:\n%s", schemaName, strings.Join(errs, "\n"))
}

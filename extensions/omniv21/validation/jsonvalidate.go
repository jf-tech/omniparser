package validation

//go:generate sh -c "go run ../../../validation/gen/gen.go -json transform_declarations.json -varname JSONSchemaTransformDeclarations > ./transformDeclarations.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json csv_file_declaration.json -varname JSONSchemaCSVFileDeclaration > ./csvFileDeclaration.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json edi_file_declaration.json -varname JSONSchemaEDIFileDeclaration > ./ediFileDeclaration.go"

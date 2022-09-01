package validation

//go:generate sh -c "go run ../../../validation/gen/gen.go -json transformDeclarations.json -varname JSONSchemaTransformDeclarations > ./transformDeclarations.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json csvFileDeclaration.json -varname JSONSchemaCSVFileDeclaration > ./csvFileDeclaration.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json csv2FileDeclaration.json -varname JSONSchemaCSV2FileDeclaration > ./csv2FileDeclaration.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json ediFileDeclaration.json -varname JSONSchemaEDIFileDeclaration > ./ediFileDeclaration.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json fixedlengthFileDeclaration.json -varname JSONSchemaFixedLengthFileDeclaration > ./fixedlengthFileDeclaration.go"
//go:generate sh -c "go run ../../../validation/gen/gen.go -json fixedlength2FileDeclaration.json -varname JSONSchemaFixedLength2FileDeclaration > ./fixedlength2FileDeclaration.go"

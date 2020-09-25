package main

import (
	"flag"
	"io/ioutil"
	"os"
	"text/template"
)

const schemaTemplate = `// Code generated - DO NOT EDIT.

package validation

const (
    {{.SchemaVarName}} =
` + "\x60" /*\x60 is this char '`' :) */ + `
{{.Schema}}
` + "\x60" + `
)
`

type templateVars struct {
	SchemaVarName string
	Schema        string
}

func main() {
	var jsonFileName string
	flag.StringVar(&jsonFileName, "json", "", "The name of json schema file.")
	var tv templateVars
	flag.StringVar(&tv.SchemaVarName, "varname", "", "The variable name of json schema string.")
	flag.Parse()

	jsonFileContent, err := ioutil.ReadFile("./" + jsonFileName)
	if err != nil {
		os.Exit(1)
	}
	tv.Schema = string(jsonFileContent)

	err = template.Must(template.New("genjsonschema").Parse(schemaTemplate)).Execute(os.Stdout, tv)
	if err != nil {
		os.Exit(1)
	}
}

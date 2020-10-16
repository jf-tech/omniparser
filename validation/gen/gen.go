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
    {{.VarName}} =
` + "`" + `
{{.Content}}
` + "`" + `
)
`

type templateVars struct {
	VarName string
	Content string
}

func main() {
	var filename string
	flag.StringVar(&filename, "json", "", "The name of json schema file.")
	var tv templateVars
	flag.StringVar(&tv.VarName, "varname", "", "The variable name of json schema string.")
	flag.Parse()

	content, err := ioutil.ReadFile("./" + filename)
	if err != nil {
		os.Exit(1)
	}
	tv.Content = string(content)

	err = template.Must(template.New("genjsonschema").Parse(schemaTemplate)).Execute(os.Stdout, tv)
	if err != nil {
		os.Exit(1)
	}
}

package sampleutil

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/testlib"
)

func createTempFile(t *testing.T, content string) string {
	return testlib.CreateTempFileWithContent(t, "", "", content).Name()
}

func TestSampleTestCommon(t *testing.T) {
	schemaFile := createTempFile(t, `
		{
    		"parser_settings": {
        		"version": "omni.2.0",
        		"file_format_type": "xml"
    		},
    		"transform_declarations": {
        		"FINAL_OUTPUT": { "xpath": "/A/B[@flag != 'skip']", "object": {
            		"name": { "xpath": "@name" },
					"age": { "xpath": "@age", "result_type": "int" }
        		}}
			}
		}`)
	inputFile := createTempFile(t, `
		<A>
			<B flag="data" name="John" age="42"/>
			<B flag="skip" name="--" age="--"/>
			<B flag="data" name="Jane" age="53"/>
		</A>`)
	cupaloy.SnapshotT(t, jsons.BPJ(SampleTestCommon(t, schemaFile, inputFile)))
}
